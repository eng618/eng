package browser

import (
	"os/exec"
	"testing"
)

func TestOpenURL(t *testing.T) {
	origExec := execCommand
	defer func() { execCommand = origExec }()

	called := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "open" || name == "xdg-open" || name == "cmd" {
			called = true
		}
		return exec.Command("echo", "success")
	}

	_ = OpenURL("https://example.com")

	if !called {
		t.Error("OpenURL did not call any system command")
	}
}
