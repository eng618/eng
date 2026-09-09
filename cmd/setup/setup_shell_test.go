package setup

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/eng618/eng/internal/sysinfo"
	"github.com/eng618/eng/internal/ui"
)

func TestSetupOhMyZsh(t *testing.T) {
	tempDir := t.TempDir()
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
		return "/bin/" + path, nil
	}

	called := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "sh" && len(args) > 1 && strings.Contains(args[1], "ohmyzsh") {
			called = true
		}
		return exec.Command("echo", "mock")
	}

	setupOhMyZsh(false)

	if !called {
		t.Error("Oh My Zsh installation command was not called")
	}
}

func TestSetupOhMyZsh_EnsuresZshWhenMissing(t *testing.T) {
	tempDir := t.TempDir()
	homeOrig := os.Getenv("HOME")
	_ = os.Setenv("HOME", tempDir)
	origLookPath := lookPath
	origUIConfirm := ui.Confirm
	origExec := execCommand
	origDetect := detectDistro
	defer func() {
		_ = os.Setenv("HOME", homeOrig)
		lookPath = origLookPath
		ui.Confirm = origUIConfirm
		execCommand = origExec
		detectDistro = origDetect
	}()

	lookPath = func(path string) (string, error) {
		return "", errors.New("not found")
	}
	detectDistro = func() sysinfo.DistroInfo {
		return sysinfo.DistroInfo{ID: "fedora", RawOS: "linux"}
	}
	ui.Confirm = func(msg string, defVal bool) (bool, error) {
		return true, nil
	}

	installedZsh := false
	installedOMZ := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "sudo" && len(args) >= 3 && args[0] == "dnf" && args[1] == "install" && args[len(args)-1] == "zsh" {
			installedZsh = true
		}
		if name == "sh" && len(args) > 1 && strings.Contains(args[1], "ohmyzsh") {
			installedOMZ = true
		}
		return exec.Command("echo", "mock")
	}

	setupOhMyZsh(false)

	if !installedZsh {
		t.Error("Expected setupOhMyZsh to prompt and install Zsh when missing")
	}
	if !installedOMZ {
		t.Error("Expected setupOhMyZsh to install Oh My Zsh after installing Zsh")
	}
}
