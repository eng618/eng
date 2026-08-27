package system

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eng618/eng/internal/sysinfo"
)

func TestEnsureGPGDependencies_MissingGPG_Fedora(t *testing.T) {
	origLookPath := lookPath
	origDetect := detectDistro
	defer func() {
		lookPath = origLookPath
		detectDistro = origDetect
	}()

	lookPath = func(path string) (string, error) {
		return "", os.ErrNotExist
	}
	detectDistro = func() sysinfo.DistroInfo {
		return sysinfo.DistroInfo{ID: "fedora", RawOS: "linux"}
	}

	err := ensureGPGDependencies(false)
	if err == nil {
		t.Fatal("expected error when gpg is not found")
	}
	if !strings.Contains(err.Error(), "sudo dnf install -y gnupg2") {
		t.Errorf("expected Fedora DNF instruction in error, got %v", err)
	}
}

func TestEnsureGPGDependencies_MissingPinentry_Fedora(t *testing.T) {
	origLookPath := lookPath
	origDetect := detectDistro
	defer func() {
		lookPath = origLookPath
		detectDistro = origDetect
	}()

	lookPath = func(path string) (string, error) {
		if path == "gpg" {
			return "/usr/bin/gpg", nil
		}
		return "", os.ErrNotExist
	}
	detectDistro = func() sysinfo.DistroInfo {
		return sysinfo.DistroInfo{ID: "fedora", RawOS: "linux"}
	}

	err := ensureGPGDependencies(false)
	if err == nil {
		t.Fatal("expected error when pinentry is not found")
	}
	if !strings.Contains(err.Error(), "sudo dnf install -y pinentry") {
		t.Errorf("expected Fedora DNF pinentry instruction in error, got %v", err)
	}
}

func TestEnsureGPGDependencies_Linux_PinentryVariants(t *testing.T) {
	origLookPath := lookPath
	origDetect := detectDistro
	defer func() {
		lookPath = origLookPath
		detectDistro = origDetect
	}()

	detectDistro = func() sysinfo.DistroInfo {
		return sysinfo.DistroInfo{ID: "fedora", RawOS: "linux"}
	}

	// Test pinentry-curses resolution
	lookPath = func(path string) (string, error) {
		if path == "gpg" || path == "pinentry-curses" {
			return "/usr/bin/" + path, nil
		}
		return "", os.ErrNotExist
	}

	if err := ensureGPGDependencies(false); err != nil {
		t.Errorf("expected success with pinentry-curses on Fedora, got %v", err)
	}
}

func TestEnsureGPGDependencies_Success_MacOS(t *testing.T) {
	origLookPath := lookPath
	origDetect := detectDistro
	defer func() {
		lookPath = origLookPath
		detectDistro = origDetect
	}()

	lookPath = func(path string) (string, error) {
		return "/usr/local/bin/" + path, nil
	}
	detectDistro = func() sysinfo.DistroInfo {
		return sysinfo.DistroInfo{ID: "macos", RawOS: "darwin"}
	}

	if err := ensureGPGDependencies(false); err != nil {
		t.Errorf("unexpected error on macOS: %v", err)
	}
}

func TestResolvePrimaryFingerprint_FullFingerprint(t *testing.T) {
	// 40-character fingerprint should resolve directly
	fpr := "BE363376C8A71C92C2E1DB137C180F0FCB31441B"
	resolved, err := resolvePrimaryFingerprint(fpr, false)
	if err != nil {
		t.Fatalf("expected 40-char fingerprint to resolve, got error: %v", err)
	}
	if resolved != fpr {
		t.Errorf("expected %s, got %s", fpr, resolved)
	}
}

func TestListLocalSecretGPGKeys(t *testing.T) {
	// Test that listLocalSecretGPGKeys runs without panic
	_, _ = listLocalSecretGPGKeys(false)
}

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
