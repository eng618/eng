package config

import (
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/utils/log"
)

// MigrateConfig handles the migration of old configuration keys to new standardized ones.
// It checks for old keys, moves their values to new keys if the new ones are empty,
// and saves the configuration if any changes were made.
func MigrateConfig() {
	isVerbose := viper.GetBool("verbose")
	log.Verbose(isVerbose, "Checking for configuration migration...")
	changed := false

	// Migration: dotfiles.repoPath -> dotfiles.bare_repo_path
	if viper.IsSet("dotfiles.repoPath") && !viper.IsSet("dotfiles.bare_repo_path") {
		oldVal := viper.Get("dotfiles.repoPath")
		viper.Set("dotfiles.bare_repo_path", oldVal)
		log.Verbose(isVerbose, "Migrated dotfiles.repoPath to dotfiles.bare_repo_path: %v", oldVal)
		changed = true
	}

	// Migration: dotfiles.worktree or dotfiles.workTree -> dotfiles.worktree_path
	if !viper.IsSet("dotfiles.worktree_path") {
		if viper.IsSet("dotfiles.worktree") {
			oldVal := viper.Get("dotfiles.worktree")
			viper.Set("dotfiles.worktree_path", oldVal)
			log.Verbose(isVerbose, "Migrated dotfiles.worktree to dotfiles.worktree_path: %v", oldVal)
			changed = true
		} else if viper.IsSet("dotfiles.workTree") {
			oldVal := viper.Get("dotfiles.workTree")
			viper.Set("dotfiles.worktree_path", oldVal)
			log.Verbose(isVerbose, "Migrated dotfiles.workTree to dotfiles.worktree_path: %v", oldVal)
			changed = true
		}
	}

	// Migration: git.devPath -> git.dev_path
	if viper.IsSet("git.devPath") && !viper.IsSet("git.dev_path") {
		oldVal := viper.Get("git.devPath")
		viper.Set("git.dev_path", oldVal)
		log.Verbose(isVerbose, "Migrated git.devPath to git.dev_path: %v", oldVal)
		changed = true
	}

	if changed {
		if err := viper.WriteConfig(); err != nil {
			log.Warn("Failed to save migrated configuration: %v", err)
		} else {
			log.Success("Configuration migrated to latest version")
		}
	} else {
		log.Verbose(isVerbose, "Configuration is up to date, no migration needed")
	}
}
