package system

import (
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

	setupASDF()

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
