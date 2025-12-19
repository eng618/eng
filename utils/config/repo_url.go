package config

import (
	"errors"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/utils/log"
)

// RepoURL checks for the dotfiles repository URL in the configuration and returns it.
// If the URL is not found, the function will call UpdateRepoURL() to update it.
func RepoURL() string {
	repoURL := viper.GetString("dotfiles.repo_url")

	if repoURL == "" {
		UpdateRepoURL()
		repoURL = viper.GetString("dotfiles.repo_url")
	}

	return repoURL
}

// UpdateRepoURL prompts the user to input their dotfiles repository URL, updates the
// configuration with the provided URL, and saves the updated configuration
// back to the configuration file.
func UpdateRepoURL() {
	var url string
	prompt := &survey.Input{
		Message: "What is your dotfiles repository URL? (e.g., https://github.com/username/dotfiles.git)",
	}
	err := survey.AskOne(prompt, &url)
	cobra.CheckErr(err)

	viper.Set("dotfiles.repo_url", url)

	// Save the updated configuration back to the file
	if err := viper.WriteConfig(); err != nil {
		err := errors.New(color.RedString("Error writing config file: %v", err))
		cobra.CheckErr(err)
	}
	log.Success("Configuration updated successfully")
}
