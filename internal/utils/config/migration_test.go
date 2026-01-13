package config

import (
	"testing"

	"github.com/spf13/viper"
)

func TestMigrateConfig(t *testing.T) {
	viper.Reset()

	// Set legacy values
	viper.Set("dotfiles.repoPath", "/old/repo")
	viper.Set("dotfiles.worktree", "/old/worktree")
	viper.Set("git.devPath", "/old/dev")

	// Pre-condition check
	if viper.IsSet("dotfiles.bare_repo_path") {
		t.Fatal("dotfiles.bare_repo_path should not be set yet")
	}

	// Run migration
	MigrateConfig()

	// Verify new values
	if val := viper.GetString("dotfiles.bare_repo_path"); val != "/old/repo" {
		t.Errorf("expected bare_repo_path to be /old/repo, got %s", val)
	}
	if val := viper.GetString("dotfiles.worktree_path"); val != "/old/worktree" {
		t.Errorf("expected worktree_path to be /old/worktree, got %s", val)
	}
	if val := viper.GetString("git.dev_path"); val != "/old/dev" {
		t.Errorf("expected dev_path to be /old/dev, got %s", val)
	}

	// Verify worktree fallback (workTree)
	viper.Reset()
	viper.Set("dotfiles.workTree", "/old/workTreeCase")
	MigrateConfig()
	if val := viper.GetString("dotfiles.worktree_path"); val != "/old/workTreeCase" {
		t.Errorf("expected worktree_path to be /old/workTreeCase, got %s", val)
	}
}
