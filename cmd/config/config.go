// Package config provides subcommands for managing the eng CLI configuration.
// It allows users to view and modify settings stored in the $HOME/.eng.yaml file.
package config

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ConfigCmd represents the base command for all configuration related operations.
var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage the cli's config file.",
	Long: `This command is used to facilitate the management of the config file specific to this cli. 

It should be located at $HOME/.eng.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("config called")
	},
}

func init() {
	ConfigCmd.AddCommand(EmailCmd)
	ConfigCmd.AddCommand(DotfilesRepoCmd)
	ConfigCmd.AddCommand(VerboseCmd)
}
