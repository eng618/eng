package asdf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseToolVersions(t *testing.T) {
	content := `
# Comment line
nodejs 24.18.0
golang 1.26.4 1.25.0
python 3.12.9 # inline comment ignored as field if split correctly
# another comment
`
	result, err := ParseToolVersions(strings.NewReader(content))
	require.NoError(t, err)

	assert.Equal(t, []string{"24.18.0"}, result["nodejs"])
	assert.Equal(t, []string{"1.26.4", "1.25.0"}, result["golang"])
	assert.Contains(t, result["python"], "3.12.9")
}

func TestParseToolVersionsFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, ".tool-versions")
	content := "ruby 3.4.8\n"

	err := os.WriteFile(filePath, []byte(content), 0o644)
	require.NoError(t, err)

	result, err := ParseToolVersionsFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, []string{"3.4.8"}, result["ruby"])
}

func TestParseASDFListOutput(t *testing.T) {
	output := `bun
  1.3.10
 *1.3.14
golang
  1.22.3
  1.25.0
 *1.26.3
  1.26.4
nodejs
 *24.18.0
`
	installed, err := ParseASDFListOutput(output)
	require.NoError(t, err)

	assert.Equal(t, []string{"1.3.10", "1.3.14"}, installed["bun"])
	assert.Equal(t, []string{"1.22.3", "1.25.0", "1.26.3", "1.26.4"}, installed["golang"])
	assert.Equal(t, []string{"24.18.0"}, installed["nodejs"])
}
