package system

import (
	"os"
	"testing"
)

func TestEnsureGPGDependencies(t *testing.T) {
	origLookPath := lookPath
	defer func() { lookPath = origLookPath }()

	// Test when gpg is not in path
	lookPath = func(path string) (string, error) {
		return "", os.ErrNotExist
	}

	if err := ensureGPGDependencies(false); err == nil {
		t.Error("expected error when gpg is not found")
	}

	// Test when gpg is in path
	lookPath = func(path string) (string, error) {
		return "/usr/bin/" + path, nil
	}

	if err := ensureGPGDependencies(false); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
