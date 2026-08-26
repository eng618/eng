package system

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/eng618/eng/internal/sysinfo"
	"github.com/eng618/eng/internal/ui"
)

func TestUpdateCmd_Fedora(t *testing.T) {
	origExec := execCommand
	ui.DisableProgress = true
	defer func() {
		execCommand = origExec
		ui.DisableProgress = false
	}()

	called := []string{}
	execCommand = func(name string, args ...string) *exec.Cmd {
		cmdStr := name + " " + strings.Join(args, " ")
		called = append(called, cmdStr)
		return exec.Command("echo", "success")
	}

	updateFedora(false, true, 60)

	expected := "bash -c sudo dnf upgrade --refresh -y"
	found := false
	for _, c := range called {
		if c == expected {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected command %q to be called, but it wasn't. Called: %v", expected, called)
	}
}

func TestUpdateCmd_Ubuntu(t *testing.T) {
	origExec := execCommand
	ui.DisableProgress = true
	defer func() {
		execCommand = origExec
		ui.DisableProgress = false
	}()

	called := []string{}
	execCommand = func(name string, args ...string) *exec.Cmd {
		cmdStr := name + " " + strings.Join(args, " ")
		called = append(called, cmdStr)
		return exec.Command("echo", "success")
	}

	updateDebianUbuntu(false, true, 60)

	expected := "bash -c sudo apt-get update && sudo apt-get upgrade -y"
	found := false
	for _, c := range called {
		if c == expected {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected command %q to be called, but it wasn't. Called: %v", expected, called)
	}
}

func TestUpdateCmd_Dispatch_Fedora(t *testing.T) {
	origDetect := detectDistro
	origExec := execCommand
	ui.DisableProgress = true
	_ = UpdateCmd.Flags().Set("yes", "true")
	defer func() {
		detectDistro = origDetect
		execCommand = origExec
		ui.DisableProgress = false
		_ = UpdateCmd.Flags().Set("yes", "false")
	}()

	detectDistro = func() sysinfo.DistroInfo {
		return sysinfo.DistroInfo{
			ID:         "fedora",
			PrettyName: "Fedora Linux 41",
			RawOS:      "linux",
		}
	}

	calledDNF := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "bash" && len(args) > 1 && strings.Contains(args[1], "dnf upgrade") {
			calledDNF = true
		}
		return exec.Command("echo", "success")
	}

	UpdateCmd.Run(UpdateCmd, []string{})

	if !calledDNF {
		t.Error("UpdateCmd dispatch for Fedora did not execute dnf upgrade")
	}
}

func TestUpdateCmd_Dispatch_Ubuntu(t *testing.T) {
	origDetect := detectDistro
	origExec := execCommand
	ui.DisableProgress = true
	_ = UpdateCmd.Flags().Set("yes", "true")
	defer func() {
		detectDistro = origDetect
		execCommand = origExec
		ui.DisableProgress = false
		_ = UpdateCmd.Flags().Set("yes", "false")
	}()

	detectDistro = func() sysinfo.DistroInfo {
		return sysinfo.DistroInfo{
			ID:         "ubuntu",
			PrettyName: "Ubuntu 24.04",
			RawOS:      "linux",
		}
	}

	calledAPT := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "bash" && len(args) > 1 && strings.Contains(args[1], "apt-get update") {
			calledAPT = true
		}
		return exec.Command("echo", "success")
	}

	UpdateCmd.Run(UpdateCmd, []string{})

	if !calledAPT {
		t.Error("UpdateCmd dispatch for Ubuntu did not execute apt-get update")
	}
}

func TestUpdateFlatpak(t *testing.T) {
	origLookPath := lookPath
	origExec := execCommand
	ui.DisableProgress = true
	defer func() {
		lookPath = origLookPath
		execCommand = origExec
		ui.DisableProgress = false
	}()

	lookPath = func(path string) (string, error) {
		if path == "flatpak" {
			return "/usr/bin/flatpak", nil
		}
		return "", exec.ErrNotFound
	}

	called := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "flatpak" && len(args) > 0 && args[0] == "update" {
			called = true
		}
		return exec.Command("echo", "success")
	}

	updateFlatpak(false)

	if !called {
		t.Error("updateFlatpak did not call flatpak update")
	}
}

func TestUpdateBrew(t *testing.T) {
	origLookPath := lookPath
	origExec := execCommand
	defer func() {
		lookPath = origLookPath
		execCommand = origExec
	}()

	lookPath = func(path string) (string, error) {
		return "/usr/local/bin/brew", nil
	}

	called := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "bash" && strings.Contains(args[1], "brew update") {
			called = true
		}
		return exec.Command("echo", "success")
	}

	updateBrew(false)

	if !called {
		t.Error("updateBrew did not call brew update command")
	}
}

func TestUpdateMacOS(t *testing.T) {
	origLookPath := lookPath
	origExec := execCommand
	defer func() {
		lookPath = origLookPath
		execCommand = origExec
	}()

	lookPath = func(path string) (string, error) {
		return "/usr/bin/" + path, nil
	}
	calledBrew := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "bash" && strings.Contains(args[1], "brew update") {
			calledBrew = true
		}
		return exec.Command("echo", "success")
	}

	updateMacOS(false)
	if !calledBrew {
		t.Error("updateMacOS should have called updateBrew")
	}
}

func TestUpdateRaspberryPi(t *testing.T) {
	origLookPath := lookPath
	origExec := execCommand
	defer func() {
		lookPath = origLookPath
		execCommand = origExec
	}()

	lookPath = func(path string) (string, error) {
		return "/usr/bin/" + path, nil
	}
	calledAPT := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "bash" && strings.Contains(args[1], "apt-get update") {
			calledAPT = true
		}
		return exec.Command("echo", "success")
	}

	updateRaspberryPi(false, true, 60)
	if !calledAPT {
		t.Error("updateRaspberryPi should have called apt-get update")
	}
}

func TestRunDnfCleanup_AutoApprove(t *testing.T) {
	origExec := execCommand
	ui.DisableProgress = true
	defer func() {
		execCommand = origExec
		ui.DisableProgress = false
	}()

	called := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "sudo" && len(args) > 1 && args[0] == "dnf" && args[1] == "autoremove" {
			called = true
		}
		return exec.Command("echo", "success")
	}

	runDnfCleanup(false, true, 60)

	if !called {
		t.Error("runDnfCleanup with autoApprove should have called dnf autoremove")
	}
}

func TestRunAptCleanup_AutoApprove(t *testing.T) {
	origExec := execCommand
	ui.DisableProgress = true
	defer func() {
		execCommand = origExec
		ui.DisableProgress = false
	}()

	called := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "sudo" && len(args) > 1 && args[0] == "apt-get" && args[1] == "autoremove" {
			called = true
		}
		return exec.Command("echo", "success")
	}

	runAptCleanup(false, true, 60)

	if !called {
		t.Error("runAptCleanup with autoApprove should have called apt-get autoremove")
	}
}
