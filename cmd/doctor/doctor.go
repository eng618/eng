// Package doctor provides the 'eng doctor' diagnostic command to check
// workstation dependencies, tools, configuration paths, and runtime health.
package doctor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/cmd/version"
	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/telemetry"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

var (
	execLookPath = exec.LookPath
	osStat       = os.Stat
)

// ToolCheck defines an external dependency check.
type ToolCheck struct {
	Name        string
	Binary      string
	Description string
	Required    bool
}

// DoctorCmd represents the doctor diagnostic command.
var DoctorCmd = &cobra.Command{
	Use:     "doctor",
	Aliases: []string{"doc"},
	Short:   "Check system health and verify dependencies",
	Long: `Inspects your workstation environment and verifies the availability and health of:
  - Core CLI tools (git, brew, docker, asdf, tailscale, bw, gpg, gh, glab)
  - Configured workspace paths and dotfiles
  - Installation and update status

Exits non-zero when a required tool is missing, so CI can gate on it.`,
	Example: `  eng doctor`,
	RunE: func(cmd *cobra.Command, _args []string) error {
		return runDoctor()
	},
}

func runDoctor() error {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary).
		MarginBottom(1)

	if !ui.DisableProgress {
		fmt.Fprintln(log.Out, headerStyle.Render("🩺 eng Workstation Diagnostics"))
	} else {
		log.Info("eng Workstation Diagnostics")
	}

	requiredOK := checkTools()
	checkConfigPaths()
	checkTelemetry()
	checkVersionStatus()

	if requiredOK {
		theme.SuccessMessage("Doctor PASS: all required tools found.")
		return nil
	}
	theme.WarningMessage(
		"Doctor FAIL: some required tools are missing. Run `eng system setup` to install prerequisites.",
	)
	return fmt.Errorf("doctor found missing required tools")
}

func checkTools() bool {
	tools := []ToolCheck{
		{Name: "Git", Binary: "git", Description: "Version control system", Required: true},
		{Name: "Homebrew", Binary: "brew", Description: "Package manager", Required: true},
		{Name: "Bash", Binary: "bash", Description: "Command shell", Required: true},
		{Name: "Docker", Binary: "docker", Description: "Container runtime for compose swarms", Required: false},
		{Name: "asdf", Binary: "asdf", Description: "Runtime version manager", Required: false},
		{Name: "Tailscale", Binary: "tailscale", Description: "Mesh VPN CLI", Required: false},
		{Name: "Bitwarden CLI", Binary: "bw", Description: "Secrets management", Required: false},
		{Name: "GPG", Binary: "gpg", Description: "Commit signing & encryption", Required: false},
		{Name: "GitHub CLI", Binary: "gh", Description: "GitHub integrations", Required: false},
		{Name: "GitLab CLI", Binary: "glab", Description: "GitLab integrations", Required: false},
	}

	var cardLines []string
	cardLines = append(cardLines, theme.BoldText.Render("CLI Tools & Dependencies:"))

	allRequiredPassed := true
	for _, tool := range tools {
		path, err := execLookPath(tool.Binary)
		var statusText string
		if err == nil {
			statusText = fmt.Sprintf(
				"  %s %-16s %s",
				theme.SuccessText.Render("✔"),
				tool.Name,
				theme.MutedText.Render(path),
			)
		} else if tool.Required {
			allRequiredPassed = false
			statusText = fmt.Sprintf(
				"  %s %-16s %s",
				theme.ErrorText.Render("✖"),
				tool.Name,
				theme.ErrorText.Render("Not Found (Required)"),
			)
		} else {
			statusText = fmt.Sprintf(
				"  %s %-16s %s",
				theme.MutedText.Render("○"),
				tool.Name,
				theme.MutedText.Render("Not Installed (Optional)"),
			)
		}
		cardLines = append(cardLines, statusText)
	}

	if !ui.DisableProgress {
		fmt.Fprintln(log.Out, theme.InfoBox.Render(strings.Join(cardLines, "\n")))
	} else {
		for _, l := range cardLines {
			log.Message("%s", l)
		}
		log.Message("")
	}

	if !allRequiredPassed {
		log.Warn("Some required tools are missing. Run `eng system setup` to install prerequisites.")
	}
	return allRequiredPassed
}

