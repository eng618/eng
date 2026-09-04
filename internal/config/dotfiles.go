package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/repo"
	"github.com/eng618/eng/internal/ui/theme"
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
	url, err := InputPrompt(
		"What is your dotfiles repository URL? (e.g., https://github.com/username/dotfiles.git)",
		"",
	)
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
	branch, err := SelectPrompt("Which branch should be used for dotfiles?", []string{"main", "work", "server"}, "main")
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

	path, err := InputPrompt("Where should the bare repository be stored?", defaultPath)
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

	path, err := InputPrompt("What is your worktree path (usually home)?", homeDir)
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

// TargetRepoPath resolves the destination git repository path for copying dotfile changes.
// It checks explicit config first, then project configuration, then scans project directories under devPath,
// and finally falls back to legacy path under devPath.
func TargetRepoPath(devPath string) string {
	if explicitPath := viper.GetString("dotfiles.target_repo_path"); explicitPath != "" {
		return os.ExpandEnv(explicitPath)
	}
	if explicitPath := viper.GetString("dotfiles.local_repo_path"); explicitPath != "" {
		return os.ExpandEnv(explicitPath)
	}

	devPath = os.ExpandEnv(devPath)

	dotfilesCfg := GetDotfilesConfig()
	targetRepoName := "eng-cfg"
	if dotfilesCfg.RepoURL != "" {
		if name, err := repo.RepoNameFromURL(dotfilesCfg.RepoURL); err == nil && name != "" {
			targetRepoName = name
		}
	}

	if devPath == "" {
		return targetRepoName
	}

	// 1. Search configured projects (GetProjects())
	projects := GetProjects()
	for _, p := range projects {
		for _, r := range p.Repos {
			effectiveName, err := r.GetEffectivePath()
			if err == nil && (effectiveName == targetRepoName || r.URL == dotfilesCfg.RepoURL) {
				candidate := filepath.Join(devPath, p.Name, effectiveName)
				if isGitRepoOrDir(candidate) {
					return candidate
				}
			}
		}
	}

	// 2. Search project directories under devPath on disk (devPath/<project_name>/<targetRepoName>)
	if entries, err := os.ReadDir(devPath); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				candidate := filepath.Join(devPath, entry.Name(), targetRepoName)
				if isGitRepoOrDir(candidate) {
					return candidate
				}
			}
		}
	}

	// 3. Check direct child under devPath (legacy path: devPath/<targetRepoName>)
	legacyPath := filepath.Join(devPath, targetRepoName)
	if isGitRepoOrDir(legacyPath) {
		return legacyPath
	}

	// 4. Search up to 3 levels deep under devPath looking for a directory matching targetRepoName containing .git
	var foundPath string
	_ = filepath.WalkDir(devPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			rel, err := filepath.Rel(devPath, path)
			if err == nil {
				if strings.Count(rel, string(filepath.Separator)) > 2 {
					return filepath.SkipDir
				}
			}
			if d.Name() == targetRepoName && isGitRepoOrDir(path) {
				foundPath = path
				return filepath.SkipAll
			}
		}
		return nil
	})
	if foundPath != "" {
		return foundPath
	}

	// 5. Fallback: if configured projects exist, match repo in project config even if not yet on disk
	if len(projects) > 0 {
		for _, p := range projects {
			for _, r := range p.Repos {
				effectiveName, _ := r.GetEffectivePath()
				if effectiveName == targetRepoName || r.URL == dotfilesCfg.RepoURL {
					return filepath.Join(devPath, p.Name, effectiveName)
				}
			}
		}
	}

	return legacyPath
}

// UpdateTargetRepoPath prompts the user to input their target repository path.
func UpdateTargetRepoPath() {
	gitCfg := GetGitConfig()
	defaultPath := TargetRepoPath(gitCfg.DevPath)

	path, err := InputPrompt("Where is your target dotfiles git repository located?", defaultPath)
	cobra.CheckErr(err)

	viper.Set("dotfiles.target_repo_path", path)
	saveConfig()
}

func isGitRepoOrDir(path string) bool {
	gitPath := filepath.Join(path, ".git")
	if _, err := os.Stat(gitPath); err == nil {
		return true
	}
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func saveConfig() {
	if err := viper.WriteConfig(); err != nil {
		cobra.CheckErr(
			fmt.Errorf(
				"%s: %w",
				lipgloss.NewStyle().Foreground(theme.Destructive).Render("Error writing config file"),
				err,
			),
		)
	}
	log.Success("Configuration updated successfully")
}
