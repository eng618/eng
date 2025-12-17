package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestRepoURL(t *testing.T) {
	// Reset viper for test
	viper.Reset()
	
	// Set a test value
	viper.Set("dotfiles.repo_url", "https://github.com/test/dotfiles.git")
	
	// Get the value
	url := viper.GetString("dotfiles.repo_url")
	
	assert.Equal(t, "https://github.com/test/dotfiles.git", url)
}

func TestBranch(t *testing.T) {
	// Reset viper for test
	viper.Reset()
	
	// Set a test value
	viper.Set("dotfiles.branch", "main")
	
	// Get the value
	branch := viper.GetString("dotfiles.branch")
	
	assert.Equal(t, "main", branch)
}

func TestBareRepoPath(t *testing.T) {
	// Reset viper for test
	viper.Reset()
	
	// Set a test value
	viper.Set("dotfiles.bare_repo_path", "$HOME/.eng-cfg")
	
	// Get the value
	path := viper.GetString("dotfiles.bare_repo_path")
	
	assert.Equal(t, "$HOME/.eng-cfg", path)
}
