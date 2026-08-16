package asdf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/asdf"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

var (
	installMissingFlag bool
	checkNoScanFlag    bool
)

// CheckCmd represents the subcommand to verify all .tool-versions requirements are installed.
var CheckCmd = &cobra.Command{
	Use:     "check",
	Aliases: []string{"verify", "doctor"},
	Short:   "Verify that all required tool versions in .tool-versions files are installed",
	Long: `Scans global and project .tool-versions files and verifies that every required tool version is installed on the system.

Use the --install (-i) flag to automatically install any missing tool versions.`,
	RunE: runCheck,
}

func init() {
	CheckCmd.Flags().BoolVarP(&installMissingFlag, "install", "i", false, "Automatically install missing tool versions")
	CheckCmd.Flags().BoolVar(&checkNoScanFlag, "no-scan", false, "Disable scanning development directories")
}

func runCheck(_cmd *cobra.Command, _args []string) error {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary).
		MarginBottom(1)
	if !ui.DisableProgress {
		fmt.Fprintln(log.Out, headerStyle.Render("🔍 ASDF Tool Version Check"))
	}

	if _, err := lookPath("asdf"); err != nil {
		log.Error("asdf executable not found in PATH")
		return fmt.Errorf("asdf executable not found in PATH: %w", err)
	}

	homeDir, homeErr := userHomeDir()
	var searchRoots []string

	if homeErr == nil {
		homeTV := filepath.Join(homeDir, ".tool-versions")
		searchRoots = append(searchRoots, homeTV)
	}

	if !checkNoScanFlag {
		if devPath := viper.GetString("git.dev_path"); devPath != "" {
			searchRoots = append(searchRoots, devPath)
		} else if homeErr == nil {
			defaultDev := filepath.Join(homeDir, "Development")
			if _, err := os.Stat(defaultDev); err == nil {
				searchRoots = append(searchRoots, defaultDev)
			}
		}

		if cwd, err := os.Getwd(); err == nil {
			searchRoots = append(searchRoots, cwd)
		}
	}

	var scanSpinner *ui.Spinner
	if !ui.DisableProgress {
		scanSpinner = ui.NewSpinner("Checking .tool-versions requirements and installed versions...")
		scanSpinner.Start()
	}

	discoveredFiles, err := asdf.FindToolVersionFilesWithProgress(searchRoots, func(_ string, count int) {
		if scanSpinner != nil {
			scanSpinner.UpdateMessage(fmt.Sprintf("Scanning for .tool-versions... (%d found)", count))
		}
	})
	if err != nil {
		if scanSpinner != nil {
			scanSpinner.Stop()
		}
		log.Error("Error scanning for .tool-versions files: %v", err)
		return err
	}

	if scanSpinner != nil {
		scanSpinner.UpdateMessage("Parsing .tool-versions requirements...")
	}

	_, summaries, err := asdf.ParseAndMergeToolVersions(discoveredFiles)
	if err != nil {
		if scanSpinner != nil {
			scanSpinner.Stop()
		}
		log.Error("Failed to parse .tool-versions files: %v", err)
		return err
	}

	if scanSpinner != nil {
		scanSpinner.UpdateMessage("Verifying installed versions via 'asdf list'...")
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

	missing := asdf.CheckMissingRequirements(summaries, installed)

	if scanSpinner != nil {
		scanSpinner.Stop()
	}

	if len(missing) == 0 {
		theme.SuccessMessage(
			fmt.Sprintf(
				"All project .tool-versions requirements are installed and ready! (Checked %d file(s))",
				len(summaries),
			),
		)
		return nil
	}

	var breakdown []string
	breakdown = append(breakdown, fmt.Sprintf("Missing %s required tool version(s):",
		theme.ErrorText.Render(fmt.Sprintf("%d", len(missing))),
	))

	for _, m := range missing {
		sourceStr := m.FormatSource(homeDir)
		breakdown = append(breakdown, fmt.Sprintf("  • %s @ %s %s",
			theme.BoldText.Render(m.Plugin),
			theme.ErrorText.Render(m.Version),
			theme.MutedText.Render(fmt.Sprintf("(required by %s)", sourceStr)),
		))
	}

	if !ui.DisableProgress {
		boxStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Destructive).
			Padding(0, 1).
			MarginBottom(1)
		fmt.Fprintln(log.Out, boxStyle.Render(strings.Join(breakdown, "\n")))
	}

	shouldInstall := installMissingFlag
	if !shouldInstall {
		confirmMsg := fmt.Sprintf("Would you like to install these %d missing tool version(s) now?", len(missing))
		confirmed, err := ui.Confirm(confirmMsg, true)
		if err == nil && confirmed {
			shouldInstall = true
		}
	}

	if !shouldInstall {
		log.Message("Skipped installation of missing versions.")
		return nil
	}

	// Run asdf install for missing requirements
	totalMissing := len(missing)
	var progressSpinner *ui.Spinner
	if !ui.DisableProgress {
		progressSpinner = ui.NewProgressSpinner(fmt.Sprintf("Installing %d missing tool version(s)...", totalMissing))
		progressSpinner.Start()
	}

	var installedCount int
	for i, m := range missing {
		ratio := float64(i+1) / float64(totalMissing)
		statusMsg := fmt.Sprintf("[%d/%d] Installing %s @ %s", i+1, totalMissing, m.Plugin, m.Version)

		if progressSpinner != nil {
			progressSpinner.SetProgressBar(ratio, statusMsg)
		}

		instCmd := execCommand("asdf", "install", m.Plugin, m.Version)
		out, err := instCmd.CombinedOutput()

		if err != nil {
			errDetail := strings.TrimSpace(string(out))
			if errDetail != "" {
				errDetail = ": " + errDetail
			}
			if progressSpinner != nil {
				progressSpinner.Logf("  %s Failed %s @ %s%s\n", theme.ErrorText.Render("✗"), m.Plugin, m.Version, errDetail)
			} else {
				log.Error("Failed to install %s @ %s%s", m.Plugin, m.Version, errDetail)
			}
		} else {
			if progressSpinner != nil {
				progressSpinner.Logf("  %s Installed %s @ %s\n", theme.SuccessText.Render("✓"), m.Plugin, m.Version)
			} else {
				log.Success("Installed %s @ %s", m.Plugin, m.Version)
			}
			installedCount++
		}
	}

	if progressSpinner != nil {
		progressSpinner.SetProgressBar(1.0, "Installation completed")
		progressSpinner.Stop()
	}

	theme.SuccessMessage(fmt.Sprintf("Installed %d missing tool version(s) successfully!", installedCount))
	return nil
}
