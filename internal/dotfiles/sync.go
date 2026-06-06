package dotfiles

import (
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/repo"
)

// SyncRepo performs the fetch and pull-rebase operations for a bare repository.
func SyncRepo(repoPath, worktreePath string, isVerbose bool) error {
	log.Verbose(isVerbose, "Syncing repository at %s with worktree %s", repoPath, worktreePath)

	log.Info("Fetching dotfiles")
	if err := FetchRepo(repoPath, worktreePath); err != nil {
		return err
	}

	log.Info("Pulling dotfiles with rebase")
	if err := PullRebaseRepo(repoPath, worktreePath); err != nil {
		return err
	}

	return nil
}

// pullRebaseRepo is a package-level variable used for testing.
var PullRebaseRepo = func(repoPath, worktreePath string) error {
	return repo.PullRebaseBareRepo(repoPath, worktreePath)
}

var FetchRepo = func(repoPath, worktreePath string) error {
	return repo.FetchBareRepo(repoPath, worktreePath)
}
