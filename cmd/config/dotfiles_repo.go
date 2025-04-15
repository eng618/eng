package config

import (
	"github.com/eng618/eng/utils/config"
	"github.com/eng618/eng/utils/log"
	"github.com/spf13/cobra"
)

var DotfilesRepoCmd = &cobra.Command{
	Use:   "dotfiles-repo",
	Short: "Set dotfiles repo path",
	Long:  `This command sets the path to the local dotfiles repository in the config file.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Checking for dotfile repo path in config file...")
		config.DotfilesRepo()
	},
}
