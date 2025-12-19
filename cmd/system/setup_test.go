package system

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlecAivazis/survey/v2"
)

func TestSetupASDF(t *testing.T) {
	tempDir := t.TempDir()
	toolVersionsPath := filepath.Join(tempDir, ".tool-versions")

	// Write a fake .tool-versions file
	plugins := []string{"nodejs 20.0.0", "python 3.11.0"}
	content := strings.Join(plugins, "\n")
	if err := os.WriteFile(toolVersionsPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write .tool-versions: %v", err)
	}

	// Set HOME to tempDir for this test
	homeOrig := os.Getenv("HOME")
	_ = os.Setenv("HOME", tempDir)
	defer func() {
		if err := os.Setenv("HOME", homeOrig); err != nil {
			t.Fatalf("Failed to restore HOME: %v", err)
		}
	}()

	// Mock exec.Command for asdf plugin add and install
	// This is a simple test to check the file parsing and command invocation logic
	// For real integration, use a test double or exec wrapper
	called := []string{}
	execCommand = func(name string, args ...string) *exec.Cmd {
		called = append(called, name+" "+strings.Join(args, " "))
		return exec.Command("echo", "mock")
	}
	defer func() { execCommand = exec.Command }()

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

func TestSetupOhMyZsh(t *testing.T) {
	tempDir := t.TempDir()
	homeOrig := os.Getenv("HOME")
	_ = os.Setenv("HOME", tempDir)
	defer func() { _ = os.Setenv("HOME", homeOrig) }()

	called := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "sh" && len(args) > 1 && strings.Contains(args[1], "ohmyzsh") {
			called = true
		}
		return exec.Command("echo", "mock")
	}
	defer func() { execCommand = exec.Command }()

	setupOhMyZsh(false)

	if !called {
		t.Error("Oh My Zsh installation command was not called")
	}
}

func TestSetupDotfiles(t *testing.T) {
	called := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		// setupDotfiles calls the executable with "dotfiles", "install"
		for _, arg := range args {
			if arg == "dotfiles" {
				called = true
			}
		}
		return exec.Command("echo", "mock")
	}
	defer func() { execCommand = exec.Command }()

	// We expect this to fail in test because os.Executable() might not be what we expect or we don't handle it
	// But we want to see if execCommand is called
	_ = setupDotfiles(false)

	if !called {
		t.Error("dotfiles install command was not called")
	}
}

func TestSetupSoftware(t *testing.T) {
	origLookPath := lookPath
	origAskOne := askOne
	origExec := execCommand
	defer func() {
		lookPath = origLookPath
		askOne = origAskOne
		execCommand = origExec
	}()

	lookPath = func(path string) (string, error) {
		return "/usr/bin/" + path, nil
	}
	// Mock select prompt
	askOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
		r := response.(*[]string)
		*r = []string{} // No optional software selected
		return nil
	}
	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("echo", "success")
	}

	setupSoftware(false)
	// If it doesn't panic and reaches here, basic flow works
}
