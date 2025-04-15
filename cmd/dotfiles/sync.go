package dotfiles

import (
	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/log"
	"github.com/eng618/eng/utils/repo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var SyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "sync your local bear repository",
	Long:  `This command fetches and pulls in remote changes to the local bare dot repository.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Syncing dotfiles")

		isVerbose := utils.IsVerbose(cmd)

		repoPath := viper.GetString("dotfiles.repoPath")
		if repoPath == "" {
			log.Error("dotfiles.repopath is not set in the configuration file")
			return
		}
		log.Verbose(isVerbose, "dotfiles.repoPath: %s", repoPath)

		worktreePath := viper.GetString("dotfiles.worktree")
		if worktreePath == "" {
			log.Error("dotfiles.worktree is not set in the configuration file")
			return
		}
		log.Verbose(isVerbose, "dotfiles.worktree: %s", worktreePath)

		log.Info("Fetching dotfiles")
		err := repo.FetchBareRepo(repoPath, worktreePath)
		if err != nil {
			log.Error("Failed to fetch dotfiles: %s", err)
			return
		}

		// Then pull with rebase
		log.Info("Pulling dotfiles with rebase")
		err = repo.PullRebaseBareRepo(repoPath, worktreePath)
		if err != nil {
			log.Error("Failed to pull and rebase dotfiles: %s", err)
			return
		}

		log.Success("Dotfiles synced successfully")
	},
}
