package gpg

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureGPGConfiguration(t *testing.T) {
	tempDir := t.TempDir()
	homeOrig := os.Getenv("HOME")
	_ = os.Setenv("HOME", tempDir)
	origExec := execCommand
	defer func() {
		_ = os.Setenv("HOME", homeOrig)
		execCommand = origExec
	}()

	// Setup fake ~/.gnupg and ~/.local/bin/pinentry
	gpgDir := filepath.Join(tempDir, ".gnupg")
	_ = os.MkdirAll(gpgDir, 0o755)
	_ = os.WriteFile(filepath.Join(gpgDir, "gpg-agent.conf"), []byte("# test"), 0o644)

	localBin := filepath.Join(tempDir, ".local", "bin")
	_ = os.MkdirAll(localBin, 0o755)
	pinentryPath := filepath.Join(localBin, "pinentry")
	_ = os.WriteFile(pinentryPath, []byte("#!/bin/sh\n"), 0o644)

	var commandsCalled []string
	execCommand = func(name string, args ...string) *exec.Cmd {
		commandsCalled = append(commandsCalled, name+" "+strings.Join(args, " "))
		return exec.Command("echo", "mock")
	}

	ensureGPGConfiguration(false)

	// Check that ~/.gnupg directory permissions are 0700
	info, err := os.Stat(gpgDir)
	if err != nil {
		t.Fatalf("failed to stat gpg dir: %v", err)
	}
	if info.Mode().Perm() != 0o700 {
		t.Errorf("expected gpgDir mode 0700, got %o", info.Mode().Perm())
	}

	// Check pinentry wrapper mode is 0755
	pinfo, err := os.Stat(pinentryPath)
	if err != nil {
		t.Fatalf("failed to stat pinentry: %v", err)
	}
	if pinfo.Mode().Perm() != 0o755 {
		t.Errorf("expected pinentry mode 0755, got %o", pinfo.Mode().Perm())
	}

	// Check gpgconf was called
	joined := strings.Join(commandsCalled, "\n")
	if !strings.Contains(joined, "gpgconf --kill gpg-agent") {
		t.Errorf("expected gpgconf --kill gpg-agent to be called, got: %s", joined)
	}
	if !strings.Contains(joined, "gpgconf --launch gpg-agent") {
		t.Errorf("expected gpgconf --launch gpg-agent to be called, got: %s", joined)
	}
}
