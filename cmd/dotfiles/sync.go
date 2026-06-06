package dotfiles

import (
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/dotfiles"
	"github.com/eng618/eng/internal/log"
)

// SyncCmd defines the cobra command for syncing the dotfiles repository.
var SyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "sync your local bare repository",
	Long:  `This command fetches and pulls in remote changes to the local bare dot repository.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Syncing dotfiles")

		isVerbose := cmdutil.IsVerbose(cmd)

		repoPath, worktreePath, err := getDotfilesConfig()
		if err != nil || repoPath == "" {
			log.Error("Dotfiles repository path is not set in configuration")
			return
		}
		log.Verbose(isVerbose, "Repository path: %s", repoPath)
		log.Verbose(isVerbose, "Worktree path:   %s", worktreePath)

		if err = dotfiles.SyncRepo(repoPath, worktreePath, isVerbose); err != nil {
			log.Error("Sync failed: %v", err)
			return
		}

		log.Success("Dotfiles synced successfully")
	},
}
