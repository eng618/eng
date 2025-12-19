package config

import (
	"github.com/eng618/eng/utils/config"
	"github.com/eng618/eng/utils/log"
	"github.com/spf13/cobra"
)

// VerboseCmd defines the command for managing the verbose output setting.
var VerboseCmd = &cobra.Command{
	Use:   "verbose",
	Short: "Update config verbose setting",
	Long:  `This command should validate and update the verbose config setting.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Checking for verbose setting in config file...")
		config.Verbose()
	},
}
