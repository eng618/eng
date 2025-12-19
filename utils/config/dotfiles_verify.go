package config

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/viper"

	"github.com/eng618/eng/utils/log"
)

// VerifyDotfilesConfig checks for Repo URL, Branch, and Bare Repo Path.
// If all are present, it offers a single multi-select to update them.
// If any are missing, it falls back to sequential mandatory prompts.
func VerifyDotfilesConfig() (string, string, string, string) {
	repoURL := viper.GetString("dotfiles.repo_url")
	branch := viper.GetString("dotfiles.branch")
	bareRepoPath := viper.GetString("dotfiles.bare_repo_path")
	if bareRepoPath == "" {
		bareRepoPath = viper.GetString("dotfiles.repoPath")
	}
	worktreePath := viper.GetString("dotfiles.workTree")
	if worktreePath == "" {
		worktreePath = viper.GetString("dotfiles.worktree")
	}

	// If any are missing, fall back to sequential (which handle missing)
	if repoURL == "" || branch == "" || bareRepoPath == "" || worktreePath == "" {
		repoURL = RepoURL()
		branch = Branch()
		bareRepoPath = BareRepoPath()
		if worktreePath == "" {
			homeDir, _ := os.UserHomeDir()
			worktreePath = homeDir
			viper.Set("dotfiles.worktree", worktreePath)
			viper.WriteConfig()
		}
		return repoURL, branch, bareRepoPath, worktreePath
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

	var selected []string
	prompt := &survey.MultiSelect{
		Message: "Which values would you like to update? (Select none if all are correct)",
		Options: options,
	}

	if err := survey.AskOne(prompt, &selected); err != nil {
		log.Fatal("Selection failed: %v", err)
	}

	updateRepo := false
	updateBranch := false
	updatePath := false
	updateWorktree := false

	for _, s := range selected {
		if s == options[0] {
			updateRepo = true
		} else if s == options[1] {
			updateBranch = true
		} else if s == options[2] {
			updatePath = true
		} else if s == options[3] {
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
		var wt string
		prompt := &survey.Input{
			Message: "What is your worktree path (usually home)?",
			Default: worktreePath,
		}
		survey.AskOne(prompt, &wt)
		viper.Set("dotfiles.worktree", wt)
		viper.WriteConfig()
		worktreePath = os.ExpandEnv(wt)
	}

	log.Success("Dotfiles configuration verified")
	return repoURL, branch, bareRepoPath, worktreePath
}
