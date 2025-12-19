package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/utils/log"
)

// BareRepoPath checks for the bare repository path in the configuration and returns it.
// If the path is not found, the function will call UpdateBareRepoPath() to update it.
func BareRepoPath() string {
	bareRepoPath := viper.GetString("dotfiles.bare_repo_path")

	if bareRepoPath == "" {
		UpdateBareRepoPath()
		bareRepoPath = viper.GetString("dotfiles.bare_repo_path")
	}

	return os.ExpandEnv(bareRepoPath)
}

// UpdateBareRepoPath prompts the user to input their bare repository path, updates the
// configuration with the provided path, and saves the updated configuration
// back to the configuration file.
func UpdateBareRepoPath() {
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
