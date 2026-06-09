package dotfiles

import (
	"context"
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/dotfiles"
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
	originalFetchRepo := dotfiles.FetchRepo
	dotfiles.FetchRepo = func(ctx context.Context, repo, worktree string) error {
		called++
		if called == 1 {
			return errors.New("simulated fetch failure")
		}
		return nil
	}
	defer func() { dotfiles.FetchRepo = originalFetchRepo }()

	cmd := &cobra.Command{}
	// First call => failure path
	FetchCmd.Run(cmd, []string{})
	// Second call => success path
	FetchCmd.Run(cmd, []string{})
}
