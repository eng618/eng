package system

import (
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/eng618/eng/internal/sysinfo"
	"github.com/eng618/eng/internal/ui"
)

func TestEnsurePrerequisites_Success(t *testing.T) {
	origLookPath := lookPath
	origUIConfirm := ui.Confirm
	origUISelect := ui.Select
	origStat := stat
	origDetect := detectDistro
	defer func() {
		lookPath = origLookPath
		ui.Confirm = origUIConfirm
		ui.Select = origUISelect
		stat = origStat
		detectDistro = origDetect
	}()

	lookPath = func(path string) (string, error) {
		return "/usr/local/bin/" + path, nil
	}
	stat = func(name string) (os.FileInfo, error) {
		return nil, nil
	}
	detectDistro = func() sysinfo.DistroInfo {
		return sysinfo.DistroInfo{ID: "macos", RawOS: "darwin"}
	}

	err := EnsurePrerequisites(false)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}

func TestEnsurePrerequisites_Fedora_NativeGitBash(t *testing.T) {
	origLookPath := lookPath
	origUIConfirm := ui.Confirm
	origStat := stat
	origDetect := detectDistro
	defer func() {
		lookPath = origLookPath
		ui.Confirm = origUIConfirm
		stat = origStat
		detectDistro = origDetect
	}()

	lookPath = func(path string) (string, error) {
		if path == "brew" {
			return "", errors.New("not found")
		}
		return "/usr/bin/" + path, nil
	}
	stat = func(name string) (os.FileInfo, error) {
		if name == "/home/linuxbrew/.linuxbrew/bin/brew" {
			return nil, os.ErrNotExist
		}
		return nil, nil
	}
	detectDistro = func() sysinfo.DistroInfo {
		return sysinfo.DistroInfo{ID: "fedora", RawOS: "linux", PrettyName: "Fedora Linux 41"}
	}
	ui.Confirm = func(msg string, defVal bool) (bool, error) {
		return false, nil // user declines optional Homebrew on Linux
	}

	err := EnsurePrerequisites(false)
	if err != nil {
		t.Errorf("Expected nil error on Fedora with native git and bash, got %v", err)
	}
}

func TestEnsureHomebrew_NotInstalled_Declined_MacOS(t *testing.T) {
	origLookPath := lookPath
	origUIConfirm := ui.Confirm
	origStat := stat
	origDetect := detectDistro
	defer func() {
		lookPath = origLookPath
		ui.Confirm = origUIConfirm
		stat = origStat
		detectDistro = origDetect
	}()

	lookPath = func(path string) (string, error) {
		if path == "brew" {
			return "", errors.New("not found")
		}
		return "/bin/" + path, nil
	}
	stat = func(name string) (os.FileInfo, error) {
		return nil, os.ErrNotExist
	}
	detectDistro = func() sysinfo.DistroInfo {
		return sysinfo.DistroInfo{ID: "macos", RawOS: "darwin"}
	}
	ui.Confirm = func(msg string, defVal bool) (bool, error) {
		return false, nil
	}

	err := ensureHomebrew(false)
	if err == nil {
		t.Error("Expected error when macOS user declines Homebrew installation, got nil")
	}
}

func TestEnsureHomebrew_NotInstalled_Declined_Linux(t *testing.T) {
	origLookPath := lookPath
	origUIConfirm := ui.Confirm
	origStat := stat
	origDetect := detectDistro
	defer func() {
		lookPath = origLookPath
		ui.Confirm = origUIConfirm
		stat = origStat
		detectDistro = origDetect
	}()

	lookPath = func(path string) (string, error) {
		if path == "brew" {
			return "", errors.New("not found")
		}
		return "/bin/" + path, nil
	}
	stat = func(name string) (os.FileInfo, error) {
		return nil, os.ErrNotExist
	}
	detectDistro = func() sysinfo.DistroInfo {
		return sysinfo.DistroInfo{ID: "fedora", RawOS: "linux"}
	}
	ui.Confirm = func(msg string, defVal bool) (bool, error) {
		return false, nil
	}

	err := ensureHomebrew(false)
	if err != nil {
		t.Errorf("Expected nil error when Linux user declines optional Homebrew, got %v", err)
	}
}

func TestEnsureGit_DNF(t *testing.T) {
	origLookPath := lookPath
	origExec := execCommand
	origDetect := detectDistro
	defer func() {
		lookPath = origLookPath
		execCommand = origExec
		detectDistro = origDetect
	}()

	lookPath = func(path string) (string, error) {
		return "", errors.New("not found")
	}
	detectDistro = func() sysinfo.DistroInfo {
		return sysinfo.DistroInfo{ID: "fedora", RawOS: "linux"}
	}

	calledDNF := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "sudo" && len(args) >= 3 && args[0] == "dnf" && args[1] == "install" && args[len(args)-1] == "git" {
			calledDNF = true
		}
		return exec.Command("echo", "success")
	}

	err := ensureGit(false)
	if err != nil {
		t.Fatalf("ensureGit failed: %v", err)
	}
	if !calledDNF {
		t.Error("ensureGit on Fedora did not call sudo dnf install -y git")
	}
}

func TestEnsureGit_APT(t *testing.T) {
	origLookPath := lookPath
	origExec := execCommand
	origDetect := detectDistro
	defer func() {
		lookPath = origLookPath
		execCommand = origExec
		detectDistro = origDetect
	}()

	lookPath = func(path string) (string, error) {
		return "", errors.New("not found")
	}
	detectDistro = func() sysinfo.DistroInfo {
		return sysinfo.DistroInfo{ID: "ubuntu", RawOS: "linux"}
	}

	calledAPT := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "sudo" && len(args) >= 3 && args[0] == "apt-get" && args[1] == "install" &&
			args[len(args)-1] == "git" {
			calledAPT = true
		}
		return exec.Command("echo", "success")
	}

	err := ensureGit(false)
	if err != nil {
		t.Fatalf("ensureGit failed: %v", err)
	}
	if !calledAPT {
		t.Error("ensureGit on Ubuntu did not call sudo apt-get install -y git")
	}
}

func TestEnsureBash_DNF(t *testing.T) {
	origLookPath := lookPath
	origExec := execCommand
	origDetect := detectDistro
	defer func() {
		lookPath = origLookPath
		execCommand = origExec
		detectDistro = origDetect
	}()

	lookPath = func(path string) (string, error) {
		return "", errors.New("not found")
	}
	detectDistro = func() sysinfo.DistroInfo {
		return sysinfo.DistroInfo{ID: "fedora", RawOS: "linux"}
	}

	calledDNF := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "sudo" && len(args) >= 3 && args[0] == "dnf" && args[1] == "install" && args[len(args)-1] == "bash" {
			calledDNF = true
		}
		return exec.Command("echo", "success")
	}

	err := ensureBash(false)
	if err != nil {
		t.Fatalf("ensureBash failed: %v", err)
	}
	if !calledDNF {
		t.Error("ensureBash on Fedora did not call sudo dnf install -y bash")
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
