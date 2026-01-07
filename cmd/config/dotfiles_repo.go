package config

import (
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/utils/config"
	"github.com/eng618/eng/internal/utils/log"
)

// DotfilesRepoCmd defines the command for setting the local dotfiles repository path.
var DotfilesRepoCmd = &cobra.Command{
	Use:   "dotfiles-repo",
	Short: "Update config dotfiles repo path",
	Long:  `This command sets the path to the local dotfiles repository in the config file.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Checking for dotfile repo path in config file...")
		config.DotfilesRepo()
	},
}
