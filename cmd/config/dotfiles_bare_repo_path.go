package config

import (
	"github.com/eng618/eng/utils/config"
	"github.com/spf13/cobra"
)

// DotfilesBareRepoPathCmd represents the command to manage the bare repository path configuration.
var DotfilesBareRepoPathCmd = &cobra.Command{
	Use:   "dotfiles-bare-repo-path",
	Short: "Update config dotfiles bare repo path",
	Long:  `Get or set the path where the bare dotfiles repository should be stored.`,
	Run: func(cmd *cobra.Command, args []string) {
		config.BareRepoPath()
	},
}
