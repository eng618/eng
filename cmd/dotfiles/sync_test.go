package dotfiles

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	fetchRepo = func(repoPath, worktreePath string) error {
		calls = append(calls, "fetch")
		return nil
	}

	// Override pullRebaseRepo to record call and simulate failure then success
	count := 0
	pullRebaseRepo = func(repoPath, worktreePath string) error {
		calls = append(calls, "pull")
		count++
		if count == 1 {
			return errors.New("simulated pull failure")
		}
		return nil
	}

	cmd := &cobra.Command{}
	// First run: fetch ok, pull fails
	SyncCmd.Run(cmd, []string{})
	// Second run: fetch ok, pull succeeds
	SyncCmd.Run(cmd, []string{})

	if len(calls) < 4 {
		t.Fatalf("expected at least 4 calls (fetch,pull x2), got %v", calls)
	}
}
