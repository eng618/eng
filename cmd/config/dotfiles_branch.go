package config

import (
	"github.com/spf13/cobra"

	"github.com/eng618/eng/utils/config"
)

// DotfilesBranchCmd represents the command to manage the dotfiles branch configuration.
var DotfilesBranchCmd = &cobra.Command{
	Use:   "dotfiles-branch",
	Short: "Update config dotfiles branch",
	Long:  `Get or set the dotfiles branch (main/work/server) in the configuration file.`,
	Run: func(cmd *cobra.Command, args []string) {
		config.Branch()
	},
}
