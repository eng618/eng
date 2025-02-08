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

// DotfilesRepo checks for the dotfiles repository path in the configuration and prompts the user to confirm it.
// If the path is not found or the user does not confirm it, the function will call updateDotfilesRepo() to update the path.
// It logs the start and success of the path checking process and returns the confirmed path as a string.
func DotfilesRepo() string {
	log.Start("Checking for dotfiles repository path")

	// Check for repo path defined in configs
	repoPath := viper.GetString("dotfiles.repoPath")

	if repoPath == "" {
		updateDotfilesRepo()
	} else {
		// Verify this is the correct repo path they are expecting to use.
		var rConfirm bool
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Confirm dotfiles repository path: %s?", color.CyanString(repoPath)),
		}
		prompt.Default = true
		err := survey.AskOne(prompt, &rConfirm)
		cobra.CheckErr(err)

		if !rConfirm {
			updateDotfilesRepo()
		}
	}

	log.Success("Confirmed dotfiles repository path")
	return repoPath
}

// GetDotfilesRepo retrieves the dotfiles repository path from the configuration.
// If the path is not found in the configuration, it prompts the user to update it.
// Logs the process of checking and finding the path.
func GetDotfilesRepo() {
	log.Start("Checking for dotfiles repository path.")

	// Check for repoPath defined in configs
	repoPath := viper.GetString("dotfiles.repoPath")

	if repoPath == "" {
		updateDotfilesRepo()
	}

	log.Success("Found dotfiles repository path to be: %s", repoPath)
}

// updateDotfilesRepo prompts the user to input their dotfiles repository path, updates the
// configuration with the provided path, and saves the updated configuration
// back to the configuration file. If any error occurs during the process,
// it is handled appropriately.
func updateDotfilesRepo() {
	var r string
	prompt := &survey.Input{
		Message: "What is your dotfiles repository path?",
	}
	err := survey.AskOne(prompt, &r)
	cobra.CheckErr(err)

	viper.Set("dotfiles.repoPath", r)
	viper.Set("dotfiles.workTree", os.Getenv("HOME"))

	// Save the updated configuration back to the file
	if err := viper.WriteConfig(); err != nil {
		err := errors.New(color.RedString("Error writing config file: %v", err))
		cobra.CheckErr(err)
	}
	log.Success("Configuration updated successfully")
}
