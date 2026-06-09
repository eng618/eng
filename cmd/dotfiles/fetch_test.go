package dotfiles

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestFetchCmd_MissingConfig(t *testing.T) {
	// Ensure viper has no config for dotfiles
	viper.Reset()

	// Run command; should return early without panic
	cmd := &cobra.Command{}
	FetchCmd.Run(cmd, []string{})
}

func TestFetchCmd_SuccessAndFailure(t *testing.T) {
	viper.Reset()
	viper.Set("dotfiles.bare_repo_path", "/tmp/repo")
	viper.Set("dotfiles.worktree_path", "/tmp/worktree")

	called := 0
	// Override to simulate failure then success
	fetchRepo = func(repoPath, worktreePath string) error {
		called++
		if called == 1 {
			return errors.New("simulated fetch failure")
		}
		return nil
	}

	cmd := &cobra.Command{}
	// First call => failure path
	FetchCmd.Run(cmd, []string{})
	// Second call => success path
	FetchCmd.Run(cmd, []string{})
}
