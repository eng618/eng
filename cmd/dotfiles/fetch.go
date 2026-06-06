package dotfiles

import (
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/dotfiles"
	"github.com/eng618/eng/internal/log"
)

// FetchCmd defines the cobra command for fetching the dotfiles repository.
var FetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "fetch your local bare repository",
	Long:  `This command fetches remote changes to the local bare dot repository.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Fetching dotfiles")

		repoPath, worktreePath, err := getDotfilesConfig()
		if err != nil || repoPath == "" {
			log.Error("Dotfiles repository path is not set in configuration")
			return
		}

		err = dotfiles.FetchRepo(repoPath, worktreePath)
		if err != nil {
			log.Error("Failed to fetch dotfiles: %s", err)
			return
		}

		log.Success("Dotfiles fetched successfully")
	},
}
