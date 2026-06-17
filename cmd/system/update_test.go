package system

import (
	"os/exec"
	"strings"
	"testing"
)

func TestUpdateCmd_Ubuntu(t *testing.T) {
	origExec := execCommand
	defer func() { execCommand = origExec }()

	called := []string{}
	execCommand = func(name string, args ...string) *exec.Cmd {
		cmdStr := name + " " + strings.Join(args, " ")
		called = append(called, cmdStr)

		// Mock uname -a to return Ubuntu
		if name == "uname" {
			return exec.Command("echo", "Linux ubuntu 5.4.0")
		}
		// Mock other commands to succeed
		return exec.Command("echo", "success")
	}

	// We can't easily run the actual cobra command here without setting up flags,
	// but we can test the logic by calling the Run function directly if we had access.
	// For now, let's test the updateUbuntu function directly.
	updateUbuntu(false, true, 60)

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
	calledBrew := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "bash" && strings.Contains(args[1], "brew update") {
			calledBrew = true
		}
		return exec.Command("echo", "success")
	}

	updateRaspberryPi(false)
	if !calledBrew {
		t.Error("updateRaspberryPi should have called updateBrew")
	}
}

func TestRunCleanup_AutoApprove(t *testing.T) {
	origExec := execCommand
	defer func() { execCommand = origExec }()

	called := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "bash" && strings.Contains(args[1], "apt-get autoremove") {
			called = true
		}
		return exec.Command("echo", "success")
	}

	runCleanup(false, true, 60)

	if !called {
		t.Error("runCleanup with autoApprove should have called apt-get autoremove")
	}
}
