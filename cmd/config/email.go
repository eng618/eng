package config

import (
	"github.com/eng618/eng/utils/config"
	"github.com/eng618/eng/utils/log"
	"github.com/spf13/cobra"
)

var EmailCmd = &cobra.Command{
	Use:   "email",
	Short: "Update config email",
	Long:  `This command should validate and update the email config.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Checking for email in config file...")
		config.Email()
	},
}
