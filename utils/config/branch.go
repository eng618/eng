package config

import (
	"errors"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/utils/log"
)

// Branch checks for the dotfiles branch in the configuration and prompts the user to confirm it.
// If the branch is not found or the user does not confirm it, the function will call updateBranch() to update the branch.
// It logs the start and success of the branch checking process and returns the confirmed branch as a string.
func Branch() string {
	log.Start("Checking for dotfiles branch")

	// Check for branch defined in configs
	branch := viper.GetString("dotfiles.branch")

	if branch == "" {
		updateBranch()
		branch = viper.GetString("dotfiles.branch")
	} else {
		// Verify this is the correct branch they are expecting to use.
		var confirm bool
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Confirm dotfiles branch: %s?", color.CyanString(branch)),
		}
		prompt.Default = true
		err := survey.AskOne(prompt, &confirm)
		cobra.CheckErr(err)

		if !confirm {
			updateBranch()
			branch = viper.GetString("dotfiles.branch")
		}
	}

	log.Success("Confirmed dotfiles branch")
	return branch
}

// GetBranch retrieves the dotfiles branch from the configuration.
// If the branch is not found in the configuration, it prompts the user to update it.
// Logs the process of checking and finding the branch.
func GetBranch() {
	log.Start("Checking for dotfiles branch.")

	// Check for branch defined in configs
	branch := viper.GetString("dotfiles.branch")

	if branch == "" {
		updateBranch()
	}

	log.Success("Found dotfiles branch to be: %s", branch)
}

// updateBranch prompts the user to select their dotfiles branch, updates the
// configuration with the selected branch, and saves the updated configuration
// back to the configuration file. If any error occurs during the process,
// it is handled appropriately.
func updateBranch() {
	var branch string
	prompt := &survey.Select{
		Message: "Which branch should be used for dotfiles?",
		Options: []string{"main", "work", "server"},
		Default: "main",
	}
	err := survey.AskOne(prompt, &branch)
	cobra.CheckErr(err)

	viper.Set("dotfiles.branch", branch)

	// Save the updated configuration back to the file
	if err := viper.WriteConfig(); err != nil {
		err := errors.New(color.RedString("Error writing config file: %v", err))
		cobra.CheckErr(err)
	}
	log.Success("Configuration updated successfully")
}
