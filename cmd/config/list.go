package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

// ListCmd represents the config list/show command.
var ListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"show", "ls"},
	Short:   "Display a summary of all active configuration settings",
	Long:    `Renders a formatted overview of all current configuration keys, values, and file paths.`,
	Run: func(cmd *cobra.Command, _args []string) {
		PrintConfigSummary()
	},
}

// PrintConfigSummary renders a stylized card of all current settings.
func PrintConfigSummary() {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary).
		MarginBottom(1)

	if !ui.DisableProgress {
		fmt.Fprintln(log.Out, headerStyle.Render("⚙️ Current eng Configuration"))
	} else {
		log.Info("Current eng Configuration")
	}

	var cardLines []string

	// Config File
	cfgFile := viper.ConfigFileUsed()
	if cfgFile == "" {
		cfgFile = "None loaded (using defaults)"
	}
	cardLines = append(
		cardLines,
		fmt.Sprintf("  %-24s %s", theme.BoldText.Render("Config File:"), theme.MutedText.Render(cfgFile)),
	)

	// User Email
	userEmail := viper.GetString("user-email")
	if userEmail == "" {
		userEmail = "Not set"
	}
	cardLines = append(
		cardLines,
		fmt.Sprintf("  %-24s %s", theme.BoldText.Render("User Email:"), theme.PrimaryText.Render(userEmail)),
	)

	// Git Dev Path
	gitCfg := config.GetGitConfig()
	devPath := gitCfg.DevPath
	if devPath != "" {
		devPath = os.ExpandEnv(devPath)
	} else {
		devPath = "Not set"
	}
	cardLines = append(
		cardLines,
		fmt.Sprintf("  %-24s %s", theme.BoldText.Render("Git Dev Path:"), theme.BaseText.Render(devPath)),
	)

	// Dotfiles
	dotfilesCfg := config.GetDotfilesConfig()
	dotfilesRepoURL := dotfilesCfg.RepoURL
	if dotfilesRepoURL == "" {
		dotfilesRepoURL = "Not set"
	}
	cardLines = append(
		cardLines,
		fmt.Sprintf("  %-24s %s", theme.BoldText.Render("Dotfiles Repo URL:"), theme.BaseText.Render(dotfilesRepoURL)),
	)

	dotfilesBranch := dotfilesCfg.Branch
	if dotfilesBranch == "" {
		dotfilesBranch = "main"
	}
	cardLines = append(
		cardLines,
		fmt.Sprintf("  %-24s %s", theme.BoldText.Render("Dotfiles Branch:"), theme.BaseText.Render(dotfilesBranch)),
	)

	barePath := dotfilesCfg.BareRepoPath
	if barePath != "" {
		barePath = os.ExpandEnv(barePath)
	} else {
		barePath = "Not set"
	}
	cardLines = append(
		cardLines,
		fmt.Sprintf("  %-24s %s", theme.BoldText.Render("Dotfiles Bare Path:"), theme.MutedText.Render(barePath)),
	)

	// IDE URL
	ideURL := viper.GetString("antigravity.ide_download_url")
	if ideURL == "" {
		ideURL = "Default"
	}
	cardLines = append(
		cardLines,
		fmt.Sprintf("  %-24s %s", theme.BoldText.Render("IDE Download URL:"), theme.MutedText.Render(ideURL)),
	)

	// Verbose Mode
	verbose := viper.GetBool("verbose")
	var verboseStr string
	if verbose {
		verboseStr = theme.SuccessText.Render("true")
	} else {
		verboseStr = theme.MutedText.Render("false")
	}
	cardLines = append(cardLines, fmt.Sprintf("  %-24s %s", theme.BoldText.Render("Verbose Mode:"), verboseStr))

	// Telemetry Status
	telemetryEnabled := config.IsTelemetryEnabled()
	var telemetryStr string
	if telemetryEnabled && config.IsTelemetryConfigured() {
		telemetryStr = theme.SuccessText.Render("true (OpenPanel)")
	} else if !config.IsTelemetryConfigured() {
		telemetryStr = theme.MutedText.Render("false (unconfigured dev build)")
	} else {
		telemetryStr = theme.MutedText.Render("false (disabled)")
	}
	cardLines = append(cardLines, fmt.Sprintf("  %-24s %s", theme.BoldText.Render("Telemetry:"), telemetryStr))

	if !ui.DisableProgress {
		fmt.Fprintln(log.Out, theme.InfoBox.Render(strings.Join(cardLines, "\n")))
	} else {
		for _, l := range cardLines {
			log.Message("%s", l)
		}
		log.Message("")
	}
}
