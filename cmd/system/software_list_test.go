package system

import (
	"os/exec"
	"testing"
)

func TestGetSoftwareList_Checks(t *testing.T) {
	origLookPath := lookPath
	origExec := execCommand
	defer func() {
		lookPath = origLookPath
		execCommand = origExec
	}()

	list := getSoftwareList()
	if len(list) == 0 {
		t.Fatal("Software list is empty")
	}

	// Mock VS Code check
	lookPath = func(path string) (string, error) {
		if path == "code" {
			return "/usr/local/bin/code", nil
		}
		return "", exec.ErrNotFound
	}

	for _, sw := range list {
		if sw.Name == "VS Code" {
			if !sw.Check() {
				t.Error("VS Code check should return true when executable is found")
			}
		}
	}
}

func TestOpenURL(t *testing.T) {
	origExec := execCommand
	defer func() { execCommand = origExec }()

	called := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		// Just verify it attempts to call a command that looks like a URL opener
		if name == "open" || name == "xdg-open" || name == "cmd" {
			called = true
		}
		return exec.Command("echo", "success")
	}

	// Note: this might fail Start() because echo doesn't take the same args as open,
	// but we just want to see if it was called.
	_ = openURL("https://example.com")

	if !called {
		t.Error("openURL did not call any system command")
	}
}
