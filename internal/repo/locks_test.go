package repo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestIsLockError(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   bool
	}{
		{
			name:   "index lock exists",
			output: "fatal: Unable to create '/repo/.git/index.lock': File exists.",
			want:   true,
		},
		{
			name:   "another process running",
			output: "fatal: Unable to create '/repo/.git/HEAD.lock': File exists. If no other git process is currently running, this probably means a git process crashed in this repository earlier.",
			want:   true,
		},
		{
			name:   "cannot lock ref",
			output: "error: cannot lock ref 'refs/heads/main': is at abc but expected def",
			want:   true,
		},
		{
			name:   "packed refs lock",
			output: "fatal: cannot lock 'packed-refs': Unable to create '/repo/.git/packed-refs.lock'",
			want:   true,
		},
		{
			name:   "ordinary failure",
			output: "fatal: Could not read from remote repository.",
			want:   false,
		},
		{
			name:   "empty",
			output: "",
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsLockError(tt.output); got != tt.want {
				t.Errorf("IsLockError(%q) = %v, want %v", tt.output, got, tt.want)
			}
		})
	}
}

func TestLockFilesEmpty(t *testing.T) {
	dir := t.TempDir()
	if locks := LockFiles(dir); len(locks) != 0 {
		t.Errorf("Expected no locks in empty dir, got %v", locks)
	}
	if removed, err := RemoveStaleLocks(dir); err != nil || len(removed) != 0 {
		t.Errorf("Expected no-op cleanup, got removed=%v err=%v", removed, err)
	}
}

func TestRemoveStaleLocks(t *testing.T) {
	repoDir := t.TempDir()
	gitDir := filepath.Join(repoDir, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}
	lockPath := filepath.Join(gitDir, "index.lock")
	if err := os.WriteFile(lockPath, []byte("stale"), 0o644); err != nil {
		t.Fatalf("Failed to create lock file: %v", err)
	}
	// Backdate so the grace period treats it as stale, not live.
	old := time.Now().Add(-time.Hour)
	if err := os.Chtimes(lockPath, old, old); err != nil {
		t.Fatalf("Failed to backdate lock file: %v", err)
	}

	locks := LockFiles(repoDir)
	if len(locks) != 1 || locks[0] != lockPath {
		t.Fatalf("Expected LockFiles to find %s, got %v", lockPath, locks)
	}

	removed, err := RemoveStaleLocks(repoDir)
	if err != nil {
		t.Fatalf("RemoveStaleLocks() error = %v", err)
	}
	if len(removed) != 1 || removed[0] != lockPath {
		t.Errorf("Expected removed=[%s], got %v", lockPath, removed)
	}
	if _, statErr := os.Stat(lockPath); !os.IsNotExist(statErr) {
		t.Error("Expected lock file to be deleted")
	}
}

func TestRemoveStaleLocksRefusesFresh(t *testing.T) {
	repoDir := t.TempDir()
	gitDir := filepath.Join(repoDir, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}
	lockPath := filepath.Join(gitDir, "HEAD.lock")
	if err := os.WriteFile(lockPath, []byte("fresh"), 0o644); err != nil {
		t.Fatalf("Failed to create lock file: %v", err)
	}

	if _, err := RemoveStaleLocks(repoDir); err == nil {
		t.Error("Expected fresh lock to be refused, got nil error")
	} else if !strings.Contains(err.Error(), "fresh") {
		t.Errorf("Expected freshness explanation, got: %v", err)
	}
	if _, statErr := os.Stat(lockPath); statErr != nil {
		t.Error("Fresh lock file must be left alone")
	}

	// Zero grace is the escape hatch for callers that guarantee quiescence.
	removed, err := RemoveStaleLocksWithGrace(repoDir, 0)
	if err != nil {
		t.Fatalf("RemoveStaleLocksWithGrace(0) error = %v", err)
	}
	if len(removed) != 1 {
		t.Errorf("Expected 1 removed lock, got %v", removed)
	}
}
