package setup

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetupASDF(t *testing.T) {
	tempDir := t.TempDir()
	toolVersionsPath := filepath.Join(tempDir, ".tool-versions")

	// Write a fake .tool-versions file
	plugins := []string{"nodejs 20.0.0", "python 3.11.0"}
	content := strings.Join(plugins, "\n")
	if err := os.WriteFile(toolVersionsPath, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to write .tool-versions: %v", err)
	}

	// Set HOME to tempDir for this test
	homeOrig := os.Getenv("HOME")
	_ = os.Setenv("HOME", tempDir)
	origLookPath := lookPath
	origExec := execCommand
	defer func() {
		_ = os.Setenv("HOME", homeOrig)
		lookPath = origLookPath
		execCommand = origExec
	}()

	lookPath = func(path string) (string, error) {
		if path == "asdf" {
			return "/usr/local/bin/asdf", nil
		}
		return "", errors.New("not found")
	}

	called := []string{}
	execCommand = func(name string, args ...string) *exec.Cmd {
		called = append(called, name+" "+strings.Join(args, " "))
		return exec.Command("echo", "mock")
	}

	setupASDF(false)

	if len(called) == 0 {
		t.Error("No commands were called, expected asdf plugin add and install")
	}
	foundInstall := false
	for _, c := range called {
		if strings.Contains(c, "asdf install") {
			foundInstall = true
		}
	}
	if !foundInstall {
		t.Error("asdf install was not called")
	}
}

func TestEnsureASDFInstalled_InstallViaBrew(t *testing.T) {
	origLookPath := lookPath
	origExec := execCommand
	defer func() {
		lookPath = origLookPath
		execCommand = origExec
	}()

	lookPath = func(path string) (string, error) {
		if path == "brew" {
			return "/usr/local/bin/brew", nil
		}
		return "", errors.New("not found")
	}

	calledBrewInstall := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "brew" && len(args) >= 2 && args[0] == "install" && args[1] == "asdf" {
			calledBrewInstall = true
		}
		return exec.Command("echo", "mock")
	}

	err := ensureASDFInstalled(false)
	if err != nil {
		t.Fatalf("ensureASDFInstalled returned error: %v", err)
	}
	if !calledBrewInstall {
		t.Error("expected brew install asdf to be called when brew is available")
	}
}

func TestEnsureASDFInstalled_InstallViaGit(t *testing.T) {
	tempDir := t.TempDir()
	homeOrig := os.Getenv("HOME")
	_ = os.Setenv("HOME", tempDir)
	origLookPath := lookPath
	origExec := execCommand
	origStat := stat
	defer func() {
		_ = os.Setenv("HOME", homeOrig)
		lookPath = origLookPath
		execCommand = origExec
		stat = origStat
	}()

	lookPath = func(path string) (string, error) {
		if path == "git" {
			return "/usr/bin/git", nil
		}
		return "", errors.New("not found")
	}

	stat = func(name string) (os.FileInfo, error) {
		return nil, os.ErrNotExist
	}

	calledGitClone := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "git" && len(args) >= 2 && args[0] == "clone" {
			calledGitClone = true
		}
		return exec.Command("echo", "mock")
	}

	err := ensureASDFInstalled(false)
	if err != nil {
		t.Fatalf("ensureASDFInstalled via git returned error: %v", err)
	}
	if !calledGitClone {
		t.Error("expected git clone asdf to be called when brew is not available")
	}
}
