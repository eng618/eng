package dotfiles

import (
	"context"
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/dotfiles"
)

func TestSyncCmd_MissingConfig(t *testing.T) {
	viper.Reset()
	cmd := &cobra.Command{}
	SyncCmd.Run(cmd, []string{})
}

func TestSyncCmd_FetchThenPull(t *testing.T) {
	viper.Reset()
	viper.Set("dotfiles.bare_repo_path", "/tmp/repo")
	viper.Set("dotfiles.worktree_path", "/tmp/worktree")

	calls := []string{}

	// Override fetchRepo to record call and simulate success
	originalFetchRepo := dotfiles.FetchRepo
	dotfiles.FetchRepo = func(ctx context.Context, repoPath, worktreePath string) error {
		calls = append(calls, "fetch")
		return nil
	}
	defer func() { dotfiles.FetchRepo = originalFetchRepo }()

	// Override pullRebaseRepo to record call and simulate failure then success
	count := 0
	originalPullRebaseRepo := dotfiles.PullRebaseRepo
	dotfiles.PullRebaseRepo = func(ctx context.Context, repoPath, worktreePath string) error {
		calls = append(calls, "pull")
		count++
		if count == 1 {
			return errors.New("simulated pull failure")
		}
		return nil
	}
	defer func() { dotfiles.PullRebaseRepo = originalPullRebaseRepo }()

	cmd := &cobra.Command{}
	// First run: fetch ok, pull fails
	SyncCmd.Run(cmd, []string{})
	// Second run: fetch ok, pull succeeds
	SyncCmd.Run(cmd, []string{})

	if len(calls) < 4 {
		t.Fatalf("expected at least 4 calls (fetch,pull x2), got %v", calls)
	}
}
