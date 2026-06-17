package project

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
)

func TestPull(t *testing.T) {
	ui.DisableProgress = true

	var buf bytes.Buffer
	log.SetWriters(&buf, &buf)
	defer log.ResetWriters()

	tmpDir := t.TempDir()

	projects := []config.Project{
		{
			Name: "TestProject",
			Repos: []config.ProjectRepo{
				{URL: "git@github.com:org/repo1.git"}, // pull success
				{URL: "git@github.com:org/repo2.git"}, // dirty
				{URL: "git@github.com:org/repo3.git"}, // pull failure
				{URL: "git@github.com:org/repo4.git"}, // up to date
				{URL: "git@github.com:org/repo5.git"}, // not cloned
			},
		},
	}

	// Create dummy repos for 1-4
	for i := 1; i <= 4; i++ {
		os.MkdirAll(filepath.Join(tmpDir, "TestProject", "repo"+string(rune('0'+i)), ".git"), 0o755)
	}

	mockRepoClient := &MockRepoClient{
		IsDirtyFunc: func(ctx context.Context, path string) (bool, error) {
			if filepath.Base(path) == "repo2" {
				return true, nil // dirty
			}
			return false, nil
		},
		PullLatestCodeFunc: func(ctx context.Context, path string) error {
			if filepath.Base(path) == "repo3" {
				return errors.New("mock pull failure")
			}
			if filepath.Base(path) == "repo4" {
				return git.NoErrAlreadyUpToDate
			}
			return nil
		},
	}

	tests := []struct {
		name        string
		opts        PullOptions
		expectedOut []string
	}{
		{
			name: "Pull DryRun",
			opts: PullOptions{
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
			name: "Pull Run",
			opts: PullOptions{
				DevPath:    tmpDir,
				Projects:   projects,
				RepoClient: mockRepoClient,
			},
			expectedOut: []string{
				"Pull complete: 2 successful, 1 skipped, 1 dirty, 1 failed",
			},
		},
		{
			name:        "No projects",
			opts:        PullOptions{},
			expectedOut: []string{"No projects configured"},
		},
		{
			name: "Nil context and default client",
			opts: PullOptions{
				DevPath: tmpDir,
				Projects: []config.Project{
					{Name: "NilContext", Repos: []config.ProjectRepo{}},
				},
				RepoClient: nil,
			},
			expectedOut: []string{"Pull complete"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			Pull(context.Background(), tt.opts)

			out := buf.String()
			for _, exp := range tt.expectedOut {
				assert.Contains(t, out, exp)
			}
		})
	}
}
