package asdf

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eng618/eng/internal/ui"
)

func TestRunCheckAllInstalled(t *testing.T) {
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
		if arg[0] == "list" {
			return exec.Command("echo", "nodejs\n *24.18.0")
		}
		return exec.Command("echo", "ok")
	}

	installMissingFlag = false
	checkNoScanFlag = true

	err := runCheck(nil, nil)
	require.NoError(t, err)
}

func TestRunCheckMissingRequirement(t *testing.T) {
	tempDir := t.TempDir()
	toolVersionsPath := filepath.Join(tempDir, ".tool-versions")
	require.NoError(t, os.WriteFile(toolVersionsPath, []byte("nodejs 20.19.5\n"), 0o644))

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

	var installed []string
	execCommand = func(name string, arg ...string) *exec.Cmd {
		if len(arg) >= 3 && arg[0] == "install" {
			installed = append(installed, arg[1]+"@"+arg[2])
			return exec.Command("echo", "installed")
		}
		if arg[0] == "list" {
			return exec.Command("echo", "nodejs\n *24.18.0")
		}
		return exec.Command("echo", "ok")
	}

	installMissingFlag = true
	checkNoScanFlag = true

	err := runCheck(nil, nil)
	require.NoError(t, err)

	assert.Equal(t, []string{"nodejs@20.19.5"}, installed)
}
