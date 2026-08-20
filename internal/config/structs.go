package config

import "github.com/spf13/viper"

// GitConfig holds all git-related configuration.
type GitConfig struct {
	DevPath string `mapstructure:"dev_path"`
	Editor  string `mapstructure:"editor"`
}

// GetGitConfig retrieves the git configuration from Viper.
func GetGitConfig() GitConfig {
	return GitConfig{
		DevPath: viper.GetString("git.dev_path"),
		Editor:  viper.GetString("git.editor"),
	}
}

// DotfilesConfig holds all dotfiles-related configuration.
type DotfilesConfig struct {
	RepoURL        string `mapstructure:"repo_url"`
	Branch         string `mapstructure:"branch"`
	BareRepoPath   string `mapstructure:"bare_repo_path"`
	WorktreePath   string `mapstructure:"worktree_path"`
	TargetRepoPath string `mapstructure:"target_repo_path"`
}

// GetDotfilesConfig retrieves the dotfiles configuration from Viper.
func GetDotfilesConfig() DotfilesConfig {
	return DotfilesConfig{
		RepoURL:        viper.GetString("dotfiles.repo_url"),
		Branch:         viper.GetString("dotfiles.branch"),
		BareRepoPath:   viper.GetString("dotfiles.bare_repo_path"),
		WorktreePath:   viper.GetString("dotfiles.worktree_path"),
		TargetRepoPath: viper.GetString("dotfiles.target_repo_path"),
	}
}

// ContainersConfig holds container-related configuration.
type ContainersConfig struct {
	Path string `mapstructure:"path"`
}

// GetContainersConfig retrieves the containers configuration from Viper.
func GetContainersConfig() ContainersConfig {
	return ContainersConfig{
		Path: viper.GetString("containers.path"),
	}
}

// AntigravityConfig holds Antigravity-related configuration.
type AntigravityConfig struct {
	IdeDownloadURL string `mapstructure:"ide_download_url"`
}

// GetAntigravityConfig retrieves the Antigravity configuration from Viper.
func GetAntigravityConfig() AntigravityConfig {
	return AntigravityConfig{
		IdeDownloadURL: viper.GetString("antigravity.ide_download_url"),
	}
}
