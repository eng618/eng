package asdf

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindToolVersionFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Root .tool-versions
	rootFile := filepath.Join(tempDir, ".tool-versions")
	require.NoError(t, os.WriteFile(rootFile, []byte("nodejs 24.18.0\n"), 0o644))

	// Subproject .tool-versions
	projDir := filepath.Join(tempDir, "proj-a")
	require.NoError(t, os.MkdirAll(projDir, 0o755))
	projFile := filepath.Join(projDir, ".tool-versions")
	require.NoError(t, os.WriteFile(projFile, []byte("golang 1.22.3\n"), 0o644))

	// Ignored node_modules directory with .tool-versions (should be skipped)
	nodeModulesDir := filepath.Join(tempDir, "node_modules", "some-pkg")
	require.NoError(t, os.MkdirAll(nodeModulesDir, 0o755))
	ignoredFile := filepath.Join(nodeModulesDir, ".tool-versions")
	require.NoError(t, os.WriteFile(ignoredFile, []byte("python 3.10.0\n"), 0o644))

	files, err := FindToolVersionFiles([]string{tempDir})
	require.NoError(t, err)

	assert.Len(t, files, 2)
	assert.Contains(t, files, rootFile)
	assert.Contains(t, files, projFile)
	assert.NotContains(t, files, ignoredFile)

	// Test progress tracking
	var reportedDirs []string
	var lastCount int
	progressFiles, err := FindToolVersionFilesWithProgress([]string{tempDir}, func(currentDir string, foundCount int) {
		reportedDirs = append(reportedDirs, currentDir)
		lastCount = foundCount
	})
	require.NoError(t, err)
	assert.Equal(t, files, progressFiles)
	assert.NotEmpty(t, reportedDirs)
	assert.Equal(t, 2, lastCount)
}

func TestParseAndMergeToolVersions(t *testing.T) {
	tempDir := t.TempDir()

	f1 := filepath.Join(tempDir, "file1.tool-versions")
	require.NoError(t, os.WriteFile(f1, []byte("nodejs 24.18.0\ngolang 1.26.4\n"), 0o644))

	f2 := filepath.Join(tempDir, "file2.tool-versions")
	require.NoError(t, os.WriteFile(f2, []byte("nodejs 20.19.5\npython 3.12.9\n"), 0o644))

	merged, summaries, err := ParseAndMergeToolVersions([]string{f1, f2})
	require.NoError(t, err)

	assert.Len(t, summaries, 2)
	assert.Equal(t, []string{"24.18.0", "20.19.5"}, merged["nodejs"])
	assert.Equal(t, []string{"1.26.4"}, merged["golang"])
	assert.Equal(t, []string{"3.12.9"}, merged["python"])
}
