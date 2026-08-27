package config

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/telemetry"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

// TelemetryCmd defines the command for managing telemetry and analytics preferences.
var TelemetryCmd = &cobra.Command{
	Use:     "telemetry",
	Aliases: []string{"analytics", "metrics"},
	Short:   "Manage OpenPanel telemetry and analytics settings",
	Long: `Configure anonymous telemetry reporting to your OpenPanel instance.
Telemetry helps track command usage, execution performance, and reliability insights.

Privacy respects: The DO_NOT_TRACK=1 and ENG_TELEMETRY_DISABLED=1 environment variables
are honored globally.`,
	Run: func(cmd *cobra.Command, _args []string) {
		config.Telemetry()
	},
}

var telemetryEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable anonymous telemetry reporting",
	RunE: func(cmd *cobra.Command, _args []string) error {
		if err := config.SetTelemetryEnabled(true); err != nil {
			return err
		}
		log.Success("Telemetry enabled successfully.")
		return nil
	},
}

var telemetryDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable anonymous telemetry reporting",
	RunE: func(cmd *cobra.Command, _args []string) error {
		if err := config.SetTelemetryEnabled(false); err != nil {
			return err
		}
		log.Success("Telemetry disabled successfully. No events will be sent.")
		return nil
	},
}

var telemetryTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test connection to the OpenPanel telemetry instance",
	RunE: func(cmd *cobra.Command, _args []string) error {
		if !config.IsTelemetryConfigured() {
			log.Warn("OpenPanel credentials are not set in this build.")
			log.Info("Set OPENPANEL_CLIENT_ID and OPENPANEL_CLIENT_SECRET environment variables to test locally.")
			return nil
		}

		cfg := config.GetEffectiveTelemetryConfig()
		log.Start("Testing connection to OpenPanel at %s...", cfg.APIURL)

		status, err := telemetry.TestConnection(cfg)
		if err != nil {
			log.Error("Failed to reach OpenPanel: %v", err)
			return err
		}

		log.Success("Successfully connected to OpenPanel! (HTTP %d)", status)
		return nil
	},
}

var telemetryStatusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"show", "info"},
	Short:   "Display active telemetry settings",
	Run: func(cmd *cobra.Command, _args []string) {
		PrintTelemetryStatus()
	},
}

// PrintTelemetryStatus prints a styled summary card of telemetry configuration.
func PrintTelemetryStatus() {
	cfg := config.GetEffectiveTelemetryConfig()

	var cardLines []string
	cardLines = append(cardLines, theme.BoldText.Render("Telemetry & Analytics Status:"))

	var enabledText string
	if cfg.Enabled {
		enabledText = theme.SuccessText.Render("Enabled (Active)")
	} else if !config.IsTelemetryConfigured() {
		enabledText = theme.MutedText.Render("Disabled (Unset credentials / Dev build)")
	} else {
		enabledText = theme.MutedText.Render("Disabled (DNT/Opt-Out)")
	}
	cardLines = append(cardLines, fmt.Sprintf("  %-20s %s", theme.BoldText.Render("Status:"), enabledText))
	cardLines = append(cardLines, fmt.Sprintf("  %-20s %s", theme.BoldText.Render("API URL:"), theme.BaseText.Render(cfg.APIURL)))
	cardLines = append(cardLines, fmt.Sprintf("  %-20s %s", theme.BoldText.Render("Profile ID:"), theme.MutedText.Render(cfg.ProfileID)))

	if !ui.DisableProgress {
		fmt.Fprintln(log.Out, theme.InfoBox.Render(strings.Join(cardLines, "\n")))
	} else {
		for _, l := range cardLines {
			log.Message("%s", l)
		}
		log.Message("")
	}
}

func init() {
	TelemetryCmd.AddCommand(telemetryEnableCmd)
	TelemetryCmd.AddCommand(telemetryDisableCmd)
	TelemetryCmd.AddCommand(telemetryTestCmd)
	TelemetryCmd.AddCommand(telemetryStatusCmd)
}
