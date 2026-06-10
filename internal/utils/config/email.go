package config

import (
	"errors"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/utils/log"
)

// Email checks for the user's email in the configuration and prompts the user to confirm it.
// If the email is not found or the user does not confirm it, the function will call updateEmail() to update the email.
// It logs the start and success of the email checking process and returns the confirmed email as a string.
func Email() string {
	log.Start("Checking for email")

	// Check for email defined in configs
	email := viper.GetString("user-email")

	if email == "" {
		updateEmail()
	} else {
		// Verify this is the correct email they are expecting to use.
		var eConfirm bool
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Confirm email: %s?", color.CyanString(email)),
		}
		prompt.Default = true
		err := survey.AskOne(prompt, &eConfirm)
		cobra.CheckErr(err)

		if !eConfirm {
			updateEmail()
		}
	}

	log.Success("Confirmed email")
	return email
}

// GetEmail retrieves the user's email from the configuration.
// If the email is not found in the configuration, it prompts the user to update it.
// Logs the process of checking and finding the email.
func GetEmail() {
	log.Start("Checking for user email.")

	// Check for email defined in configs
	email := viper.GetString("user-email")

	if email == "" {
		updateEmail()
	}

	log.Success("Found user email to be: %s", email)
}

// updateEmail prompts the user to input their email address, updates the
// configuration with the provided email, and saves the updated configuration
// back to the configuration file. If any error occurs during the process,
// it is handled appropriately.
func updateEmail() {
	var e string
	prompt := &survey.Input{
		Message: "What is your email?",
	}
	err := survey.AskOne(prompt, &e)
	cobra.CheckErr(err)

	viper.Set("user-email", e)

	// Save the updated configuration back to the file
	if err := viper.WriteConfig(); err != nil {
		err := errors.New(color.RedString("Error writing config file: %v", err))
		cobra.CheckErr(err)
	}
	log.Success("Configuration updated successfully")
}
