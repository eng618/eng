package asdf

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/eng618/eng/internal/ui"
)

func TestRunStatus(t *testing.T) {
	tempDir := t.TempDir()
	toolVersionsPath := filepath.Join(tempDir, ".tool-versions")
	require.NoError(t, os.WriteFile(toolVersionsPath, []byte("nodejs 24.18.0\n"), 0644))

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
		if arg[0] == "list" {
			return exec.Command("echo", "nodejs\n  20.19.5\n *24.18.0")
		}
		return exec.Command("echo", "ok")
	}

	err := runStatus(nil, nil)
	require.NoError(t, err)
}
