package asdf

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilterRemovableVersions(t *testing.T) {
	installed := map[string][]string{
		"nodejs": {"20.19.5", "22.14.0", "24.18.0"},
		"golang": {"1.22.3", "1.26.4"},
		"python": {"3.12.9"},
	}

	protected := ToolVersions{
		"nodejs": {"24.18.0"},
		"golang": {"1.26.4"},
		"python": {"3.12.9"},
	}

	removable := FilterRemovableVersions(installed, protected, "", "")
	assert.Len(t, removable, 3)

	expectedPlugins := []string{"nodejs", "nodejs", "golang"}
	for _, target := range removable {
		assert.Contains(t, expectedPlugins, target.Plugin)
	}

	// Test with targetPlugin filter
	nodeRemovable := FilterRemovableVersions(installed, protected, "nodejs", "")
	assert.Len(t, nodeRemovable, 2)
}

func TestCalculateDirSize(t *testing.T) {
	tempDir := t.TempDir()
	f1 := filepath.Join(tempDir, "file1.txt")
	require.NoError(t, os.WriteFile(f1, []byte("1234567890"), 0o644)) // 10 bytes

	subDir := filepath.Join(tempDir, "subdir")
	require.NoError(t, os.MkdirAll(subDir, 0o755))
	f2 := filepath.Join(subDir, "file2.txt")
	require.NoError(t, os.WriteFile(f2, []byte("12345"), 0o644)) // 5 bytes

	size := CalculateDirSize(tempDir)
	assert.Equal(t, int64(15), size)
}