func checkConfigPaths() {
	var cardLines []string
	cardLines = append(cardLines, theme.BoldText.Render("Configuration & Workspaces:"))

	// Check config file
	cfgFile := viper.ConfigFileUsed()
	if cfgFile != "" {
		if _, err := osStat(cfgFile); err == nil {
			cardLines = append(
				cardLines,
				fmt.Sprintf(
					"  %s %-20s %s",
					theme.SuccessText.Render("✔"),
					"Config File:",
					theme.MutedText.Render(cfgFile),
				),
			)
		} else {
			cardLines = append(
				cardLines,
				fmt.Sprintf(
					"  %s %-20s %s",
					theme.ErrorText.Render("✖"),
					"Config File:",
					theme.ErrorText.Render("Not Accessible"),
				),
			)
		}
	} else {
		cardLines = append(
			cardLines,
			fmt.Sprintf(
				"  %s %-20s %s",
				theme.MutedText.Render("○"),
				"Config File:",
				theme.MutedText.Render("None loaded (using defaults)"),
			),
		)
	}

	// Check Git Dev Path
	gitCfg := config.GetGitConfig()
	devPath := gitCfg.DevPath
	if devPath != "" {
		devPath = os.ExpandEnv(devPath)
		if fi, err := osStat(devPath); err == nil && fi.IsDir() {
			cardLines = append(
				cardLines,
				fmt.Sprintf(
					"  %s %-20s %s",
					theme.SuccessText.Render("✔"),
					"Git Dev Path:",
					theme.MutedText.Render(devPath),
				),
			)
		} else {
			cardLines = append(
				cardLines,
				fmt.Sprintf(
					"  %s %-20s %s",
					theme.ErrorText.Render("✖"),
					"Git Dev Path:",
					theme.ErrorText.Render(fmt.Sprintf("%s (Directory not found)", devPath)),
				),
			)
		}
	} else {
		cardLines = append(
			cardLines,
			fmt.Sprintf(
				"  %s %-20s %s",
				theme.MutedText.Render("○"),
				"Git Dev Path:",
				theme.MutedText.Render("Not set (set with 'eng config git-dev-path')"),
			),
		)
	}

	// Check Dotfiles Bare Repo Path
	dotfilesCfg := config.GetDotfilesConfig()
	barePath := dotfilesCfg.BareRepoPath
	if barePath != "" {
		barePath = os.ExpandEnv(barePath)
		if fi, err := osStat(barePath); err == nil && fi.IsDir() {
			cardLines = append(
				cardLines,
				fmt.Sprintf(
					"  %s %-20s %s",
					theme.SuccessText.Render("✔"),
					"Dotfiles Bare Path:",
					theme.MutedText.Render(barePath),
				),
			)
		} else {
			cardLines = append(
				cardLines,
				fmt.Sprintf(
					"  %s %-20s %s",
					theme.MutedText.Render("○"),
					"Dotfiles Bare Path:",
					theme.MutedText.Render(fmt.Sprintf("%s (Not found)", barePath)),
				),
			)
		}
	}

	if !ui.DisableProgress {
		fmt.Fprintln(log.Out, theme.InfoBox.Render(strings.Join(cardLines, "\n")))
	} else {
		for _, l := range cardLines {
			log.Message("%s", l)
		}
		log.Message("")
	}
}

func checkTelemetry() {
	var cardLines []string
	cardLines = append(cardLines, theme.BoldText.Render("Telemetry & Diagnostics:"))

	cfg := config.GetEffectiveTelemetryConfig()
	if cfg.Enabled {
		cardLines = append(
			cardLines,
			fmt.Sprintf(
				"  %s %-20s %s",
				theme.SuccessText.Render("✔"),
				"Status:",
				theme.SuccessText.Render("Enabled (OpenPanel)"),
			),
		)
		cardLines = append(
			cardLines,
			fmt.Sprintf(
				"  %s %-20s %s",
				theme.SuccessText.Render("✔"),
				"Endpoint:",
				theme.MutedText.Render(cfg.APIURL),
			),
		)

		// Test connection reachability
		status, err := telemetry.TestConnection(cfg)
		if err == nil {
			cardLines = append(
				cardLines,
				fmt.Sprintf(
					"  %s %-20s %s",
					theme.SuccessText.Render("✔"),
					"Reachability:",
					theme.SuccessText.Render(fmt.Sprintf("Connected (HTTP %d)", status)),
				),
			)
		} else {
			cardLines = append(
				cardLines,
				fmt.Sprintf(
					"  %s %-20s %s",
					theme.MutedText.Render("○"),
					"Reachability:",
					theme.MutedText.Render("Offline / Endpoint unreachable"),
				),
			)
		}
	} else if !config.IsTelemetryConfigured() {
		cardLines = append(
			cardLines,
			fmt.Sprintf(
				"  %s %-20s %s",
				theme.MutedText.Render("○"),
				"Status:",
				theme.MutedText.Render("Disabled (Unset credentials / Dev build)"),
			),
		)
	} else {
		cardLines = append(
			cardLines,
			fmt.Sprintf(
				"  %s %-20s %s",
				theme.MutedText.Render("○"),
				"Status:",
				theme.MutedText.Render("Disabled (DNT/Opt-Out)"),
			),
		)
	}

	if !ui.DisableProgress {
		fmt.Fprintln(log.Out, theme.InfoBox.Render(strings.Join(cardLines, "\n")))
	} else {
		for _, l := range cardLines {
			log.Message("%s", l)
		}
		log.Message("")
	}
}

func checkVersionStatus() {
	var cardLines []string
	cardLines = append(cardLines, theme.BoldText.Render("Runtime & Installation:"))

	cardLines = append(
		cardLines,
		fmt.Sprintf("  %-22s %s", theme.BoldText.Render("Current Version:"), theme.PrimaryText.Render(version.Version)),
	)
	cardLines = append(
		cardLines,
		fmt.Sprintf("  %-22s %s", theme.BoldText.Render("Build Commit:"), theme.MutedText.Render(version.Commit)),
	)
	cardLines = append(
		cardLines,
		fmt.Sprintf("  %-22s %s", theme.BoldText.Render("Build Date:"), theme.MutedText.Render(version.Date)),
	)

	execPath, err := os.Executable()
	if err == nil {
		resolvedPath, err := filepath.EvalSymlinks(execPath)
		if err == nil {
			execPath = resolvedPath
		}
		cardLines = append(
			cardLines,
			fmt.Sprintf("  %-22s %s", theme.BoldText.Render("Executable:"), theme.MutedText.Render(execPath)),
		)
	}

	if !ui.DisableProgress {
		fmt.Fprintln(log.Out, theme.InfoBox.Render(strings.Join(cardLines, "\n")))
	} else {
		for _, l := range cardLines {
			log.Message("%s", l)
		}
		log.Message("")
	}
}
