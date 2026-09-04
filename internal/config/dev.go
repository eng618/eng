package config

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui/theme"
)

// GitDevPath checks for the development folder path in the configuration and prompts the user to confirm it.
// If the path is not found or the user does not confirm it, the function will call updateGitDevPath() to update the path.
// It logs the start and success of the path checking process and returns the confirmed path as a string.
func GitDevPath() string {
	log.Start("Checking for development folder path")

	// Check for dev path defined in configs
	devPath := viper.GetString("git.dev_path")

	if devPath == "" {
		updateGitDevPath()
	} else {
		// Verify this is the correct dev path they are expecting to use.
		dConfirm, err := ConfirmPrompt(
			fmt.Sprintf("Confirm development folder path: %s?", theme.PrimaryText.Render(devPath)),
			true,
		)
		cobra.CheckErr(err)

		if !dConfirm {
			updateGitDevPath()
		}
	}

	log.Success("Confirmed development folder path")
	return devPath
}

// updateGitDevPath prompts the user to input their development folder path, updates the
// configuration with the provided path, and saves the updated configuration
// back to the configuration file. If any error occurs during the process,
// it is handled appropriately.
func updateGitDevPath() {
	d, err := InputPrompt("What is your development folder path?", os.ExpandEnv("$HOME/Development"))
	cobra.CheckErr(err)

	viper.Set("git.dev_path", d)

	// Save the updated configuration back to the file
	if err := viper.WriteConfig(); err != nil {
		cobra.CheckErr(
			fmt.Errorf(
				"%s: %w",
				lipgloss.NewStyle().Foreground(theme.Destructive).Render("Error writing config file"),
				err,
			),
		)
	}
	log.Success("Configuration updated successfully")
}
