// Package dotfiles provides cobra commands for managing user dotfiles
// via a bare git repository.
package dotfiles

import (
	"os"

	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// DotfilesCmd serves as the base command for all dotfiles related operations.
// It doesn't perform any action itself but groups subcommands like sync and fetch.
var DotfilesCmd = &cobra.Command{
	Use:     "dotfiles",
	Short:   "Manage dotfiles",
	Long:    `This command is used to facilitate the management of private hidden dot files.`,
	Aliases: []string{"cfg"},
	Run: func(cmd *cobra.Command, args []string) {
		showInfo, _ := cmd.Flags().GetBool("info")
		isVerbose := utils.IsVerbose(cmd)

		if showInfo {
			log.Info("Current dotfiles configuration:")
			repoPath := os.ExpandEnv(viper.GetString("dotfiles.repoPath"))
			worktreePath := os.ExpandEnv(viper.GetString("dotfiles.worktree"))

			if repoPath == "" {
				log.Warn("  Repository Path (dotfiles.repoPath): Not Set")
			} else {
				log.Info("  Repository Path (dotfiles.repoPath): %s", repoPath)
			}
			if worktreePath == "" {
				log.Warn("  Worktree Path (dotfiles.worktree): Not Set")
			} else {
				log.Info("  Worktree Path (dotfiles.worktree): %s", worktreePath)
			}
			return // Don't show help if info flag is used
		}

		// If no subcommand is given, print the help information.
		if len(args) == 0 {
			log.Verbose(isVerbose, "No subcommand provided, showing help.")
			err := cmd.Help()
			cobra.CheckErr(err)
		} else {
			log.Verbose(isVerbose, "Subcommand '%s' provided.", args[0])
		}
	},
}

func init() {
	DotfilesCmd.Flags().BoolP("info", "i", false, "Show current dotfiles configuration")

	DotfilesCmd.AddCommand(InstallCmd)
	DotfilesCmd.AddCommand(SyncCmd)
	DotfilesCmd.AddCommand(FetchCmd)
	DotfilesCmd.AddCommand(StatusCmd)
	DotfilesCmd.AddCommand(CopyChangesCmd)
}
