package config

import (
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/utils/config"
)

// DotfilesRepoURLCmd represents the command to manage the dotfiles repository URL configuration.
var DotfilesRepoURLCmd = &cobra.Command{
	Use:   "dotfiles-repo-url",
	Short: "Update config dotfiles repo URL",
	Long:  `Get or set the dotfiles repository URL in the configuration file.`,
	Run: func(cmd *cobra.Command, args []string) {
		config.RepoURL()
	},
}
