package dotfiles

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestCheckoutCmd_MissingConfig(t *testing.T) {
	// Ensure viper has no config for dotfiles
	viper.Reset()

	// Run command; should return early without panic
	cmd := &cobra.Command{}
	CheckoutCmd.Run(cmd, []string{})
}

func TestCheckoutCmd_SuccessAndFailure(t *testing.T) {
	viper.Reset()
	viper.Set("dotfiles.bare_repo_path", "/tmp/repo")
	viper.Set("dotfiles.worktree_path", "/tmp/worktree")

	// Save original checkoutRepo function to restore it later
	originalCheckoutRepo := checkoutRepo
	defer func() {
		checkoutRepo = originalCheckoutRepo
	}()

	called := 0
	// Override to simulate failure then success
	checkoutRepo = func(repoPath, worktreePath string, force, all bool) error {
		called++
		if called == 1 {
			return errors.New("simulated checkout failure")
		}
		return nil
	}

	cmd := &cobra.Command{}
	// First call => failure path
	CheckoutCmd.Run(cmd, []string{})

	// Second call => success path
	CheckoutCmd.Run(cmd, []string{})

	if called != 2 {
		t.Errorf("Expected checkoutRepo to be called 2 times, got %d", called)
	}
}

func TestCheckoutCmd_Flags(t *testing.T) {
	viper.Reset()
	viper.Set("dotfiles.bare_repo_path", "/tmp/repo")
	viper.Set("dotfiles.worktree_path", "/tmp/worktree")

	// Save original checkoutRepo function to restore it later
	originalCheckoutRepo := checkoutRepo
	defer func() {
		checkoutRepo = originalCheckoutRepo
	}()

	var calledForce, calledAll bool
	checkoutRepo = func(repoPath, worktreePath string, force, all bool) error {
		calledForce = force
		calledAll = all
		return nil
	}

	// Setup a cobra command with flags to test the parsing
	cmd := &cobra.Command{}
	cmd.Flags().BoolP("all", "a", false, "checkout all files from the index/HEAD")
	cmd.Flags().BoolP("force", "f", false, "force checkout, discarding any local changes")

	// Set the flags to true
	err := cmd.Flags().Set("all", "true")
	if err != nil {
		t.Fatalf("Failed to set all flag: %v", err)
	}
	err = cmd.Flags().Set("force", "true")
	if err != nil {
		t.Fatalf("Failed to set force flag: %v", err)
	}

	// Run with flags
	CheckoutCmd.Run(cmd, []string{})

	if !calledForce {
		t.Errorf("Expected force flag to be true")
	}
	if !calledAll {
		t.Errorf("Expected all flag to be true")
	}
}
