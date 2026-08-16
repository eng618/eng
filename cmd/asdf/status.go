package asdf

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/asdf"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

// StatusCmd represents the subcommand to display an environment dashboard of asdf tools.
var StatusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"ls", "list-summary"},
	Short:   "Display an overview dashboard of installed asdf plugins, active versions, and disk usage",
	Long:    `Provides a high-level summary dashboard of all managed asdf plugins, installed version counts, active pinned versions, and total disk space used.`,
	RunE:    runStatus,
}

func runStatus(_cmd *cobra.Command, _args []string) error {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary).
		MarginBottom(1)
	if !ui.DisableProgress {
		fmt.Fprintln(log.Out, headerStyle.Render("📊 ASDF Environment Dashboard"))
	}

	if _, err := lookPath("asdf"); err != nil {
		log.Error("asdf executable not found in PATH")
		return fmt.Errorf("asdf executable not found in PATH: %w", err)
	}

	homeDir, homeErr := userHomeDir()
	asdfDataDir := asdf.GetASDFDataDir(homeDir)

	var scanSpinner *ui.Spinner
	if !ui.DisableProgress {
		scanSpinner = ui.NewSpinner("Inspecting asdf plugins and active versions...")
		scanSpinner.Start()
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

	// Parse root .tool-versions for active pins
	var activePins asdf.ToolVersions
	if homeErr == nil {
		rootTVPath := filepath.Join(homeDir, ".tool-versions")
		activePins, _ = asdf.ParseToolVersionsFile(rootTVPath)
	}

	if scanSpinner != nil {
		scanSpinner.Stop()
	}

	if len(installed) == 0 {
		theme.InfoMessage("No asdf plugins or versions currently installed.")
		return nil
	}

	var pluginNames []string
	for p := range installed {
		pluginNames = append(pluginNames, p)
	}
	sort.Strings(pluginNames)

	var totalVersions int
	var totalDiskBytes int64

	type PluginSummary struct {
		Name         string
		ActivePin    string
		VersionCount int
		DiskBytes    int64
	}

	var summaries []PluginSummary
	totalPlugins := len(pluginNames)

	var progressSpinner *ui.Spinner
	if !ui.DisableProgress {
		progressSpinner = ui.NewProgressSpinner(fmt.Sprintf("Calculating disk usage for %d plugin(s)...", totalPlugins))
		progressSpinner.Start()
	}

	for idx, plugin := range pluginNames {
		if progressSpinner != nil {
			ratio := float64(idx+1) / float64(totalPlugins)
			progressSpinner.SetProgressBar(
				ratio,
				fmt.Sprintf("[%d/%d] Inspecting disk usage for %s...", idx+1, totalPlugins, plugin),
			)
		}

		versions := installed[plugin]
		vCount := len(versions)
		totalVersions += vCount

		var pluginBytes int64
		for _, v := range versions {
			installDir := filepath.Join(asdfDataDir, "installs", plugin, v)
			pluginBytes += asdf.CalculateDirSize(installDir)
		}
		totalDiskBytes += pluginBytes

		pinStr := "none"
		if pins, ok := activePins[plugin]; ok && len(pins) > 0 {
			pinStr = strings.Join(pins, ", ")
		}

		summaries = append(summaries, PluginSummary{
			Name:         plugin,
			ActivePin:    pinStr,
			VersionCount: vCount,
			DiskBytes:    pluginBytes,
		})
	}

	if progressSpinner != nil {
		progressSpinner.SetProgressBar(1.0, "Disk usage calculation complete")
		progressSpinner.Stop()
	}

	// Render Dashboard Callout
	var dashboardLines []string
	dashboardLines = append(
		dashboardLines,
		fmt.Sprintf("Summary: %s plugins  •  %s installed versions  •  %s total disk space",
			theme.PrimaryText.Bold(true).Render(fmt.Sprintf("%d", len(pluginNames))),
			theme.PrimaryText.Bold(true).Render(fmt.Sprintf("%d", totalVersions)),
			theme.SuccessText.Bold(true).Render(humanize.Bytes(uint64(totalDiskBytes))),
		),
	)
	dashboardLines = append(dashboardLines, "")
	dashboardLines = append(dashboardLines, fmt.Sprintf("  %-25s %-20s %-12s %s",
		theme.BoldText.Render("Plugin"),
		theme.BoldText.Render("Active Pin"),
		theme.BoldText.Render("Versions"),
		theme.BoldText.Render("Disk Space"),
	))
	dashboardLines = append(dashboardLines, "  "+strings.Repeat("─", 68))

	for _, s := range summaries {
		sizeStr := "0 B"
		if s.DiskBytes > 0 {
			sizeStr = humanize.Bytes(uint64(s.DiskBytes))
		}

		dashboardLines = append(dashboardLines, fmt.Sprintf("  %-25s %-20s %-12s %s",
			theme.PrimaryText.Render(s.Name),
			theme.MutedText.Render(s.ActivePin),
			fmt.Sprintf("%d version(s)", s.VersionCount),
			theme.SuccessText.Render(sizeStr),
		))
	}

	if !ui.DisableProgress {
		boxStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Primary).
			Padding(0, 1).
			MarginBottom(1)
		fmt.Fprintln(log.Out, boxStyle.Render(strings.Join(dashboardLines, "\n")))
	}

	return nil
}
