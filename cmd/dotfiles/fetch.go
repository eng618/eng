package dotfiles

import (
	"os"

	"github.com/eng618/eng/utils/log"
	"github.com/eng618/eng/utils/repo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// FetchCmd defines the cobra command for fetching the dotfiles repository.
// It only fetches remote changes without merging them.
var FetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "fetch your local bare repository",
	Long:  `This command fetches remote changes to the local bare dot repository.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Fetching dotfiles")

		repoPath := os.ExpandEnv(viper.GetString("dotfiles.repoPath"))
		if repoPath == "" {
			log.Error("dotfiles.repoPath is not set or resolves to an empty string in the configuration file")
			return
		}

		worktreePath := os.ExpandEnv(viper.GetString("dotfiles.worktree"))
		if worktreePath == "" {
			log.Error("dotfiles.worktree is not set or resolves to an empty string in the configuration file")
			return
		}

		err := repo.FetchBareRepo(repoPath, worktreePath)
		if err != nil {
			log.Error("Failed to fetch dotfiles: %s", err)
			return
		}

		log.Success("Dotfiles fetched successfully")
	},
}
