package dotfiles

import (
	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/log"
	"github.com/eng618/eng/utils/repo"
	"github.com/spf13/cobra"
)

// SyncCmd defines the cobra command for syncing the dotfiles repository.
// It fetches remote changes and then performs a pull with rebase.
var SyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "sync your local bare repository",
	Long:  `This command fetches and pulls in remote changes to the local bare dot repository.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Syncing dotfiles")

		isVerbose := utils.IsVerbose(cmd)

		repoPath, worktreePath, err := getDotfilesConfig()
		if err != nil || repoPath == "" {
			log.Error("Dotfiles repository path is not set in configuration")
			return
		}
		log.Verbose(isVerbose, "Repository path: %s", repoPath)
		log.Verbose(isVerbose, "Worktree path:   %s", worktreePath)

		if err = SyncRepo(repoPath, worktreePath, isVerbose); err != nil {
			log.Error("Sync failed: %v", err)
			return
		}

		log.Success("Dotfiles synced successfully")
	},
}

// SyncRepo performs the fetch and pull-rebase operations for a bare repository.
func SyncRepo(repoPath, worktreePath string, isVerbose bool) error {
	log.Verbose(isVerbose, "Syncing repository at %s with worktree %s", repoPath, worktreePath)

	log.Info("Fetching dotfiles")
	if err := repo.FetchBareRepo(repoPath, worktreePath); err != nil {
		return err
	}

	log.Info("Pulling dotfiles with rebase")
	if err := repo.PullRebaseBareRepo(repoPath, worktreePath); err != nil {
		return err
	}

	return nil
}

// pullRebaseRepo is a package-level variable used for testing.
var pullRebaseRepo = func(repoPath, worktreePath string) error {
	return repo.PullRebaseBareRepo(repoPath, worktreePath)
}
