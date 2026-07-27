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

func TestRunPruneDryRun(t *testing.T) {
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

	lookPath = func(file string) (string, error) {
		return "/usr/bin/asdf", nil
	}

	userHomeDir = func() (string, error) {
		return tempDir, nil
	}

	var executedCommands []string
	execCommand = func(name string, arg ...string) *exec.Cmd {
		cmdStr := name + " " + filepath.Base(arg[0])
		if len(arg) > 1 {
			cmdStr += " " + arg[1]
		}
		executedCommands = append(executedCommands, cmdStr)

		if arg[0] == "list" {
			return exec.Command("echo", "nodejs\n  20.19.5\n *24.18.0")
		}
		return exec.Command("echo", "ok")
	}

	interactiveFlag = false
	dryRunFlag = true
	forceFlag = false
	pluginFlag = ""
	configFlag = toolVersionsPath
	noScanFlag = true
	scanDirFlag = ""

	err := runPrune(nil, nil)
	require.NoError(t, err)

	assert.Contains(t, executedCommands, "asdf list")
	for _, cmd := range executedCommands {
		assert.NotContains(t, cmd, "asdf uninstall")
	}
}

func TestRunPruneMultiProject(t *testing.T) {
	tempDir := t.TempDir()
	globalFile := filepath.Join(tempDir, ".tool-versions")
	require.NoError(t, os.WriteFile(globalFile, []byte("nodejs 24.18.0\n"), 0644))

	projDir := filepath.Join(tempDir, "Development", "legacy-app")
	require.NoError(t, os.MkdirAll(projDir, 0755))
	projFile := filepath.Join(projDir, ".tool-versions")
	require.NoError(t, os.WriteFile(projFile, []byte("nodejs 20.19.5\n"), 0644))

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

	lookPath = func(file string) (string, error) {
		return "/usr/bin/asdf", nil
	}

	userHomeDir = func() (string, error) {
		return tempDir, nil
	}

	execCommand = func(name string, arg ...string) *exec.Cmd {
		if arg[0] == "list" {
			return exec.Command("echo", "nodejs\n  20.19.5\n  22.14.0\n *24.18.0")
		}
		return exec.Command("echo", "ok")
	}

	interactiveFlag = false
	dryRunFlag = true
	forceFlag = false
	pluginFlag = ""
	configFlag = ""
	noScanFlag = false
	scanDirFlag = tempDir

	err := runPrune(nil, nil)
	require.NoError(t, err)
}
