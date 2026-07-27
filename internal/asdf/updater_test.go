package asdf

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckMissingRequirements(t *testing.T) {
	summaries := []FileProtectionSummary{
		{
			FilePath: "/home/user/.tool-versions",
			Protected: ToolVersions{
				"nodejs": {"24.18.0"},
				"golang": {"1.26.4"},
			},
		},
		{
			FilePath: "/home/user/Development/legacy/.tool-versions",
			Protected: ToolVersions{
				"nodejs": {"20.19.5"},
			},
		},
	}

	installed := map[string][]string{
		"nodejs": {"24.18.0"},
		"golang": {"1.26.4"},
	}

	missing := CheckMissingRequirements(summaries, installed)
	require.Len(t, missing, 1)

	assert.Equal(t, "nodejs", missing[0].Plugin)
	assert.Equal(t, "20.19.5", missing[0].Version)
	assert.Equal(t, "/home/user/Development/legacy/.tool-versions", missing[0].SourceFile)
}

func TestWriteToolVersions(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, ".tool-versions")

	tv := ToolVersions{
		"nodejs": {"26.5.0"},
		"golang": {"1.26.5"},
	}

	err := WriteToolVersions(filePath, tv)
	require.NoError(t, err)

	readBack, err := ParseToolVersionsFile(filePath)
	require.NoError(t, err)

	assert.Equal(t, []string{"26.5.0"}, readBack["nodejs"])
	assert.Equal(t, []string{"1.26.5"}, readBack["golang"])
}
