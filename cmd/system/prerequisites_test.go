package system

import (
	"errors"
	"os"
	"testing"

	"github.com/AlecAivazis/survey/v2"
)

func TestEnsurePrerequisites_Success(t *testing.T) {
	// Backup original values
	origLookPath := lookPath
	origAskOne := askOne
	origStat := stat
	defer func() {
		lookPath = origLookPath
		askOne = origAskOne
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
	origAskOne := askOne
	defer func() {
		lookPath = origLookPath
		askOne = origAskOne
	}()

	// Mock brew not found
	lookPath = func(path string) (string, error) {
		if path == "brew" {
			return "", errors.New("not found")
		}
		return "/bin/" + path, nil
	}

	// Mock user declining installation
	askOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
		r := response.(*bool)
		*r = false
		return nil
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
