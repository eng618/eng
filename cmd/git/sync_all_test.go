package git

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestSyncAllCmd_DryRun(t *testing.T) {
	workspace, cleanup := setupTestCommandEnvironment(t, []string{"repo1", "repo2"})
	defer cleanup()

	// Change to workspace directory
	originalDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Warning: failed to restore original directory: %v", err)
		}
	}()
	if err := os.Chdir(workspace); err != nil {
		t.Fatalf("Failed to change to workspace directory: %v", err)
	}

	cmd := createTestCommandWithFlags(true)
	if cmd.Flags().Lookup("dry-run") == nil {
		cmd.Flags().Bool("dry-run", true, "Perform a dry run without making actual changes")
	}

	// Mock SyncAllCmd to test internal functionality while providing it with valid flag structure
	oldRun := SyncAllCmd.Run
	SyncAllCmd.Run = func(c *cobra.Command, args []string) {
		if err := cmd.Flags().Set("dry-run", "true"); err != nil {
			panic(err)
		}
		oldRun(cmd, args)
	}
	defer func() { SyncAllCmd.Run = oldRun }()

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	SyncAllCmd.Run(SyncAllCmd, []string{})
}

func TestSyncAllCmd_DirtyRepo(t *testing.T) {
	workspace, cleanup := setupTestCommandEnvironment(t, []string{"repo1", "repo2"})
	defer cleanup()

	// Make repo1 dirty
	dirtyRepoPath := filepath.Join(workspace, "repo1")
	testFile := filepath.Join(dirtyRepoPath, "new-file.txt")
	if err := os.WriteFile(testFile, []byte("uncommitted content"), 0o644); err != nil {
		t.Fatalf("Failed to create uncommitted file: %v", err)
	}

	originalDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Warning: failed to restore original directory: %v", err)
		}
	}()
	if err := os.Chdir(workspace); err != nil {
		t.Fatalf("Failed to change to workspace directory: %v", err)
	}

	cmd := createTestCommandWithFlags(true)
	if cmd.Flags().Lookup("dry-run") == nil {
		cmd.Flags().Bool("dry-run", false, "Perform a dry run without making actual changes")
	}

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	oldRun := SyncAllCmd.Run
	SyncAllCmd.Run = func(c *cobra.Command, args []string) {
		oldRun(cmd, args)
	}
	defer func() { SyncAllCmd.Run = oldRun }()

	SyncAllCmd.Run(SyncAllCmd, []string{})
}

func TestSyncAllCmd_CleanRepo(t *testing.T) {
	workspace, cleanup := setupTestCommandEnvironment(t, []string{"repo1"})
	defer cleanup()

	// Need a remote to pull from for clean pull. We can create a bare repo to act as remote
	remoteRepo := filepath.Join(workspace, "remote.git")
	if err := os.MkdirAll(remoteRepo, 0o755); err != nil {
		t.Fatalf("Failed to make remote dir: %v", err)
	}
	cmdInit := exec.Command("git", "init", "--bare")
	cmdInit.Dir = remoteRepo
	if err := cmdInit.Run(); err != nil {
		t.Fatalf("Failed to init bare remote: %v", err)
	}

	localRepo := filepath.Join(workspace, "repo1")

	// Add remote and push to it to set up tracking branch
	cmdRemote := exec.Command("git", "remote", "add", "origin", remoteRepo)
	cmdRemote.Dir = localRepo
	if err := cmdRemote.Run(); err != nil {
		t.Fatalf("Failed to add remote: %v", err)
	}

	cmdPush := exec.Command("git", "push", "-u", "origin", "main")
	cmdPush.Dir = localRepo
	if err := cmdPush.Run(); err != nil {
		t.Fatalf("Failed to push to remote: %v", err)
	}

	originalDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Warning: failed to restore original directory: %v", err)
		}
	}()
	if err := os.Chdir(workspace); err != nil {
		t.Fatalf("Failed to change to workspace directory: %v", err)
	}

	cmd := createTestCommandWithFlags(true)
	if cmd.Flags().Lookup("dry-run") == nil {
		cmd.Flags().Bool("dry-run", false, "Perform a dry run without making actual changes")
	}

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	oldRun := SyncAllCmd.Run
	SyncAllCmd.Run = func(c *cobra.Command, args []string) {
		oldRun(cmd, args)
	}
	defer func() { SyncAllCmd.Run = oldRun }()

	SyncAllCmd.Run(SyncAllCmd, []string{})
}

func TestSyncAllCmd_NonDefaultBranch(t *testing.T) {
	workspace, cleanup := setupTestCommandEnvironment(t, []string{"repo1"})
	defer cleanup()

	localRepo := filepath.Join(workspace, "repo1")

	// Need a remote so ensure on default branch has something to pull
	remoteRepo := filepath.Join(workspace, "remote.git")
	if err := os.MkdirAll(remoteRepo, 0o755); err != nil {
		t.Fatalf("Failed to make remote dir: %v", err)
	}
	cmdInit := exec.Command("git", "init", "--bare")
	cmdInit.Dir = remoteRepo
	if err := cmdInit.Run(); err != nil {
		t.Fatalf("Failed to init bare remote: %v", err)
	}

	cmdRemote := exec.Command("git", "remote", "add", "origin", remoteRepo)
	cmdRemote.Dir = localRepo
	if err := cmdRemote.Run(); err != nil {
		t.Fatalf("Failed to add remote: %v", err)
	}

	cmdPush := exec.Command("git", "push", "-u", "origin", "main")
	cmdPush.Dir = localRepo
	if err := cmdPush.Run(); err != nil {
		t.Fatalf("Failed to push to remote: %v", err)
	}

	// Create and switch to new branch
	cmdBranch := exec.Command("git", "checkout", "-b", "feature-branch")
	cmdBranch.Dir = localRepo
	if err := cmdBranch.Run(); err != nil {
		t.Fatalf("Failed to checkout branch: %v", err)
	}

	originalDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Warning: failed to restore original directory: %v", err)
		}
	}()
	if err := os.Chdir(workspace); err != nil {
		t.Fatalf("Failed to change to workspace directory: %v", err)
	}

	cmd := createTestCommandWithFlags(true)
	if cmd.Flags().Lookup("dry-run") == nil {
		cmd.Flags().Bool("dry-run", false, "Perform a dry run without making actual changes")
	}

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	oldRun := SyncAllCmd.Run
	SyncAllCmd.Run = func(c *cobra.Command, args []string) {
		oldRun(cmd, args)
	}
	defer func() { SyncAllCmd.Run = oldRun }()

	SyncAllCmd.Run(SyncAllCmd, []string{})

	// Verify we are back on main branch
	cmdCheck := exec.Command("git", "branch", "--show-current")
	cmdCheck.Dir = localRepo
	out, err := cmdCheck.Output()
	if err != nil {
		t.Fatalf("Failed to get current branch: %v", err)
	}
	if strings.TrimSpace(string(out)) != "main" {
		t.Errorf("Expected branch to be main, got %s", strings.TrimSpace(string(out)))
	}
}
