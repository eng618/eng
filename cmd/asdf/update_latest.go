package asdf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/asdf"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

var (
	updateYesFlag     bool
	updateInstallFlag bool
	updateConfigFlag  string
)

// UpdateLatestCmd represents the subcommand to update .tool-versions to latest available versions.
var UpdateLatestCmd = &cobra.Command{
	Use:     "update-latest",
	Aliases: []string{"update-root", "update-tools", "upgrade"},
	Short:   "Update .tool-versions file to the latest available releases for each tool (defaults to $HOME/.tool-versions)",
	Long: `Reads a .tool-versions file (defaults to the user level global config at $HOME/.tool-versions), queries 'asdf latest' for each plugin, and updates the file with the newest available releases.

Use the --config (-c) flag to specify a custom .tool-versions file path.
Use the --install (-i) flag to automatically install the upgraded tool versions.
Use the --yes (-y) flag to skip prompts and apply all upgrades.`,
	RunE: runUpdateLatest,
}

func init() {
	UpdateLatestCmd.Flags().BoolVarP(&updateYesFlag, "yes", "y", false, "Skip prompts and apply all tool upgrades")
	UpdateLatestCmd.Flags().
		BoolVarP(&updateInstallFlag, "install", "i", false, "Automatically run asdf install after updating .tool-versions")
	UpdateLatestCmd.Flags().
		StringVarP(&updateConfigFlag, "config", "c", "", "Path to specific .tool-versions file (defaults to user level global config at $HOME/.tool-versions)")
}

