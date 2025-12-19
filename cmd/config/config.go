// Package config provides subcommands for managing the eng CLI configuration.
// It allows users to view and modify settings stored in the $HOME/.eng.yaml file.
package config

import (
	"github.com/eng618/eng/utils/config"
	"github.com/spf13/cobra"
)

// ConfigCmd represents the base command for all configuration related operations.
var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage the cli's config file.",
	Long: `This command is used to facilitate the management of the config file specific to this cli.

It should be located at $HOME/.eng.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		// Validate all configuration settings one by one
		config.Email()
		config.DotfilesRepo()
		config.RepoURL()
		config.Branch()
		config.BareRepoPath()
		config.GitDevPath()
		config.Verbose()
	},
}

func init() {
	ConfigCmd.AddCommand(EmailCmd)
	ConfigCmd.AddCommand(DotfilesRepoCmd)
	ConfigCmd.AddCommand(DotfilesRepoURLCmd)
	ConfigCmd.AddCommand(DotfilesBranchCmd)
	ConfigCmd.AddCommand(DotfilesBareRepoPathCmd)
	ConfigCmd.AddCommand(GitDevPathCmd)
	ConfigCmd.AddCommand(VerboseCmd)
}
