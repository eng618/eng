package asdf

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eng618/eng/internal/asdf"
	"github.com/eng618/eng/internal/ui"
)

func TestRunUpdateRoot(t *testing.T) {
	tempDir := t.TempDir()
	toolVersionsPath := filepath.Join(tempDir, ".tool-versions")
	require.NoError(t, os.WriteFile(toolVersionsPath, []byte("nodejs 24.18.0\n"), 0o644))

	origDisableProgress := ui.DisableProgress
	ui.DisableProgress = true
	defer func() {
		ui.DisableProgress = origDisableProgress
	}()

	origLookPath := lookPath
	origUserHomeDir := userHomeDir
	origExecCommand := execCommand

	defer func() {
		lookPath = origLookPath
		userHomeDir = origUserHomeDir
		execCommand = origExecCommand
	}()

	lookPath = func(_file string) (string, error) {
		return "/usr/bin/asdf", nil
	}

	userHomeDir = func() (string, error) {
		return tempDir, nil
	}

	execCommand = func(name string, arg ...string) *exec.Cmd {
		if arg[0] == "latest" && arg[1] == "nodejs" {
			return exec.Command("echo", "26.5.0")
		}
		return exec.Command("echo", "ok")
	}

	updateYesFlag = true
	updateInstallFlag = false
	updateConfigFlag = toolVersionsPath

	err := runUpdateRoot(nil, nil)
	require.NoError(t, err)

	updatedTV, err := asdf.ParseToolVersionsFile(toolVersionsPath)
	require.NoError(t, err)

	assert.Equal(t, []string{"26.5.0"}, updatedTV["nodejs"])
}
