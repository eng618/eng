package repo

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// staleLockGrace is how long a *.lock file must exist before it is treated as
// stale when no live holder can be identified. Fresh locks are left alone so a
// concurrent git process is never disturbed.
const staleLockGrace = 30 * time.Second

// IsLockError reports whether output looks like a git lock-contention failure
// (e.g. "Unable to create '.git/index.lock': File exists"). Callers use it to
// decide whether a stale-lock cleanup and single retry are warranted.
func IsLockError(output string) bool {
	lowered := strings.ToLower(output)
	for _, marker := range []string{
		"index.lock",
		"head.lock",
		"packed-refs.lock",
		".lock exists",
		"another git process seems to be running",
		"unable to create",
		"cannot lock ref",
	} {
		if strings.Contains(lowered, marker) {
			return true
		}
	}
	return false
}

// gitDirFor returns the git directory for repoPath, following worktree gitfiles
// (a .git file containing "gitdir: <path>") back to their real location.
func gitDirFor(repoPath string) string {
	dotGit := filepath.Join(repoPath, ".git")
	info, err := os.Stat(dotGit)
	if err != nil {
		return dotGit
	}
	if info.IsDir() {
		return dotGit
	}
	raw, readErr := os.ReadFile(dotGit)
	if readErr != nil {
		return dotGit
	}
	target := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(string(raw)), "gitdir:"))
	if target == "" {
		return dotGit
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(repoPath, target)
	}
	return target
}

// LockFiles returns the paths of existing git lock files (*.lock) under the
// repository's git directory. It returns nil when there are none.
func LockFiles(repoPath string) []string {
	var locks []string
	_ = filepath.Walk(gitDirFor(repoPath), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(info.Name(), ".lock") {
			locks = append(locks, path)
		}
		return nil
	})
	return locks
}

// RemoveStaleLocks deletes leftover *.lock files in the repository's git
// directory, e.g. after a killed git process. Locks held by a live process or
// younger than staleLockGrace are refused with an error explaining why.
func RemoveStaleLocks(repoPath string) ([]string, error) {
	return RemoveStaleLocksWithGrace(repoPath, staleLockGrace)
}

// RemoveStaleLocksWithGrace is RemoveStaleLocks with an explicit freshness
// grace period. A grace <= 0 disables the age check; only use it when the
// caller guarantees no git process is running on the repository.
func RemoveStaleLocksWithGrace(repoPath string, grace time.Duration) ([]string, error) {
	locks := LockFiles(repoPath)
	if len(locks) == 0 {
		return nil, nil
	}
	if holder := lockHolder(locks[0]); holder != "" {
		return nil, fmt.Errorf(
			"git lock %s is held by a running process (%s); wait for it to finish instead of deleting the lock",
			locks[0],
			holder,
		)
	}
	if grace > 0 {
		var fresh []string
		now := time.Now()
		for _, l := range locks {
			if info, err := os.Stat(l); err == nil && now.Sub(info.ModTime()) < grace {
				fresh = append(fresh, l)
			}
		}
		if len(fresh) > 0 {
			return nil, fmt.Errorf(
				"git lock(s) %s look fresh (younger than %s); another git process may still be running: "+
					"wait and retry, or delete them manually",
				strings.Join(fresh, ", "),
				grace,
			)
		}
	}
	var removed []string
	for _, l := range locks {
		if err := os.Remove(l); err != nil && !os.IsNotExist(err) {
			return removed, fmt.Errorf("failed to remove stale git lock %s: %w", l, err)
		}
		removed = append(removed, l)
	}
	return removed, nil
}

// lockHolder reports which process (if any) currently holds lockPath open, on
// a best-effort basis. It returns "" when no holder is found or when holder
// detection tools (lsof/fuser) are unavailable.
func lockHolder(lockPath string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// #nosec G204 -- fixed binary, lock path comes from our own git dir walk.
	if out, err := exec.CommandContext(ctx, "lsof", "-t", lockPath).Output(); err == nil {
		if pids := strings.TrimSpace(string(bytes.TrimSpace(out))); pids != "" {
			return "PID " + strings.ReplaceAll(pids, "\n", ", PID ")
		}
	}
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	// #nosec G204 -- fixed binary, lock path comes from our own git dir walk.
	if out, err := exec.CommandContext(ctx2, "fuser", lockPath).CombinedOutput(); err == nil {
		// fuser prints "<path>: <pids...>"; a bare "<path>:" means no holder.
		// (Exit codes differ by platform, so parse the output instead.)
		text := strings.TrimSpace(string(bytes.TrimSpace(out)))
		if i := strings.Index(text, ":"); i >= 0 && strings.TrimSpace(text[i+1:]) != "" {
			return strings.TrimSpace(text[i+1:])
		}
	}
	return ""
}
