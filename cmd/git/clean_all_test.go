package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func setupLocalTestRepo(t *testing.T) (string, func()) {
	t.Helper()
	workspace, err := os.MkdirTemp("", "clean-all-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp workspace: %v", err)
	}

	repoPath := filepath.Join(workspace, "repo1")
	if err := os.MkdirAll(repoPath, 0o755); err != nil {
		t.Fatalf("Failed to create repo dir: %v", err)
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(workspace)
	}
	return workspace, cleanup
}

func setupLocalTestCommand() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.PersistentFlags().Bool("current", true, "Use current working directory")
	cmd.Flags().Bool("dry-run", false, "Perform a dry run")
	cmd.Flags().Bool("force", false, "Force flag")
	cmd.Flags().Bool("directories", false, "Directories flag")
	return cmd
}

func TestHasUntrackedFiles(t *testing.T) {
	workspace, cleanup := setupLocalTestRepo(t)
	defer cleanup()

	repoPath := filepath.Join(workspace, "repo1")

	// 1. Clean repo
	hasUntracked, err := hasUntrackedFiles(repoPath)
	if err != nil {
		t.Fatalf("hasUntrackedFiles failed on clean repo: %v", err)
	}
	if hasUntracked {
		t.Errorf("hasUntrackedFiles returned true for clean repo, expected false")
	}

	// 2. Dirty repo with untracked file
	testFile := filepath.Join(repoPath, "untracked.txt")
	if err := os.WriteFile(testFile, []byte("untracked file content"), 0o644); err != nil {
		t.Fatalf("Failed to create untracked file: %v", err)
	}

	hasUntracked, err = hasUntrackedFiles(repoPath)
	if err != nil {
		t.Fatalf("hasUntrackedFiles failed on dirty repo: %v", err)
	}
	if !hasUntracked {
		t.Errorf("hasUntrackedFiles returned false for dirty repo, expected true")
	}
}

func TestCleanRepository(t *testing.T) {
	workspace, cleanup := setupLocalTestRepo(t)
	defer cleanup()

	repoPath := filepath.Join(workspace, "repo1")

	// Create untracked file
	if err := os.WriteFile(filepath.Join(repoPath, "untracked.txt"), []byte("content"), 0o644); err != nil {
		t.Fatalf("Failed to create untracked file: %v", err)
	}

	// Create untracked dir
	untrackedDir := filepath.Join(repoPath, "untracked_dir")
	if err := os.MkdirAll(untrackedDir, 0o755); err != nil {
		t.Fatalf("Failed to create untracked dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(untrackedDir, "file.txt"), []byte("content"), 0o644); err != nil {
		t.Fatalf("Failed to create untracked file in dir: %v", err)
	}

	// Clean with directories=false
	if err := cleanRepository(repoPath, false); err != nil {
		t.Fatalf("cleanRepository failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(repoPath, "untracked.txt")); !os.IsNotExist(err) {
		t.Errorf("untracked.txt should have been deleted")
	}
	if _, err := os.Stat(untrackedDir); os.IsNotExist(err) {
		t.Errorf("untracked_dir should NOT have been deleted when directories=false")
	}

	// Create another file to test directories=true
	if err := os.WriteFile(filepath.Join(repoPath, "untracked2.txt"), []byte("content"), 0o644); err != nil {
		t.Fatalf("Failed to create untracked file: %v", err)
	}

	// Clean with directories=true
	if err := cleanRepository(repoPath, true); err != nil {
		t.Fatalf("cleanRepository failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(repoPath, "untracked2.txt")); !os.IsNotExist(err) {
		t.Errorf("untracked2.txt should have been deleted")
	}
	if _, err := os.Stat(untrackedDir); !os.IsNotExist(err) {
		t.Errorf("untracked_dir should have been deleted when directories=true")
	}
}

func TestCleanAllCommand_DryRun(t *testing.T) {
	workspace, cleanup := setupLocalTestRepo(t)
	defer cleanup()

	originalDir, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(originalDir)
	}()
	if err := os.Chdir(workspace); err != nil {
		t.Fatalf("Failed to change to workspace directory: %v", err)
	}

	repoPath := filepath.Join(workspace, "repo1")
	untrackedFile := filepath.Join(repoPath, "untracked.txt")
	if err := os.WriteFile(untrackedFile, []byte("content"), 0o644); err != nil {
		t.Fatalf("Failed to create untracked file: %v", err)
	}

	cmd := setupLocalTestCommand()
	if err := cmd.Flags().Set("dry-run", "true"); err != nil {
		t.Fatalf("Failed to set dry-run flag: %v", err)
	}

	CleanAllCmd.Run(cmd, []string{})

	if _, err := os.Stat(untrackedFile); os.IsNotExist(err) {
		t.Errorf("untracked.txt should NOT have been deleted during dry-run")
	}
}

func TestCleanAllCommand_Force(t *testing.T) {
	workspace, cleanup := setupLocalTestRepo(t)
	defer cleanup()

	originalDir, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(originalDir)
	}()
	if err := os.Chdir(workspace); err != nil {
		t.Fatalf("Failed to change to workspace directory: %v", err)
	}

	repoPath := filepath.Join(workspace, "repo1")
	untrackedFile := filepath.Join(repoPath, "untracked.txt")
	if err := os.WriteFile(untrackedFile, []byte("content"), 0o644); err != nil {
		t.Fatalf("Failed to create untracked file: %v", err)
	}

	cmd := setupLocalTestCommand()
	if err := cmd.Flags().Set("force", "true"); err != nil {
		t.Fatalf("Failed to set force flag: %v", err)
	}

	CleanAllCmd.Run(cmd, []string{})

	if _, err := os.Stat(untrackedFile); !os.IsNotExist(err) {
		t.Errorf("untracked.txt should have been deleted when force is true")
	}
}

func TestCleanAllCommand_NoForceNoDryRun(t *testing.T) {
	workspace, cleanup := setupLocalTestRepo(t)
	defer cleanup()

	originalDir, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(originalDir)
	}()
	if err := os.Chdir(workspace); err != nil {
		t.Fatalf("Failed to change to workspace directory: %v", err)
	}

	repoPath := filepath.Join(workspace, "repo1")
	untrackedFile := filepath.Join(repoPath, "untracked.txt")
	if err := os.WriteFile(untrackedFile, []byte("content"), 0o644); err != nil {
		t.Fatalf("Failed to create untracked file: %v", err)
	}

	cmd := setupLocalTestCommand()

	CleanAllCmd.Run(cmd, []string{})

	if _, err := os.Stat(untrackedFile); os.IsNotExist(err) {
		t.Errorf("untracked.txt should NOT have been deleted without --force or --dry-run")
	}
}
