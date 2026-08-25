package system

import (
	"os"
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
