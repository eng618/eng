package config

import (
	"github.com/eng618/eng/utils/config"
	"github.com/spf13/cobra"
)

// DotfilesRepoURLCmd represents the command to manage the dotfiles repository URL configuration.
var DotfilesRepoURLCmd = &cobra.Command{
	Use:   "dotfiles-repo-url",
	Short: "Get or set the dotfiles repository URL",
	Long:  `Get or set the dotfiles repository URL in the configuration file.`,
	Run: func(cmd *cobra.Command, args []string) {
		config.GetRepoURL()
	},
}
