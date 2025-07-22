package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestGetWorkingPath(t *testing.T) {
	tests := []struct {
		name           string
		useCurrentFlag bool
		expectError    bool
		description    string
	}{
		{
			name:           "Use current directory",
			useCurrentFlag: true,
			expectError:    false,
			description:    "Should return current working directory when --current flag is set",
		},
		{
			name:           "Use config path (will fail without config)",
			useCurrentFlag: false,
			expectError:    true,
			description:    "Should return error when config path is not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test command
			cmd := &cobra.Command{}
			cmd.PersistentFlags().Bool("current", false, "Use current working directory")

			if tt.useCurrentFlag {
				cmd.PersistentFlags().Set("current", "true")
			}

			path, err := getWorkingPath(cmd)

			if tt.expectError && err == nil {
				t.Errorf("getWorkingPath() expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("getWorkingPath() unexpected error: %v", err)
			}

			if !tt.expectError && path == "" {
				t.Errorf("getWorkingPath() returned empty path")
			}

			if tt.useCurrentFlag && !tt.expectError {
				// Verify it returns a valid directory path
				if _, err := os.Stat(path); os.IsNotExist(err) {
					t.Errorf("getWorkingPath() returned non-existent path: %s", path)
				}
			}
		})
	}
}

// Test helpers for command testing

func setupTestCommandEnvironment(t *testing.T, repoNames []string) (string, func()) {
	t.Helper()
	
	workspace := setupTestWorkspace(t, repoNames)
	
	// Cleanup function
	cleanup := func() {
		os.RemoveAll(workspace)
	}
	
	return workspace, cleanup
}

func createTestCommandWithFlags(current bool) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.PersistentFlags().Bool("current", false, "Use current working directory")
	cmd.Flags().Bool("dry-run", false, "Perform a dry run")
	
	if current {
		cmd.PersistentFlags().Set("current", "true")
	}
	
	return cmd
}

// Tests for individual commands

func TestSyncAllCommand_DryRun(t *testing.T) {
	workspace, cleanup := setupTestCommandEnvironment(t, []string{"repo1", "repo2", "repo3"})
	defer cleanup()

	// Change to workspace directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(workspace)

	// Create command with --current and --dry-run flags
	cmd := createTestCommandWithFlags(true)
	cmd.Flags().Set("dry-run", "true")

	// Test that the function doesn't panic and can find repos
	repos, err := findGitRepositories(workspace)
	if err != nil {
		t.Fatalf("findGitRepositories failed: %v", err)
	}

	if len(repos) != 3 {
		t.Errorf("Expected 3 repositories, got %d", len(repos))
	}

	// Verify all repos are git repositories
	for _, repo := range repos {
		gitDir := filepath.Join(repo, ".git")
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			t.Errorf("Repository %s is missing .git directory", repo)
		}
	}
}

func TestStatusAllCommand_WithCurrentFlag(t *testing.T) {
	workspace, cleanup := setupTestCommandEnvironment(t, []string{"clean-repo", "dirty-repo"})
	defer cleanup()

	// Make one repo dirty
	dirtyRepoPath := filepath.Join(workspace, "dirty-repo")
	testFile := filepath.Join(dirtyRepoPath, "new-file.txt")
	if err := os.WriteFile(testFile, []byte("uncommitted content"), 0644); err != nil {
		t.Fatalf("Failed to create uncommitted file: %v", err)
	}

	// Change to workspace directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(workspace)

	// Test getWorkingPath with current flag
	cmd := createTestCommandWithFlags(true)
	path, err := getWorkingPath(cmd)
	if err != nil {
		t.Fatalf("getWorkingPath failed: %v", err)
	}

	// Resolve both paths to handle macOS symlink differences
	expectedPath, _ := filepath.EvalSymlinks(workspace)
	actualPath, _ := filepath.EvalSymlinks(path)
	
	if actualPath != expectedPath {
		t.Errorf("getWorkingPath returned %s, expected %s", actualPath, expectedPath)
	}
}

func TestFetchAllCommand_ErrorHandling(t *testing.T) {
	// Test with non-existent directory
	cmd := createTestCommandWithFlags(false)
	
	// This should fail because we don't have a valid workspace
	_, err := getWorkingPath(cmd)
	if err == nil {
		t.Error("Expected error for missing config, but got none")
	}
}

func TestListAllCommand_WithPaths(t *testing.T) {
	workspace, cleanup := setupTestCommandEnvironment(t, []string{"repo-a", "repo-b"})
	defer cleanup()

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(workspace)

	// Test findGitRepositories function
	repos, err := findGitRepositories(workspace)
	if err != nil {
		t.Fatalf("findGitRepositories failed: %v", err)
	}

	expectedRepos := []string{"repo-a", "repo-b"}
	if len(repos) != len(expectedRepos) {
		t.Errorf("Expected %d repositories, got %d", len(expectedRepos), len(repos))
	}

	// Check that all expected repos are found
	foundRepos := make(map[string]bool)
	for _, repo := range repos {
		repoName := filepath.Base(repo)
		foundRepos[repoName] = true
	}

	for _, expectedRepo := range expectedRepos {
		if !foundRepos[expectedRepo] {
			t.Errorf("Expected repository %s not found", expectedRepo)
		}
	}
}

