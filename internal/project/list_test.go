package project

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
)

func TestList(t *testing.T) {
	// Setup test environment with custom logger
	var buf bytes.Buffer
	log.SetWriters(&buf, &buf)
	defer log.ResetWriters()

	tmpDir := t.TempDir()

	// Create dummy project paths
	projectA := filepath.Join(tmpDir, "ProjectA")
	repo1Path := filepath.Join(projectA, "repo1", ".git")
	err := os.MkdirAll(repo1Path, 0o755)
	require.NoError(t, err)

	projects := []config.Project{
		{
			Name: "ProjectA",
			Repos: []config.ProjectRepo{
				{URL: "git@github.com:org/repo1.git"}, // Cloned
				{URL: "git@github.com:org/repo2.git"}, // Not Cloned
			},
		},
		{
			Name: "ProjectB",
			Repos: []config.ProjectRepo{
				{URL: "git@github.com:org/repo3.git"}, // Not Cloned
			},
		},
	}

	tests := []struct {
		name          string
		opts          ListOptions
		expectedOut   []string
		unexpectedOut []string
	}{
		{
			name: "List All Projects",
			opts: ListOptions{
				DevPath:  tmpDir,
				Projects: projects,
			},
			expectedOut: []string{
				"○ ProjectA (1/2 repos cloned)",
				"○ ProjectB (0/1 repos cloned)",
			},
		},
		{
			name: "List Verbose",
			opts: ListOptions{
				IsVerbose: true,
				DevPath:   tmpDir,
				Projects:  projects,
			},
			expectedOut: []string{
				"Project: ProjectA",
				"✓ repo1",
				"✗ repo2 (not cloned)",
				"Project: ProjectB",
				"✗ repo3 (not cloned)",
			},
		},
		{
			name: "List Filtered",
			opts: ListOptions{
				ProjectFilter: "ProjectB",
				DevPath:       tmpDir,
				Projects:      projects,
			},
			expectedOut: []string{
				"○ ProjectB (0/1 repos cloned)",
			},
			unexpectedOut: []string{
				"ProjectA",
			},
		},
		{
			name: "No Dev Path",
			opts: ListOptions{
				DevPath:  "",
				Projects: projects,
			},
			expectedOut: []string{
				"Development folder path is not set",
			},
		},
		{
			name: "No Projects",
			opts: ListOptions{
				DevPath:  tmpDir,
				Projects: []config.Project{},
			},
			expectedOut: []string{
				"No projects configured",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			List(tt.opts)

			out := buf.String()
			for _, exp := range tt.expectedOut {
				assert.Contains(t, out, exp)
			}
			for _, unexp := range tt.unexpectedOut {
				assert.NotContains(t, out, unexp)
			}
		})
	}
}
