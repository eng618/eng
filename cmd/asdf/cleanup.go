package asdf

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/asdf"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

var (
	interactiveFlag bool
	dryRunFlag      bool
	forceFlag       bool
	pluginFlag      string
	configFlag      string
	noScanFlag      bool
	scanDirFlag     string
)

// PruneCmd represents the subcommand to prune outdated asdf installs.
var PruneCmd = &cobra.Command{
	Use:     "prune",
	Aliases: []string{"cleanup", "clean"},
	Short:   "Prune outdated asdf tool versions not listed in .tool-versions",
	Long: `Reads the root (user level) .tool-versions file and scans project directories for additional .tool-versions files, then removes installed tool versions that are not explicitly defined.

Use the --interactive (-i) flag to present a multi-select prompt for choosing specific versions to uninstall.
Use the --dry-run (-d) flag to preview which versions would be uninstalled.
Use the --no-scan flag to disable searching development directories for project .tool-versions files.`,
	RunE: runPrune,
}

// CleanupCmd is an alias variable for backward compatibility.
var CleanupCmd = PruneCmd

func bindPruneFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&interactiveFlag, "interactive", "i", false, "Interactively select installed tool versions to delete")
	cmd.Flags().BoolVarP(&dryRunFlag, "dry-run", "d", false, "Show what versions would be deleted without removing them")
	cmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Skip confirmation prompt in default mode")
	cmd.Flags().StringVarP(&pluginFlag, "plugin", "p", "", "Limit cleanup to a specific plugin (e.g. nodejs)")
	cmd.Flags().StringVarP(&configFlag, "config", "c", "", "Path to specific .tool-versions file (default scans $HOME and dev folders)")
	cmd.Flags().BoolVar(&noScanFlag, "no-scan", false, "Disable scanning development directories for project .tool-versions files")
	cmd.Flags().StringVar(&scanDirFlag, "scan-dir", "", "Additional directory to recursively scan for .tool-versions files")
}

