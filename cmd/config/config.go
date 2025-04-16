package config

import (
	"fmt"

	"github.com/spf13/cobra"
)

var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage the cli's config file config file.",
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
