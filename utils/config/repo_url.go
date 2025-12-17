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

// RepoURL checks for the dotfiles repository URL in the configuration and prompts the user to confirm it.
// If the URL is not found or the user does not confirm it, the function will call updateRepoURL() to update the URL.
// It logs the start and success of the URL checking process and returns the confirmed URL as a string.
func RepoURL() string {
	log.Start("Checking for dotfiles repository URL")

	// Check for repo URL defined in configs
	repoURL := viper.GetString("dotfiles.repo_url")

	if repoURL == "" {
		updateRepoURL()
		repoURL = viper.GetString("dotfiles.repo_url")
	} else {
		// Verify this is the correct URL they are expecting to use.
		var confirm bool
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Confirm dotfiles repository URL: %s?", color.CyanString(repoURL)),
		}
		prompt.Default = true
		err := survey.AskOne(prompt, &confirm)
		cobra.CheckErr(err)

		if !confirm {
			updateRepoURL()
			repoURL = viper.GetString("dotfiles.repo_url")
		}
	}

	log.Success("Confirmed dotfiles repository URL")
	return repoURL
}

// GetRepoURL retrieves the dotfiles repository URL from the configuration.
// If the URL is not found in the configuration, it prompts the user to update it.
// Logs the process of checking and finding the URL.
func GetRepoURL() {
	log.Start("Checking for dotfiles repository URL.")

	// Check for URL defined in configs
	repoURL := viper.GetString("dotfiles.repo_url")

	if repoURL == "" {
		updateRepoURL()
	}

	log.Success("Found dotfiles repository URL to be: %s", repoURL)
}

// updateRepoURL prompts the user to input their dotfiles repository URL, updates the
// configuration with the provided URL, and saves the updated configuration
// back to the configuration file. If any error occurs during the process,
// it is handled appropriately.
func updateRepoURL() {
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
