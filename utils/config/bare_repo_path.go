package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/utils/log"
)

// BareRepoPath checks for the bare repository path in the configuration and prompts the user to confirm it.
// If the path is not found or the user does not confirm it, the function will call updateBareRepoPath() to update the path.
// It logs the start and success of the path checking process and returns the confirmed path as a string.
func BareRepoPath() string {
	log.Start("Checking for bare repository path")

	// Check for bare repo path defined in configs
	bareRepoPath := viper.GetString("dotfiles.bare_repo_path")

	if bareRepoPath == "" {
		updateBareRepoPath()
		bareRepoPath = viper.GetString("dotfiles.bare_repo_path")
	} else {
		// Expand environment variables in the path
		bareRepoPath = os.ExpandEnv(bareRepoPath)

		// Verify this is the correct path they are expecting to use.
		var confirm bool
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Confirm bare repository path: %s?", color.CyanString(bareRepoPath)),
		}
		prompt.Default = true
		err := survey.AskOne(prompt, &confirm)
		cobra.CheckErr(err)

		if !confirm {
			updateBareRepoPath()
			bareRepoPath = viper.GetString("dotfiles.bare_repo_path")
			bareRepoPath = os.ExpandEnv(bareRepoPath)
		}
	}

	log.Success("Confirmed bare repository path")
	return bareRepoPath
}

// GetBareRepoPath retrieves the bare repository path from the configuration.
// If the path is not found in the configuration, it prompts the user to update it.
// Logs the process of checking and finding the path.
func GetBareRepoPath() {
	log.Start("Checking for bare repository path.")

	// Check for path defined in configs
	bareRepoPath := viper.GetString("dotfiles.bare_repo_path")

	if bareRepoPath == "" {
		updateBareRepoPath()
		bareRepoPath = viper.GetString("dotfiles.bare_repo_path")
	}

	bareRepoPath = os.ExpandEnv(bareRepoPath)
	log.Success("Found bare repository path to be: %s", bareRepoPath)
}

// updateBareRepoPath prompts the user to input their bare repository path, updates the
// configuration with the provided path, and saves the updated configuration
// back to the configuration file. If any error occurs during the process,
// it is handled appropriately.
func updateBareRepoPath() {
	homeDir, err := os.UserHomeDir()
	cobra.CheckErr(err)

	defaultPath := filepath.Join(homeDir, ".eng-cfg")

	var path string
	prompt := &survey.Input{
		Message: "Where should the bare repository be stored?",
		Default: defaultPath,
	}
	err = survey.AskOne(prompt, &path)
	cobra.CheckErr(err)

	viper.Set("dotfiles.bare_repo_path", path)

	// Save the updated configuration back to the file
	if err := viper.WriteConfig(); err != nil {
		err := errors.New(color.RedString("Error writing config file: %v", err))
		cobra.CheckErr(err)
	}
	log.Success("Configuration updated successfully")
}
