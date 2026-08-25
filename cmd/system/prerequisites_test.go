package system

import (
	"errors"
	"os"
	"os/exec"
	"strings"
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

func TestEnsureZsh_Success(t *testing.T) {
	origLookPath := lookPath
	defer func() { lookPath = origLookPath }()

	lookPath = func(path string) (string, error) {
		if path == "zsh" {
			return "/usr/bin/zsh", nil
		}
		return "", errors.New("not found")
	}

	if err := ensureZsh(false); err != nil {
		t.Errorf("expected nil error when zsh is installed, got %v", err)
	}
}

func TestEnsureZsh_Missing_PromptDeclined(t *testing.T) {
	origLookPath := lookPath
	origUIConfirm := ui.Confirm
	defer func() {
		lookPath = origLookPath
		ui.Confirm = origUIConfirm
	}()

	lookPath = func(path string) (string, error) {
		return "", errors.New("not found")
	}
	ui.Confirm = func(msg string, defVal bool) (bool, error) {
		return false, nil // user declines
	}

	err := ensureZsh(false)
	if err == nil {
		t.Fatal("expected error when user declines zsh installation, got nil")
	}
	if !strings.Contains(err.Error(), "declined") {
		t.Errorf("expected declined message, got %v", err)
	}
}

func TestEnsureZsh_Missing_Install_DNF(t *testing.T) {
	origLookPath := lookPath
	origExec := execCommand
	origDetect := detectDistro
	origUIConfirm := ui.Confirm
	defer func() {
		lookPath = origLookPath
		execCommand = origExec
		detectDistro = origDetect
		ui.Confirm = origUIConfirm
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

	calledDNF := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "sudo" && len(args) >= 3 && args[0] == "dnf" && args[1] == "install" && args[len(args)-1] == "zsh" {
			calledDNF = true
		}
		return exec.Command("echo", "success")
	}

	err := ensureZsh(false)
	if err != nil {
		t.Fatalf("ensureZsh failed: %v", err)
	}
	if !calledDNF {
		t.Error("ensureZsh on Fedora did not call sudo dnf install -y zsh")
	}
}

func TestEnsureZsh_Missing_Install_APT(t *testing.T) {
	origLookPath := lookPath
	origExec := execCommand
	origDetect := detectDistro
	origUIConfirm := ui.Confirm
	defer func() {
		lookPath = origLookPath
		execCommand = origExec
		detectDistro = origDetect
		ui.Confirm = origUIConfirm
	}()

	lookPath = func(path string) (string, error) {
		return "", errors.New("not found")
	}
	detectDistro = func() sysinfo.DistroInfo {
		return sysinfo.DistroInfo{ID: "ubuntu", RawOS: "linux"}
	}
	ui.Confirm = func(msg string, defVal bool) (bool, error) {
		return true, nil
	}

	calledAPT := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "sudo" && len(args) >= 3 && args[0] == "apt-get" && args[1] == "install" &&
			args[len(args)-1] == "zsh" {
			calledAPT = true
		}
		return exec.Command("echo", "success")
	}

	err := ensureZsh(false)
	if err != nil {
		t.Fatalf("ensureZsh failed: %v", err)
	}
	if !calledAPT {
		t.Error("ensureZsh on Ubuntu did not call sudo apt-get install -y zsh")
	}
}

func TestEnsureZsh_Missing_Install_Brew(t *testing.T) {
	origLookPath := lookPath
	origExec := execCommand
	origDetect := detectDistro
	origUIConfirm := ui.Confirm
	defer func() {
		lookPath = origLookPath
		execCommand = origExec
		detectDistro = origDetect
		ui.Confirm = origUIConfirm
	}()

	lookPath = func(path string) (string, error) {
		if path == "brew" {
			return "/usr/local/bin/brew", nil
		}
		return "", errors.New("not found")
	}
	detectDistro = func() sysinfo.DistroInfo {
		return sysinfo.DistroInfo{ID: "macos", RawOS: "darwin"}
	}
	ui.Confirm = func(msg string, defVal bool) (bool, error) {
		return true, nil
	}

	calledBrew := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "brew" && len(args) >= 2 && args[0] == "install" && args[1] == "zsh" {
			calledBrew = true
		}
		return exec.Command("echo", "success")
	}

	err := ensureZsh(false)
	if err != nil {
		t.Fatalf("ensureZsh failed: %v", err)
	}
	if !calledBrew {
		t.Error("ensureZsh with Homebrew did not call brew install zsh")
	}
}
