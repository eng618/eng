package project

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsRepoCloned(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-is-cloned-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Test non-existent path
	assert.False(t, isRepoCloned(filepath.Join(tmpDir, "nonexistent")))

	// Test directory without .git
	noGitDir := filepath.Join(tmpDir, "no-git")
	err = os.MkdirAll(noGitDir, 0o755)
	require.NoError(t, err)
	assert.False(t, isRepoCloned(noGitDir))

	// Test directory with .git file (not directory)
	gitFileDir := filepath.Join(tmpDir, "git-file")
	err = os.MkdirAll(gitFileDir, 0o755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(gitFileDir, ".git"), []byte("gitdir: ../other"), 0o644)
	require.NoError(t, err)
	assert.False(t, isRepoCloned(gitFileDir)) // .git must be a directory

	// Test directory with .git directory
	gitDirPath := filepath.Join(tmpDir, "has-git")
	err = os.MkdirAll(filepath.Join(gitDirPath, ".git"), 0o755)
	require.NoError(t, err)
	assert.True(t, isRepoCloned(gitDirPath))
}
