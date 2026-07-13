package system

import (
	"errors"
	"os"
	"testing"

	"github.com/eng618/eng/internal/ui"
)

func TestEnsurePrerequisites_Success(t *testing.T) {
	// Backup original values
	origLookPath := lookPath
	origUIConfirm := ui.Confirm
	origUISelect := ui.Select
	origStat := stat
	defer func() {
		lookPath = origLookPath
		ui.Confirm = origUIConfirm
		ui.Select = origUISelect
		stat = origStat
	}()

	// Mock all checks to succeed
	lookPath = func(path string) (string, error) {
		return "/usr/local/bin/" + path, nil
	}
	stat = func(name string) (os.FileInfo, error) {
		return nil, nil // SSH key exists
	}

	err := EnsurePrerequisites(false)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}

func TestEnsureHomebrew_NotInstalled_Declined(t *testing.T) {
	origLookPath := lookPath
	origUIConfirm := ui.Confirm
	origUISelect := ui.Select
	defer func() {
		lookPath = origLookPath
		ui.Confirm = origUIConfirm
		ui.Select = origUISelect
	}()

	// Mock brew not found
	lookPath = func(path string) (string, error) {
		if path == "brew" {
			return "", errors.New("not found")
		}
		return "/bin/" + path, nil
	}

	// Mock user declining installation
	ui.Confirm = func(msg string, defVal bool) (bool, error) {
		return false, nil
	}

	err := ensureHomebrew(false)
	if err == nil {
		t.Error("Expected error when user declines installation, got nil")
	}
}

func TestEnsureGitHubSSH_Missing(t *testing.T) {
	origUserHomeDir := userHomeDir
	origStat := stat
	defer func() {
		userHomeDir = origUserHomeDir
		stat = origStat
	}()

	userHomeDir = func() (string, error) {
		return "/tmp/fakehome", nil
	}
	stat = func(name string) (os.FileInfo, error) {
		return nil, os.ErrNotExist
	}

	err := ensureGitHubSSH(false)
	if err == nil {
		t.Error("Expected error when SSH key is missing, got nil")
	}
}
