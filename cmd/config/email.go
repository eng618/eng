package config

import (
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/utils/config"
	"github.com/eng618/eng/internal/utils/log"
)

// EmailCmd defines the command for setting the user's email in the configuration.
var EmailCmd = &cobra.Command{
	Use:   "email",
	Short: "Update config email",
	Long:  `This command should validate and update the email config.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Checking for email in config file...")
		config.Email()
	},
}
