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

func TestSetup(t *testing.T) {
	ui.DisableProgress = true

	var buf bytes.Buffer
	log.SetWriters(&buf, &buf)
	defer log.ResetWriters()

	tmpDir := t.TempDir()

	projects := []config.Project{
		{
			Name: "TestProject",
			Repos: []config.ProjectRepo{
				{URL: "git@github.com:org/repo1.git"}, // will mock success
				{URL: "git@github.com:org/repo2.git"}, // will mock fail
				{URL: "git@github.com:org/repo3.git"}, // already cloned
			},
		},
	}

	// repo3 is already cloned
	os.MkdirAll(filepath.Join(tmpDir, "TestProject", "repo3", ".git"), 0o755)

	mockRepoClient := &MockRepoClient{
		CloneFunc: func(ctx context.Context, url, path string) error {
			if url == "git@github.com:org/repo2.git" {
				return errors.New("mock clone failure")
			}
			// Simulate cloning by creating .git
			os.MkdirAll(filepath.Join(path, ".git"), 0o755)
			return nil
		},
	}

	tests := []struct {
		name        string
		opts        SetupOptions
		expectedOut []string
	}{
		{
			name: "Setup DryRun",
			opts: SetupOptions{
				DryRun:     true,
				DevPath:    tmpDir,
				Projects:   projects,
				RepoClient: mockRepoClient,
			},
			expectedOut: []string{
				"Dry run mode",
				"Processing project: TestProject",
				"[DRY RUN] Would clone git@github.com:org/repo1.git",
			},
		},
		{
			name: "Setup Run",
			opts: SetupOptions{
				DevPath:    tmpDir,
				Projects:   projects,
				RepoClient: mockRepoClient,
			},
			expectedOut: []string{
				"Cloned: 1",
				"Failed: 1",
				"Some repositories failed to clone",
			},
		},
		{
			name:        "No projects",
			opts:        SetupOptions{},
			expectedOut: []string{"No projects configured"},
		},
		{
			name: "Nil context and default client",
			opts: SetupOptions{
				DevPath: tmpDir,
				Projects: []config.Project{
					{Name: "NilContext", Repos: []config.ProjectRepo{}},
				},
				RepoClient: nil,
			},
			expectedOut: []string{"Setup complete"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			Setup(context.Background(), tt.opts)

			out := buf.String()
			for _, exp := range tt.expectedOut {
				assert.Contains(t, out, exp)
			}
		})
	}
}
