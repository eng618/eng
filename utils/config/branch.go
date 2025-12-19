package config

import (
	"errors"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/utils/log"
)

// Branch checks for the dotfiles branch in the configuration and returns it.
// If the branch is not found, the function will call UpdateBranch() to update it.
func Branch() string {
	branch := viper.GetString("dotfiles.branch")

	if branch == "" {
		UpdateBranch()
		branch = viper.GetString("dotfiles.branch")
	}

	return branch
}

// UpdateBranch prompts the user to select their dotfiles branch, updates the
// configuration with the selected branch, and saves the updated configuration
// back to the configuration file.
func UpdateBranch() {
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
