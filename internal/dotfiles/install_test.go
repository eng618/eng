package dotfiles

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

func TestInstall(t *testing.T) {
	// Backup original hooks
	origEnsurePrerequisites := EnsurePrerequisites
	origHandleExistingRepo := HandleExistingRepo
	origUpdateBareRepoWorktree := UpdateBareRepoWorktree
	origFreshInstall := FreshInstall
	origCloneBareRepo := CloneBareRepo
	origBackupConflicts := BackupConflicts
	origCheckoutWorktree := CheckoutWorktree
	origInitSubmodules := InitSubmodules
	origConfigureBareRepo := ConfigureBareRepo
	origStat := Stat
	origEnsureSSH := EnsureSSH

	defer func() {
		EnsurePrerequisites = origEnsurePrerequisites
		HandleExistingRepo = origHandleExistingRepo
		UpdateBareRepoWorktree = origUpdateBareRepoWorktree
		FreshInstall = origFreshInstall
		CloneBareRepo = origCloneBareRepo
		BackupConflicts = origBackupConflicts
		CheckoutWorktree = origCheckoutWorktree
		InitSubmodules = origInitSubmodules
		ConfigureBareRepo = origConfigureBareRepo
		Stat = origStat
		EnsureSSH = origEnsureSSH
	}()

	tests := []struct {
		name                 string
		prereqsErr           error
		statErr              error
		existingRepoAction   string
		existingRepoErr      error
		updateErr            error
		freshInstallErr      error
		cloneErr             error
		backupPath           string
		hasConflicts         bool
		backupErr            error
		checkoutErr          error
		initSubmodulesErr    error
		configureBareRepoErr error
		expectedErr          error
	}{
		{
			name:         "Success - New Install",
			prereqsErr:   nil,
			statErr:      os.ErrNotExist,
			cloneErr:     nil,
			backupPath:   "",
			hasConflicts: false,
			backupErr:    nil,
			checkoutErr:  nil,
			expectedErr:  nil,
		},
		{
			name:               "Success - Existing Repo Skip",
			prereqsErr:         nil,
			statErr:            nil,
			existingRepoAction: "skip",
			expectedErr:        nil,
		},
		{
			name:               "Success - Existing Repo Update",
			prereqsErr:         nil,
			statErr:            nil,
			existingRepoAction: "update",
			updateErr:          nil,
			expectedErr:        nil,
		},
		{
			name:               "Success - Existing Repo Fresh",
			prereqsErr:         nil,
			statErr:            nil,
			existingRepoAction: "fresh",
			freshInstallErr:    nil,
			backupPath:         "",
			hasConflicts:       false,
			backupErr:          nil,
			checkoutErr:        nil,
			expectedErr:        nil,
		},
		{
			name:        "Failure - Prerequisites fail",
			prereqsErr:  errors.New("prereq check failed"),
			expectedErr: errors.New("prereq check failed"),
		},
		{
			name:            "Failure - Existing Repo Action Select fails",
			prereqsErr:      nil,
			statErr:         nil,
			existingRepoErr: errors.New("prompt select failed"),
			expectedErr:     errors.New("prompt select failed"),
		},
		{
			name:               "Failure - Existing Repo Update fails",
			prereqsErr:         nil,
			statErr:            nil,
			existingRepoAction: "update",
			updateErr:          errors.New("update failed"),
			expectedErr:        errors.New("update failed"),
		},
		{
			name:               "Failure - Existing Repo Fresh Install fails",
			prereqsErr:         nil,
			statErr:            nil,
			existingRepoAction: "fresh",
			freshInstallErr:    errors.New("fresh install clone failed"),
			expectedErr:        errors.New("fresh install clone failed"),
		},
		{
			name:        "Failure - Clone fails",
			prereqsErr:  nil,
			statErr:     os.ErrNotExist,
			cloneErr:    errors.New("git clone failed"),
			expectedErr: errors.New("git clone failed"),
		},
		{
			name:        "Failure - Backup Conflicts fails",
			prereqsErr:  nil,
			statErr:     os.ErrNotExist,
			cloneErr:    nil,
			backupErr:   errors.New("conflict scanning failed"),
			expectedErr: errors.New("conflict scanning failed"),
		},
		{
			name:        "Failure - Checkout fails",
			prereqsErr:  nil,
			statErr:     os.ErrNotExist,
			cloneErr:    nil,
			backupErr:   nil,
			checkoutErr: errors.New("git checkout failed"),
			expectedErr: errors.New("git checkout failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mocking hook functions
			EnsurePrerequisites = func(verbose bool) error {
				return tt.prereqsErr
			}
			Stat = func(name string) (os.FileInfo, error) {
				return nil, tt.statErr
			}
			HandleExistingRepo = func(bareRepoPath string) (string, error) {
				return tt.existingRepoAction, tt.existingRepoErr
			}
			UpdateBareRepoWorktree = func(ctx context.Context, bareRepoPath, homeDir string, isVerbose bool) error {
				return tt.updateErr
			}
			FreshInstall = func(ctx context.Context, bareRepoPath, repoURL, branch string) error {
				return tt.freshInstallErr
			}
			CloneBareRepo = func(ctx context.Context, repoURL, branch, bareRepoPath string) error {
				return tt.cloneErr
			}
			BackupConflicts = func(bareRepoPath, homeDir string) (string, bool, error) {
				return tt.backupPath, tt.hasConflicts, tt.backupErr
			}
			CheckoutWorktree = func(ctx context.Context, repoPath, worktreePath string) error {
				return tt.checkoutErr
			}
			InitSubmodules = func(ctx context.Context, repoPath, worktreePath string, auth *ssh.PublicKeys) error {
				return tt.initSubmodulesErr
			}
			ConfigureBareRepo = func(ctx context.Context, repoPath string) error {
				return tt.configureBareRepoErr
			}
			EnsureSSH = func(repoURL string, verbose bool) error {
				return nil
			}

			opts := InstallOptions{
				RepoURL:      "git@github.com:user/dotfiles.git",
				Branch:       "main",
				BareRepoPath: "/mock/bare/path",
				WorktreePath: "/mock/worktree/path",
				Verbose:      false,
			}

			err := Install(context.Background(), opts)

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			}
		})
	}
}