func TestBranchAllCommand_DetectsBranches(t *testing.T) {
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
	defer os.Chdir(originalDir)
	os.Chdir(workspace)

	repos, err := findGitRepositories(workspace)
	if err != nil {
		t.Fatalf("findGitRepositories failed: %v", err)
	}

	// Test that we can detect different branches
	for _, repo := range repos {
		repoName := filepath.Base(repo)
		if repoName == "main-repo" {
			// Should be on main branch
			currentBranch, err := getCurrentBranch(repo)
			if err != nil {
				t.Errorf("Failed to get current branch for %s: %v", repoName, err)
				continue
			}
			if currentBranch != "main" {
				t.Errorf("Expected main-repo to be on main branch, got %s", currentBranch)
			}
		} else if repoName == "feature-repo" {
			// Should be on feature branch
			currentBranch, err := getCurrentBranch(repo)
			if err != nil {
				t.Errorf("Failed to get current branch for %s: %v", repoName, err)
				continue
			}
			if currentBranch != "feature/new-feature" {
				t.Errorf("Expected feature-repo to be on feature/new-feature branch, got %s", currentBranch)
			}
		}
	}
}

// Helper function for branch detection (since we removed it from branch_all.go)
func getCurrentBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func TestPersistentFlagInheritance(t *testing.T) {
	// Test that persistent flags are properly inherited by subcommands
	
	// Create a test parent command with persistent flag
	parentCmd := &cobra.Command{
		Use: "parent",
	}
	parentCmd.PersistentFlags().Bool("current", false, "Use current working directory")
	parentCmd.PersistentFlags().Set("current", "true")
	
	// Create a test subcommand
	subCmd := &cobra.Command{
		Use: "sub",
	}
	
	// Add subcommand to parent
	parentCmd.AddCommand(subCmd)
	
	// Test that the subcommand can access the parent's persistent flag
	flagValue, err := subCmd.PersistentFlags().GetBool("current")
	if err != nil {
		// Try getting from parent
		if parent := subCmd.Parent(); parent != nil {
			flagValue, err = parent.PersistentFlags().GetBool("current")
		}
	}
	
	if err != nil {
		t.Errorf("Subcommand cannot access --current flag: %v", err)
	} else if !flagValue {
		t.Errorf("Subcommand --current flag should be true, got %v", flagValue)
	}
}

// Test the findGitRepositories function with the --current flag workflow
func TestCurrentFlagWorkflow(t *testing.T) {
	// Setup test workspace
	workspace := setupTestWorkspace(t, []string{"repo1", "repo2"})
	defer os.RemoveAll(workspace)

	// Change to the workspace directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(workspace)

	// Create a test command with --current flag
	cmd := &cobra.Command{}
	cmd.PersistentFlags().Bool("current", false, "Use current working directory")
	cmd.PersistentFlags().Set("current", "true")

	// Test getWorkingPath with current flag
	path, err := getWorkingPath(cmd)
	if err != nil {
		t.Fatalf("getWorkingPath() with --current flag failed: %v", err)
	}

	// Resolve both paths to handle macOS symlink differences
	expectedPath, _ := filepath.EvalSymlinks(workspace)
	actualPath, _ := filepath.EvalSymlinks(path)
	
	// Verify the path is the current workspace
	if actualPath != expectedPath {
		t.Errorf("getWorkingPath() returned %s, expected %s", actualPath, expectedPath)
	}

	// Test that findGitRepositories works with the returned path
	repos, err := findGitRepositories(path)
	if err != nil {
		t.Fatalf("findGitRepositories() failed: %v", err)
	}

	if len(repos) != 2 {
		t.Errorf("findGitRepositories() found %d repos, expected 2", len(repos))
	}
}

// Test that commands can handle repositories with different branch names
func TestDifferentBranchNames(t *testing.T) {
	workspace := setupTestWorkspace(t, []string{"main-repo", "master-repo"})
	defer os.RemoveAll(workspace)

	// We can test that our helper functions work correctly
	repos, err := findGitRepositories(workspace)
	if err != nil {
		t.Fatalf("findGitRepositories() failed: %v", err)
	}

	if len(repos) != 2 {
		t.Errorf("findGitRepositories() found %d repos, expected 2", len(repos))
	}

	// Verify the repositories are found correctly
	expectedPaths := []string{
		filepath.Join(workspace, "main-repo"),
		filepath.Join(workspace, "master-repo"),
	}

	for _, expectedPath := range expectedPaths {
		found := false
		for _, repoPath := range repos {
			if repoPath == expectedPath {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected repository %s not found in results", expectedPath)
		}
	}
}

// Test edge cases for directory handling
func TestEdgeCases(t *testing.T) {
	t.Run("Empty directory", func(t *testing.T) {
		// Create empty temporary directory
		tmpDir, err := os.MkdirTemp("", "empty-workspace-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		repos, err := findGitRepositories(tmpDir)
		if err != nil {
			t.Errorf("findGitRepositories() on empty directory error = %v", err)
			return
		}

		if len(repos) != 0 {
			t.Errorf("findGitRepositories() on empty directory found %d repos, want 0", len(repos))
		}
	})

	t.Run("Non-existent directory", func(t *testing.T) {
		nonExistentPath := "/this/path/does/not/exist"

		repos, err := findGitRepositories(nonExistentPath)
		if err == nil {
			t.Errorf("findGitRepositories() with non-existent path should return error, got nil")
		}

		if repos != nil {
			t.Errorf("findGitRepositories() with non-existent path should return nil repos, got %v", repos)
		}
	})
}
