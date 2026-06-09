package config

import (
	"errors"
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
)

// Verbose checks for the verbose setting in the configuration and prompts the user to confirm it.
// If the verbose setting is not found or the user does not confirm it, the function will call updateVerbose() to update the setting.
// It logs the start and success of the verbose checking process and returns the confirmed verbose setting as a bool.
func Verbose() bool {
	log.Start("Checking for verbose setting")

	// Check for verbose defined in configs
	verbose := viper.GetBool("verbose")

	// Verify this is the correct verbose setting they are expecting to use.
	vConfirm, err := ui.Confirm(
		fmt.Sprintf("Confirm verbose mode: %s?", color.CyanString(fmt.Sprintf("%t", verbose))),
		verbose,
	)
	cobra.CheckErr(err)

	if !vConfirm {
		updateVerbose()
		verbose = viper.GetBool("verbose")
	}

	log.Success("Confirmed verbose setting")
	return verbose
}

// GetVerbose retrieves the verbose setting from the configuration.
// If the verbose setting is not found in the configuration, it prompts the user to update it.
// Logs the process of checking and finding the verbose setting.
func GetVerbose() {
	log.Start("Checking for verbose setting.")

	// Check for verbose defined in configs
	verbose := viper.GetBool("verbose")

	log.Success("Found verbose setting to be: %t", verbose)
}

// updateVerbose prompts the user to choose their verbose setting, updates the
// configuration with the provided setting, and saves the updated configuration
// back to the configuration file. If any error occurs during the process,
// it is handled appropriately.
func updateVerbose() {
	verbose, err := ui.Confirm("Enable verbose mode?", false)
	cobra.CheckErr(err)

	viper.Set("verbose", verbose)

	// Save the updated configuration back to the file
	if err := viper.WriteConfig(); err != nil {
		err := errors.New(color.RedString("Error writing config file: %v", err))
		cobra.CheckErr(err)
	}
	log.Success("Configuration updated successfully")
}
