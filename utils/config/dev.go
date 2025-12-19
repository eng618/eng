package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/utils/log"
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
		var dConfirm bool
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Confirm development folder path: %s?", color.CyanString(devPath)),
		}
		prompt.Default = true
		err := survey.AskOne(prompt, &dConfirm)
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
	var d string
	prompt := &survey.Input{
		Message: "What is your development folder path?",
		Default: os.ExpandEnv("$HOME/Development"),
	}
	err := survey.AskOne(prompt, &d)
	cobra.CheckErr(err)

	viper.Set("git.dev_path", d)

	// Save the updated configuration back to the file
	if err := viper.WriteConfig(); err != nil {
		err := errors.New(color.RedString("Error writing config file: %v", err))
		cobra.CheckErr(err)
	}
	log.Success("Configuration updated successfully")
}
