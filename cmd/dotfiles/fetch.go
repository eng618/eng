package dotfiles

import (
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/utils/log"
	"github.com/eng618/eng/internal/utils/repo"
)

// FetchCmd defines the cobra command for fetching the dotfiles repository.
// It only fetches remote changes without merging them.
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

		// Use an injectable function so tests can replace it and avoid running real git.
		err = fetchRepo(repoPath, worktreePath)
		if err != nil {
			log.Error("Failed to fetch dotfiles: %s", err)
			return
		}

		log.Success("Dotfiles fetched successfully")
	},
}

// fetchRepo is a package-level variable so tests can override the implementation.
// By default it calls repo.FetchBareRepo.
var fetchRepo = func(repoPath, worktreePath string) error {
	return repo.FetchBareRepo(repoPath, worktreePath)
}
