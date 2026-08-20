package config

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

// IdeURL checks for the Antigravity IDE download URL in the configuration and prompts the user to confirm or set it.
// If an optionalURL argument is passed, it sets the URL directly.
func IdeURL(optionalURL ...string) string {
	if len(optionalURL) > 0 && strings.TrimSpace(optionalURL[0]) != "" {
		newURL := strings.TrimSpace(optionalURL[0])
		SetIdeURL(newURL)
		return newURL
	}

	log.Start("Checking for Antigravity IDE download URL in config")
	currentURL := viper.GetString("antigravity.ide_download_url")

	if currentURL == "" {
		return updateIdeURL()
	}

	confirmed, err := ui.Confirm(
		fmt.Sprintf("Confirm Antigravity IDE download URL: %s?", theme.PrimaryText.Render(currentURL)),
		true,
	)
	cobra.CheckErr(err)

	if !confirmed {
		return updateIdeURL()
	}

	log.Success("Confirmed Antigravity IDE download URL")
	return currentURL
}

func updateIdeURL() string {
	inputURL, err := ui.Input("Enter Antigravity IDE download URL (e.g. https://.../Antigravity-IDE.tar.gz):", "")
	cobra.CheckErr(err)

	SetIdeURL(inputURL)
	return inputURL
}

// SetIdeURL updates the antigravity.ide_download_url setting in viper and writes to config.
func SetIdeURL(url string) {
	viper.Set("antigravity.ide_download_url", strings.TrimSpace(url))

	if err := viper.WriteConfig(); err != nil {
		cobra.CheckErr(
			fmt.Errorf(
				"%s: %w",
				lipgloss.NewStyle().Foreground(theme.Destructive).Render("Error writing config file"),
				err,
			),
		)
	}
	log.Success("Antigravity IDE download URL updated successfully")
}