func runUpdateLatest(_cmd *cobra.Command, _args []string) error {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary).
		MarginBottom(1)
	if !ui.DisableProgress {
		fmt.Fprintln(log.Out, headerStyle.Render("🚀 ASDF Tool Upgrade"))
	}

	if _, err := lookPath("asdf"); err != nil {
		log.Error("asdf executable not found in PATH")
		return fmt.Errorf("asdf executable not found in PATH: %w", err)
	}

	toolVersionsPath := updateConfigFlag
	if toolVersionsPath == "" {
		homeDir, err := userHomeDir()
		if err != nil {
			log.Error("Could not determine user home directory: %v", err)
			return fmt.Errorf("could not determine user home directory: %w", err)
		}
		toolVersionsPath = filepath.Join(homeDir, ".tool-versions")
	}

	if _, err := os.Stat(toolVersionsPath); os.IsNotExist(err) {
		log.Error(".tool-versions file not found at %s", toolVersionsPath)
		return fmt.Errorf(".tool-versions file not found at %s", toolVersionsPath)
	}

	currentTV, err := asdf.ParseToolVersionsFile(toolVersionsPath)
	if err != nil {
		log.Error("Failed to parse %s: %v", toolVersionsPath, err)
		return err
	}

	if len(currentTV) == 0 {
		log.Message("No tools defined in %s", toolVersionsPath)
		return nil
	}

	type ToolUpgrade struct {
		Plugin         string
		CurrentVersion string
		LatestVersion  string
		CanUpgrade     bool
	}

	var upgrades []ToolUpgrade
	var upgradableCount int
	totalPlugins := len(currentTV)

	var fetchSpinner *ui.Spinner
	if !ui.DisableProgress {
		fetchSpinner = ui.NewProgressSpinner(fmt.Sprintf("Querying latest available releases for %d tool(s)...", totalPlugins))
		fetchSpinner.Start()
	}

	idx := 0
	for plugin, versions := range currentTV {
		idx++
		ratio := float64(idx) / float64(totalPlugins)
		if fetchSpinner != nil {
			fetchSpinner.SetProgressBar(ratio, fmt.Sprintf("[%d/%d] Checking latest version for %s...", idx, totalPlugins, plugin))
		}

		currentVer := ""
		if len(versions) > 0 {
			currentVer = versions[0]
		}

		latestCmd := execCommand("asdf", "latest", plugin)
		out, err := latestCmd.Output()

		latestVer := strings.TrimSpace(string(out))
		canUp := err == nil && latestVer != "" && latestVer != currentVer

		if canUp {
			upgradableCount++
		}

		upgrades = append(upgrades, ToolUpgrade{
			Plugin:         plugin,
			CurrentVersion: currentVer,
			LatestVersion:  latestVer,
			CanUpgrade:     canUp,
		})
	}

	if fetchSpinner != nil {
		fetchSpinner.SetProgressBar(1.0, "Release check complete")
		fetchSpinner.Stop()
	}

	if upgradableCount == 0 {
		theme.SuccessMessage("All tools in .tool-versions are up to date with their latest releases!")
		return nil
	}

	// Render diff table
	var tableLines []string
	tableLines = append(tableLines, fmt.Sprintf(
		"Found %s tool(s) with available upgrades:",
		theme.PrimaryText.Bold(true).Render(fmt.Sprintf("%d", upgradableCount)),
	))

	for _, u := range upgrades {
		if u.CanUpgrade {
			tableLines = append(tableLines, fmt.Sprintf(
				"  • %-20s %s ➔ %s",
				theme.BoldText.Render(u.Plugin),
				theme.MutedText.Render(u.CurrentVersion),
				theme.SuccessText.Bold(true).Render(u.LatestVersion),
			))
		} else {
			tableLines = append(tableLines, fmt.Sprintf(
				"  • %-20s %s %s",
				theme.MutedText.Render(u.Plugin),
				theme.MutedText.Render(u.CurrentVersion),
				theme.MutedText.Render("(up to date)"),
			))
		}
	}

	if !ui.DisableProgress {
		boxContent := strings.Join(tableLines, "\n")
		fmt.Println(theme.InfoBox.Render(boxContent))
	}

	var selectedToUpgrade []ToolUpgrade

	if updateYesFlag {
		for _, u := range upgrades {
			if u.CanUpgrade {
				selectedToUpgrade = append(selectedToUpgrade, u)
			}
		}
	} else {
		var options []string
		var defaultSelected []string

		for _, u := range upgrades {
			if u.CanUpgrade {
				optLabel := fmt.Sprintf("%s @ %s ➔ %s", u.Plugin, u.CurrentVersion, u.LatestVersion)
				options = append(options, optLabel)
				defaultSelected = append(defaultSelected, optLabel)
			}
		}

		selected, err := ui.MultiSelect("Select tools to upgrade in .tool-versions:", options, defaultSelected)
		if err != nil || len(selected) == 0 {
			log.Message("Upgrade cancelled.")
			return nil
		}

		for _, item := range selected {
			parts := strings.Split(item, " @ ")
			pluginName := parts[0]
			for _, u := range upgrades {
				if u.Plugin == pluginName && u.CanUpgrade {
					selectedToUpgrade = append(selectedToUpgrade, u)
					break
				}
			}
		}
	}

	if len(selectedToUpgrade) == 0 {
		log.Message("No tool upgrades selected.")
		return nil
	}

	// Apply updates to ToolVersions map
	updatedTV := make(asdf.ToolVersions)
	for plugin, versions := range currentTV {
		updatedTV[plugin] = versions
	}

	for _, u := range selectedToUpgrade {
		updatedTV[u.Plugin] = []string{u.LatestVersion}
	}

	if err := asdf.WriteToolVersions(toolVersionsPath, updatedTV); err != nil {
		log.Error("Failed to update %s: %v", toolVersionsPath, err)
		return err
	}

	theme.SuccessMessage(
		fmt.Sprintf("Updated %s with %d upgraded tool version(s)!", toolVersionsPath, len(selectedToUpgrade)),
	)

	// Install newly upgraded versions if requested or confirmed
	shouldInstall := updateInstallFlag
	if !shouldInstall && !updateYesFlag {
		confirmMsg := fmt.Sprintf(
			"Would you like to run 'asdf install' for the %d upgraded tool(s) now?",
			len(selectedToUpgrade),
		)
		confirmed, err := ui.Confirm(confirmMsg, true)
		if err == nil && confirmed {
			shouldInstall = true
		}
	}

	if !shouldInstall {
		return nil
	}

	totalUpgrades := len(selectedToUpgrade)
	var progressSpinner *ui.Spinner
	if !ui.DisableProgress {
		progressSpinner = ui.NewProgressSpinner(fmt.Sprintf("Installing %d upgraded tool(s)...", totalUpgrades))
		progressSpinner.Start()
	}

	for i, u := range selectedToUpgrade {
		ratio := float64(i+1) / float64(totalUpgrades)
		statusMsg := fmt.Sprintf("[%d/%d] Installing %s @ %s", i+1, totalUpgrades, u.Plugin, u.LatestVersion)

		if progressSpinner != nil {
			progressSpinner.SetProgressBar(ratio, statusMsg)
		}

		instCmd := execCommand("asdf", "install", u.Plugin, u.LatestVersion)
		out, err := instCmd.CombinedOutput()

		if err != nil {
			errDetail := strings.TrimSpace(string(out))
			if errDetail != "" {
				errDetail = ": " + errDetail
			}
			if progressSpinner != nil {
				progressSpinner.Logf(
					"  %s Failed %s @ %s%s\n",
					theme.ErrorText.Render("✗"),
					u.Plugin,
					u.LatestVersion,
					errDetail,
				)
			} else {
				log.Error("Failed to install %s @ %s%s", u.Plugin, u.LatestVersion, errDetail)
			}
		} else {
			if progressSpinner != nil {
				progressSpinner.Logf(
					"  %s Installed %s @ %s\n",
					theme.SuccessText.Render("✓"),
					u.Plugin,
					u.LatestVersion,
				)
			} else {
				log.Success("Installed %s @ %s", u.Plugin, u.LatestVersion)
			}
		}
	}

	if progressSpinner != nil {
		progressSpinner.SetProgressBar(1.0, "Installation completed")
		progressSpinner.Stop()
	}

	theme.SuccessMessage("All upgraded tool versions installed successfully!")
	return nil
}
