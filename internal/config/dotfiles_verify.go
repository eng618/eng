package config

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
)

// VerifyDotfilesConfig checks for Repo URL, Branch, and Bare Repo Path.
// If all are present, it offers a single multi-select to update them.
// If any are missing, it falls back to sequential mandatory prompts.
func VerifyDotfilesConfig() (string, string, string, string, error) {
	repoURL := viper.GetString("dotfiles.repo_url")
	branch := viper.GetString("dotfiles.branch")
	bareRepoPath := viper.GetString("dotfiles.bare_repo_path")
	worktreePath := viper.GetString("dotfiles.worktree_path")

	// If any are missing, fall back to sequential (which handle missing)
	if repoURL == "" || branch == "" || bareRepoPath == "" || worktreePath == "" {
		repoURL = RepoURL()
		branch = Branch()
		bareRepoPath = BareRepoPath()
		worktreePath = WorktreePath()
		return repoURL, branch, bareRepoPath, worktreePath, nil
	}

	// All are present, offer multi-select
	bareRepoPath = os.ExpandEnv(bareRepoPath)
	worktreePath = os.ExpandEnv(worktreePath)

	options := []string{
		fmt.Sprintf("Repo URL: %s", color.CyanString(repoURL)),
		fmt.Sprintf("Branch:   %s", color.CyanString(branch)),
		fmt.Sprintf("Bare Path:%s", color.CyanString(bareRepoPath)),
		fmt.Sprintf("Worktree: %s", color.CyanString(worktreePath)),
	}

	selected, err := ui.MultiSelect(
		"Which values would you like to update? (Select none if all are correct)",
		options,
		nil,
	)
	if err != nil {
		return "", "", "", "", fmt.Errorf("selection failed: %w", err)
	}

	updateRepo := false
	updateBranch := false
	updatePath := false
	updateWorktree := false

	for _, s := range selected {
		switch s {
		case options[0]:
			updateRepo = true
		case options[1]:
			updateBranch = true
		case options[2]:
			updatePath = true
		case options[3]:
			updateWorktree = true
		}
	}

	if updateRepo {
		UpdateRepoURL()
		repoURL = viper.GetString("dotfiles.repo_url")
	}
	if updateBranch {
		UpdateBranch()
		branch = viper.GetString("dotfiles.branch")
	}
	if updatePath {
		UpdateBareRepoPath()
		bareRepoPath = viper.GetString("dotfiles.bare_repo_path")
		bareRepoPath = os.ExpandEnv(bareRepoPath)
	}
	if updateWorktree {
		UpdateWorktreePath()
		worktreePath = viper.GetString("dotfiles.worktree_path")
		worktreePath = os.ExpandEnv(worktreePath)
	}

	log.Success("Dotfiles configuration verified")
	return repoURL, branch, bareRepoPath, worktreePath, nil
}
