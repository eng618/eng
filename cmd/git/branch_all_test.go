package git

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eng618/eng/internal/utils/log"
)

func TestBranchAllCmd_Run(t *testing.T) {
	workspace, cleanup := setupTestCommandEnvironment(t, []string{"main-repo", "feature-repo"})
	defer cleanup()

	// Create a feature branch in one repo
	featureRepoPath := filepath.Join(workspace, "feature-repo")
	cmd := exec.Command("git", "checkout", "-b", "feature/new-feature")
	cmd.Dir = featureRepoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create feature branch: %v", err)
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

	// Capture log output
	var out bytes.Buffer
	log.SetWriters(&out, &out)
	defer log.ResetWriters()

	// Run command
	cobraCmd := createTestCommandWithFlags(true)
	BranchAllCmd.Run(cobraCmd, []string{})

	output := out.String()

	// Verify output
	if !strings.Contains(output, "Checking current branch of all git repositories") {
		t.Errorf("Expected output to contain start message, got: %s", output)
	}
	if !strings.Contains(output, "Checking branches of 2 repositories:") {
		t.Errorf("Expected output to contain repo count message, got: %s", output)
	}
	if !strings.Contains(output, "main-repo: main") {
		t.Errorf("Expected output to contain main-repo branch, got: %s", output)
	}
	if !strings.Contains(output, "feature-repo: feature/new-feature") {
		t.Errorf("Expected output to contain feature-repo branch, got: %s", output)
	}
	if !strings.Contains(output, "Branch summary: 1 on default branch, 1 on other branches") {
		t.Errorf("Expected output to contain branch summary, got: %s", output)
	}
}

func TestBranchAllCmd_NoRepos(t *testing.T) {
	// Create an empty temporary directory
	workspace, err := os.MkdirTemp("", "empty-workspace-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(workspace)

	originalDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Warning: failed to restore original directory: %v", err)
		}
	}()
	if err := os.Chdir(workspace); err != nil {
		t.Fatalf("Failed to change to workspace directory: %v", err)
	}

	// Capture log output
	var out bytes.Buffer
	log.SetWriters(&out, &out)
	defer log.ResetWriters()

	// Run command
	cobraCmd := createTestCommandWithFlags(true)
	BranchAllCmd.Run(cobraCmd, []string{})

	output := out.String()

	// Verify output
	if !strings.Contains(output, "No git repositories found") {
		t.Errorf("Expected output to contain no repos warning, got: %s", output)
	}
}

func TestBranchAllCmd_NoWorkingPath(t *testing.T) {
	// Capture log output
	var out bytes.Buffer
	log.SetWriters(&out, &out)
	defer log.ResetWriters()

	// Create command without --current flag
	cobraCmd := createTestCommandWithFlags(false)

	BranchAllCmd.Run(cobraCmd, []string{})

	output := out.String()

	// Verify output contains the error message from getWorkingPath
	if !strings.Contains(output, "development folder path is not set") {
		t.Errorf("Expected output to contain path error, got: %s", output)
	}
}
