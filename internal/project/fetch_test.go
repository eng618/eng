package project

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
)

func TestFetch(t *testing.T) {
	ui.DisableProgress = true

	var buf bytes.Buffer
	log.SetWriters(&buf, &buf)
	defer log.ResetWriters()

	tmpDir := t.TempDir()

	projects := []config.Project{
		{
			Name: "TestProject",
			Repos: []config.ProjectRepo{
				{URL: "git@github.com:org/repo1.git"}, // fetch success
				{URL: "git@github.com:org/repo2.git"}, // fetch failure
				{URL: "git@github.com:org/repo3.git"}, // not cloned
			},
		},
	}

	// Create dummy repos
	os.MkdirAll(filepath.Join(tmpDir, "TestProject", "repo1", ".git"), 0o755)
	os.MkdirAll(filepath.Join(tmpDir, "TestProject", "repo2", ".git"), 0o755)

	mockRepoClient := &MockRepoClient{
		FetchAllPruneFunc: func(ctx context.Context, path string) error {
			if filepath.Base(path) == "repo1" {
				return nil
			}
			return errors.New("mock fetch failure")
		},
	}

	tests := []struct {
		name        string
		opts        FetchOptions
		expectedOut []string
	}{
		{
			name: "Fetch DryRun",
			opts: FetchOptions{
				DryRun:     true,
				DevPath:    tmpDir,
				Projects:   projects,
				RepoClient: mockRepoClient,
			},
			expectedOut: []string{
				"Dry run mode",
			},
		},
		{
			name: "Fetch Run",
			opts: FetchOptions{
				DevPath:    tmpDir,
				Projects:   projects,
				RepoClient: mockRepoClient,
			},
			expectedOut: []string{
				"Fetch complete: 1 successful, 1 skipped, 1 failed",
			},
		},
		{
			name:        "No projects",
			opts:        FetchOptions{},
			expectedOut: []string{"No projects configured"},
		},
		{
			name: "Project filter not found",
			opts: FetchOptions{
				Projects:      projects,
				ProjectFilter: "NonExistent",
			},
			expectedOut: []string{}, // should just return
		},
		{
			name: "Verbose mode skipping",
			opts: FetchOptions{
				IsVerbose:  true,
				DevPath:    tmpDir,
				Projects:   projects,
				RepoClient: mockRepoClient,
			},
			expectedOut: []string{},
		},
		{
			name: "Nil context and default client",
			opts: FetchOptions{
				DevPath: tmpDir,
				Projects: []config.Project{
					{Name: "NilContext", Repos: []config.ProjectRepo{}},
				},
				RepoClient: nil,
			},
			expectedOut: []string{"Fetch complete"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			Fetch(context.Background(), tt.opts)

			out := buf.String()
			for _, exp := range tt.expectedOut {
				assert.Contains(t, out, exp)
			}
		})
	}
}

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
