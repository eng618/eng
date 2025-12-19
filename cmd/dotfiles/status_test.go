package dotfiles

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestStatusCmd_MissingConfig(t *testing.T) {
	// Ensure viper has no config for dotfiles
	viper.Reset()

	// Run command; should return early without panic
	cmd := &cobra.Command{}
	StatusCmd.Run(cmd, []string{})
}

func TestStatusCmd_SuccessAndFailure(t *testing.T) {
	viper.Reset()
	viper.Set("dotfiles.bare_repo_path", "/tmp/repo")
	viper.Set("dotfiles.worktree_path", "/tmp/worktree")

	called := 0
	// Override to simulate failure then success
	checkStatus = func(repoPath, worktreePath string) error {
		called++
		if called == 1 {
			return errors.New("simulated status check failure")
		}
		return nil
	}

	cmd := &cobra.Command{}
	// First call => failure path
	StatusCmd.Run(cmd, []string{})
	// Second call => success path
	StatusCmd.Run(cmd, []string{})
}
