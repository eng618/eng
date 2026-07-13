package dotfiles

import (
	"context"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/repo"
)

// SyncRepo performs the fetch and pull-rebase operations for a bare repository.
func SyncRepo(ctx context.Context, repoPath, worktreePath string, isVerbose bool) error {
	log.Verbose(isVerbose, "Syncing repository at %s with worktree %s", repoPath, worktreePath)

	log.Info("Fetching dotfiles")
	if err := FetchRepo(ctx, repoPath, worktreePath); err != nil {
		return err
	}

	log.Info("Pulling dotfiles with rebase")
	if err := PullRebaseRepo(ctx, repoPath, worktreePath); err != nil {
		return err
	}

	return nil
}

// pullRebaseRepo is a package-level variable used for testing.
var PullRebaseRepo = func(ctx context.Context, repoPath, worktreePath string) error {
	return repo.PullRebaseBareRepo(ctx, repoPath, worktreePath)
}

var FetchRepo = func(ctx context.Context, repoPath, worktreePath string) error {
	return repo.FetchBareRepo(ctx, repoPath, worktreePath)
}