func runPrune(cmd *cobra.Command, _args []string) error {
	// Header styling
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary).
		MarginBottom(1)
	if !ui.DisableProgress {
		fmt.Println(headerStyle.Render("🧹 ASDF Tool Prune"))
	}

	// 1. Verify asdf is installed
	if _, err := lookPath("asdf"); err != nil {
		log.Error("asdf executable not found in PATH")
		return fmt.Errorf("asdf executable not found in PATH: %w", err)
	}

	// 2. Resolve user home & search directories
	homeDir, homeErr := userHomeDir()
	var searchRoots []string

	if configFlag != "" {
		searchRoots = append(searchRoots, configFlag)
	} else {
		if homeErr == nil {
			homeTV := filepath.Join(homeDir, ".tool-versions")
			searchRoots = append(searchRoots, homeTV)
		}

		if !noScanFlag {
			// Include configured dev path if set
			if devPath := viper.GetString("git.dev_path"); devPath != "" {
				searchRoots = append(searchRoots, devPath)
			} else if homeErr == nil {
				// Default fallback dev path
				defaultDev := filepath.Join(homeDir, "Development")
				if _, err := os.Stat(defaultDev); err == nil {
					searchRoots = append(searchRoots, defaultDev)
				}
			}

			// Include current working directory
			if cwd, err := os.Getwd(); err == nil {
				searchRoots = append(searchRoots, cwd)
			}

			if scanDirFlag != "" {
				searchRoots = append(searchRoots, scanDirFlag)
			}
		}
	}

	// 3. Scan & parse .tool-versions and installed tools
	var scanSpinner *ui.Spinner
	if !ui.DisableProgress {
		scanSpinner = ui.NewSpinner("Scanning for .tool-versions files and calculating disk usage...")
	}

	discoveredFiles, err := asdf.FindToolVersionFiles(searchRoots)
	if err != nil {
		if scanSpinner != nil {
			scanSpinner.Stop()
		}
		log.Error("Error scanning for .tool-versions files: %v", err)
		return err
	}

	protected, summaries, err := asdf.ParseAndMergeToolVersions(discoveredFiles)
	if err != nil {
		if scanSpinner != nil {
			scanSpinner.Stop()
		}
		log.Error("Failed to parse .tool-versions files: %v", err)
		return err
	}

	listCmd := execCommand("asdf", "list")
	out, err := listCmd.Output()
	if err != nil {
		if scanSpinner != nil {
			scanSpinner.Stop()
		}
		log.Error("Failed to run 'asdf list': %v", err)
		return fmt.Errorf("failed to run 'asdf list': %w", err)
	}

	installed, err := asdf.ParseASDFListOutput(string(out))
	if err != nil {
		if scanSpinner != nil {
			scanSpinner.Stop()
		}
		log.Error("Failed to parse 'asdf list' output: %v", err)
		return err
	}

	asdfDataDir := asdf.GetASDFDataDir(homeDir)
	removable := asdf.FilterRemovableVersions(installed, protected, pluginFlag, asdfDataDir)

	if scanSpinner != nil {
		scanSpinner.Stop()
	}

	// Print summary of discovered .tool-versions protection files
	if !ui.DisableProgress && len(summaries) > 0 {
		fmt.Printf("Discovered %d .tool-versions file(s) protecting active versions:\n", len(summaries))
		for _, s := range summaries {
			fmt.Printf("  %s %s\n", theme.SuccessText.Render("✓"), s.FormatFileSummary(homeDir))
		}
		fmt.Println()
	}

	if len(removable) == 0 {
		theme.SuccessMessage("No outdated or extra asdf versions found. Your system is clean!")
		return nil
	}

	// Map label -> target
	labelMap := make(map[string]asdf.CleanupTarget)
	var options []string
	for _, target := range removable {
		label := target.FormatTargetLabel()
		labelMap[label] = target
		options = append(options, label)
	}

	var targets []asdf.CleanupTarget

	// 4. Interactive mode selection vs default mode
	if interactiveFlag {
		selected, err := ui.MultiSelect("Select asdf versions to uninstall:", options, nil)
		if err != nil {
			log.Message("Interactive selection cancelled.")
			return nil
		}

		if len(selected) == 0 {
			log.Message("No versions selected for removal.")
			return nil
		}

		for _, item := range selected {
			if target, ok := labelMap[item]; ok {
				targets = append(targets, target)
			} else {
				parts := strings.Split(item, " @ ")
				if len(parts) >= 2 {
					versionPart := strings.Fields(parts[1])[0]
					targets = append(targets, asdf.CleanupTarget{
						Plugin:  parts[0],
						Version: versionPart,
					})
				}
			}
		}
	} else {
		targets = removable

		// Group targets by plugin for clean display
		pluginCounts := make(map[string]int)
		pluginSizes := make(map[string]int64)
		var totalReclaimableBytes int64

		for _, t := range targets {
			pluginCounts[t.Plugin]++
			pluginSizes[t.Plugin] += t.SizeBytes
			totalReclaimableBytes += t.SizeBytes
		}

		var pluginNames []string
		for p := range pluginCounts {
			pluginNames = append(pluginNames, p)
		}
		sort.Strings(pluginNames)

		var breakdown []string
		sizeFormatted := humanize.Bytes(uint64(totalReclaimableBytes))

		breakdown = append(breakdown, fmt.Sprintf("Found %s removable version(s) reclaiming %s space across %s plugin(s):",
			theme.PrimaryText.Bold(true).Render(fmt.Sprintf("%d", len(targets))),
			theme.SuccessText.Bold(true).Render(sizeFormatted),
			theme.PrimaryText.Bold(true).Render(fmt.Sprintf("%d", len(pluginNames))),
		))

		for _, p := range pluginNames {
			sizeStr := ""
			if pluginSizes[p] > 0 {
				sizeStr = fmt.Sprintf(" (%s)", humanize.Bytes(uint64(pluginSizes[p])))
			}
			breakdown = append(breakdown, fmt.Sprintf("  • %s: %s%s",
				theme.BoldText.Render(p),
				theme.MutedText.Render(fmt.Sprintf("%d version(s)", pluginCounts[p])),
				theme.MutedText.Render(sizeStr),
			))
		}

		if !ui.DisableProgress {
			boxContent := strings.Join(breakdown, "\n")
			fmt.Println(theme.InfoBox.Render(boxContent))
		}

		if !dryRunFlag && !forceFlag {
			confirmMsg := fmt.Sprintf("Are you sure you want to uninstall these %d version(s) to reclaim %s?", len(targets), sizeFormatted)
			confirmed, err := ui.Confirm(confirmMsg, false)
			if err != nil || !confirmed {
				log.Message("Prune operation cancelled.")
				return nil
			}
		}
	}

	totalTargets := len(targets)
	var progressSpinner *ui.Spinner
	if !ui.DisableProgress {
		progressSpinner = ui.NewProgressSpinner(fmt.Sprintf("Pruning %d asdf version(s)...", totalTargets))
	}

	// 5. Execute uninstalls with live progress indicator and disk space tracking
	var uninstalledCount int
	var freedBytes int64
	cleanedPlugins := make(map[string]bool)

	for i, target := range targets {
		ratio := float64(i+1) / float64(totalTargets)
		statusMsg := fmt.Sprintf("[%d/%d] Uninstalling %s @ %s", i+1, totalTargets, target.Plugin, target.Version)

		if progressSpinner != nil {
			progressSpinner.SetProgressBar(ratio, statusMsg)
		}

		sizeTag := ""
		if target.SizeBytes > 0 {
			sizeTag = fmt.Sprintf(" (freed %s)", humanize.Bytes(uint64(target.SizeBytes)))
		}

		if dryRunFlag {
			if progressSpinner != nil {
				progressSpinner.Logf("  %s [Dry Run] %s @ %s%s\n", theme.MutedText.Render("•"), target.Plugin, target.Version, sizeTag)
			} else {
				log.Message("[Dry Run] Would uninstall: %s @ %s%s", target.Plugin, target.Version, sizeTag)
			}
			uninstalledCount++
			freedBytes += target.SizeBytes
			cleanedPlugins[target.Plugin] = true
			continue
		}

		uninstCmd := execCommand("asdf", "uninstall", target.Plugin, target.Version)
		if err := uninstCmd.Run(); err != nil {
			if progressSpinner != nil {
				progressSpinner.Logf("  %s Failed %s @ %s: %v\n", theme.ErrorText.Render("✗"), target.Plugin, target.Version, err)
			} else {
				log.Error("Failed to uninstall %s @ %s: %v", target.Plugin, target.Version, err)
			}
		} else {
			if progressSpinner != nil {
				progressSpinner.Logf("  %s Uninstalled %s @ %s%s\n", theme.SuccessText.Render("✓"), target.Plugin, target.Version, sizeTag)
			} else {
				log.Success("Uninstalled %s @ %s%s", target.Plugin, target.Version, sizeTag)
			}
			uninstalledCount++
			freedBytes += target.SizeBytes
			cleanedPlugins[target.Plugin] = true
		}
	}

	if progressSpinner != nil {
		progressSpinner.SetProgressBar(1.0, "Prune completed")
		progressSpinner.Stop()
	}

	// 6. Render final summary banner with reclaimed disk space
	freedStr := humanize.Bytes(uint64(freedBytes))
	if dryRunFlag {
		theme.InfoMessage(fmt.Sprintf("Dry run complete. Previewed %d uninstallation(s) reclaiming %s disk space.", uninstalledCount, freedStr))
	} else {
		summaryMsg := fmt.Sprintf("ASDF prune completed! Successfully freed %s of disk space (%d version(s) across %d plugin(s)).", freedStr, uninstalledCount, len(cleanedPlugins))
		theme.SuccessMessage(summaryMsg)
	}

	return nil
}
