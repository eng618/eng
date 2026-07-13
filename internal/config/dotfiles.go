package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
)

// RepoURL checks for the dotfiles repository URL in the configuration and returns it.
func RepoURL() string {
	repoURL := viper.GetString("dotfiles.repo_url")
	if repoURL == "" {
		UpdateRepoURL()
		repoURL = viper.GetString("dotfiles.repo_url")
	}
	return repoURL
}

// UpdateRepoURL prompts the user to input their dotfiles repository URL.
func UpdateRepoURL() {
	url, err := ui.Input("What is your dotfiles repository URL? (e.g., https://github.com/username/dotfiles.git)", "")
	cobra.CheckErr(err)

	viper.Set("dotfiles.repo_url", url)
	saveConfig()
}

// Branch checks for the dotfiles branch in the configuration and returns it.
func Branch() string {
	branch := viper.GetString("dotfiles.branch")
	if branch == "" {
		UpdateBranch()
		branch = viper.GetString("dotfiles.branch")
	}
	return branch
}

// UpdateBranch prompts the user to select their dotfiles branch.
func UpdateBranch() {
	branch, err := ui.Select("Which branch should be used for dotfiles?", []string{"main", "work", "server"}, "main")
	cobra.CheckErr(err)

	viper.Set("dotfiles.branch", branch)
	saveConfig()
}

// BareRepoPath checks for the bare repository path in the configuration and returns it.
func BareRepoPath() string {
	bareRepoPath := viper.GetString("dotfiles.bare_repo_path")
	if bareRepoPath == "" {
		UpdateBareRepoPath()
		bareRepoPath = viper.GetString("dotfiles.bare_repo_path")
	}
	return os.ExpandEnv(bareRepoPath)
}

// UpdateBareRepoPath prompts the user to input their bare repository path.
func UpdateBareRepoPath() {
	homeDir, err := os.UserHomeDir()
	cobra.CheckErr(err)

	defaultPath := filepath.Join(homeDir, ".eng-cfg")

	path, err := ui.Input("Where should the bare repository be stored?", defaultPath)
	cobra.CheckErr(err)

	viper.Set("dotfiles.bare_repo_path", path)
	saveConfig()
}

// WorktreePath checks for the worktree path in the configuration and returns it.
func WorktreePath() string {
	worktreePath := viper.GetString("dotfiles.worktree_path")
	if worktreePath == "" {
		UpdateWorktreePath()
		worktreePath = viper.GetString("dotfiles.worktree_path")
	}
	return os.ExpandEnv(worktreePath)
}

// UpdateWorktreePath prompts the user to input their worktree path.
func UpdateWorktreePath() {
	homeDir, err := os.UserHomeDir()
	cobra.CheckErr(err)

	path, err := ui.Input("What is your worktree path (usually home)?", homeDir)
	cobra.CheckErr(err)

	viper.Set("dotfiles.worktree_path", path)
	saveConfig()
}

// DotfilesRepo is an alias for BareRepoPath for backward compatibility with external callers.
func DotfilesRepo() string {
	return BareRepoPath()
}

// UpdateDotfilesRepo is an alias for UpdateBareRepoPath.
func UpdateDotfilesRepo() {
	UpdateBareRepoPath()
}

// GetDotfilesRepo logs the current dotfiles repository path.
func GetDotfilesRepo() {
	path := BareRepoPath()
	log.Success("Dotfiles repository path: %s", path)
}

func saveConfig() {
	if err := viper.WriteConfig(); err != nil {
		err := errors.New(color.RedString("Error writing config file: %v", err))
		cobra.CheckErr(err)
	}
	log.Success("Configuration updated successfully")
}
