package dotfiles

import (
	"os"

	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/log"
	"github.com/eng618/eng/utils/repo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

		repoPath := viper.GetString("dotfiles.repoPath")
		repoPath = os.ExpandEnv(repoPath) // Expand environment variables
		if repoPath == "" {
			log.Error("dotfiles.repoPath is not set or resolves to an empty string in the configuration file")
			return
		}
		log.Verbose(isVerbose, "dotfiles.repoPath: %s", repoPath)

		worktreePath := viper.GetString("dotfiles.worktree")
		worktreePath = os.ExpandEnv(worktreePath) // Expand environment variables
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
