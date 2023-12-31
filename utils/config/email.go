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

// Email checks the config file for an email, or prompts to create one.
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

// GetEmail simplified command to ensure there is an email listed in the config,
// and update if not present.
func GetEmail() {
	log.Start("Checking for user email.")

	// Check for email defined in configs
	email := viper.GetString("user-email")

	if email == "" {
		updateEmail()
	}

	log.Success("Found user email to be: %s", email)
}

// updateEmail prompts user to input email, than updates the config file with value.
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
