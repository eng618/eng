package project

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
)

func TestSync(t *testing.T) {
	ui.DisableProgress = true

	var buf bytes.Buffer
	log.SetWriters(&buf, &buf)
	defer log.ResetWriters()

	tmpDir := t.TempDir()

	projects := []config.Project{
		{
			Name: "TestProject",
			Repos: []config.ProjectRepo{
				{URL: "git@github.com:org/repo1.git"}, // clone success
				{URL: "git@github.com:org/repo2.git"}, // clone failure
				{URL: "git@github.com:org/repo3.git"}, // dirty
				{URL: "git@github.com:org/repo4.git"}, // pull success
				{URL: "git@github.com:org/repo5.git"}, // pull failure
			},
		},
	}

	os.MkdirAll(filepath.Join(tmpDir, "TestProject", "repo3", ".git"), 0o755)
	os.MkdirAll(filepath.Join(tmpDir, "TestProject", "repo4", ".git"), 0o755)
	os.MkdirAll(filepath.Join(tmpDir, "TestProject", "repo5", ".git"), 0o755)

	mockRepoClient := &MockRepoClient{
		CloneFunc: func(ctx context.Context, url, path string) error {
			if url == "git@github.com:org/repo2.git" {
				return errors.New("mock clone failure")
			}
			os.MkdirAll(filepath.Join(path, ".git"), 0o755)
			return nil
		},
		IsDirtyFunc: func(ctx context.Context, path string) (bool, error) {
			if filepath.Base(path) == "repo3" {
				return true, nil
			}
			return false, nil
		},
		PullLatestCodeFunc: func(ctx context.Context, path string) error {
			if filepath.Base(path) == "repo5" {
				return errors.New("mock pull failure")
			}
			return nil
		},
	}

	tests := []struct {
		name        string
		opts        SyncOptions
		expectedOut []string
	}{
		{
			name: "Sync DryRun",
			opts: SyncOptions{
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
			name: "Sync Run",
			opts: SyncOptions{
				DevPath:    tmpDir,
				Projects:   projects,
				RepoClient: mockRepoClient,
			},
			expectedOut: []string{
				"Sync complete",
			},
		},
		{
			name:        "No projects",
			opts:        SyncOptions{},
			expectedOut: []string{"No projects configured"},
		},
		{
			name: "Nil context and default client",
			opts: SyncOptions{
				DevPath: tmpDir,
				Projects: []config.Project{
					{Name: "NilContext", Repos: []config.ProjectRepo{}},
				},
				RepoClient: nil,
			},
			expectedOut: []string{"Sync complete"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			Sync(context.Background(), tt.opts)

			out := buf.String()
			for _, exp := range tt.expectedOut {
				assert.Contains(t, out, exp)
			}
		})
	}
}