func TestIsSSHRepoURL(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"git@github.com:user/repo.git", true},
		{"ssh://git@github.com/user/repo.git", true},
		{"https://github.com/user/repo.git", false},
		{"http://github.com/user/repo.git", false},
		{"/path/to/local/repo", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			if isSSHRepoURL(tt.url) != tt.expected {
				t.Errorf("isSSHRepoURL(%s) = %v, expected %v", tt.url, !tt.expected, tt.expected)
			}
		})
	}
}

func TestHandleExistingRepo(t *testing.T) {
	origUISelect := UISelect
	defer func() { UISelect = origUISelect }()

	tests := []struct {
		name      string
		selectVal string
		selectErr error
		expected  string
		expectErr bool
	}{
		{"select skip", "skip - Use existing repository without changes", nil, "skip", false},
		{"select update", "update - Fetch and pull latest changes", nil, "update", false},
		{"select fresh", "fresh - Delete and re-clone repository", nil, "fresh", false},
		{"select error", "", errors.New("select error"), "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			UISelect = func(label string, options []string, defaultOpt string) (string, error) {
				return tt.selectVal, tt.selectErr
			}
			res, err := handleExistingRepo("/mock/path")
			if tt.expectErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if res != tt.expected {
					t.Errorf("got %q, expected %q", res, tt.expected)
				}
			}
		})
	}
}

func TestEnsureSSHIfRequired(t *testing.T) {
	err := ensureSSHIfRequired("https://github.com/user/repo", false)
	if err != nil {
		t.Errorf("unexpected error for HTTPS: %v", err)
	}
}

func TestCloneBareRepo(t *testing.T) {
	origBareClone := BareClone
	defer func() { BareClone = origBareClone }()

	BareClone = func(ctx context.Context, url, branch, path string, auth *ssh.PublicKeys) error {
		return nil
	}

	err := cloneBareRepo(context.Background(), "https://github.com/user/repo", "main", "/mock/path")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	BareClone = func(ctx context.Context, url, branch, path string, auth *ssh.PublicKeys) error {
		return errors.New("clone error")
	}

	err = cloneBareRepo(context.Background(), "https://github.com/user/repo", "main", "/mock/path")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestUpdateBareRepoWorktree(t *testing.T) {
	origFetchRepo := FetchRepo
	origPullRebaseRepo := PullRebaseRepo
	defer func() {
		FetchRepo = origFetchRepo
		PullRebaseRepo = origPullRebaseRepo
	}()

	FetchRepo = func(ctx context.Context, repoPath, worktreePath string) error { return nil }
	PullRebaseRepo = func(ctx context.Context, repoPath, worktreePath string) error { return nil }

	err := updateBareRepoWorktree(context.Background(), "/mock/repo", "/mock/worktree", false)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	FetchRepo = func(ctx context.Context, repoPath, worktreePath string) error { return errors.New("fetch error") }
	err = updateBareRepoWorktree(context.Background(), "/mock/repo", "/mock/worktree", false)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestFreshInstall(t *testing.T) {
	origBareClone := BareClone
	defer func() { BareClone = origBareClone }()

	BareClone = func(ctx context.Context, url, branch, path string, auth *ssh.PublicKeys) error {
		return nil
	}

	tmpDir, err := os.MkdirTemp("", "fresh-install-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = freshInstall(context.Background(), tmpDir, "https://github.com/user/repo", "main")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBackupConflicts(t *testing.T) {
	bareDir, err := os.MkdirTemp("", "bare-repo-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(bareDir)

	worktreeDir, err := os.MkdirTemp("", "worktree-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(worktreeDir)

	r, err := git.PlainInit(bareDir, false)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	w, err := r.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}

	fileName := "conflict.txt"
	filePath := filepath.Join(bareDir, fileName)
	err = os.WriteFile(filePath, []byte("repo content"), 0o644)
	if err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	_, err = w.Add(fileName)
	if err != nil {
		t.Fatalf("failed to add file: %v", err)
	}

	_, err = w.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	worktreeFile := filepath.Join(worktreeDir, fileName)
	err = os.WriteFile(worktreeFile, []byte("user content"), 0o644)
	if err != nil {
		t.Fatalf("failed to write worktree file: %v", err)
	}

	backupPath, hasConflicts, err := backupConflicts(bareDir, worktreeDir)
	if err != nil {
		t.Fatalf("backupConflicts failed: %v", err)
	}

	if !hasConflicts {
		t.Error("expected hasConflicts to be true")
	}

	if backupPath == "" {
		t.Error("expected backupPath to be non-empty")
	}

	backupFile := filepath.Join(backupPath, fileName)
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		t.Errorf("expected backup file %s to exist", backupFile)
	}

	if _, err := os.Stat(worktreeFile); !os.IsNotExist(err) {
		t.Error("expected conflicting file in worktree to be moved")
	}
}
