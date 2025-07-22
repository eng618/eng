package repo

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestRepo creates a temporary git repository for testing
func setupTestRepo(t *testing.T, branchName string) string {
	t.Helper()
	
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "test-repo-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		if removeErr := os.RemoveAll(tmpDir); removeErr != nil {
			t.Logf("Warning: failed to cleanup temp dir: %v", removeErr)
		}
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git user (required for commits)
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Logf("Warning: failed to set git user.name: %v", err)
	}
	
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Logf("Warning: failed to set git user.email: %v", err)
	}

	// Create initial commit
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Warning: failed to cleanup tmpDir: %v", err)
		}
		t.Fatalf("Failed to create test file: %v", err)
	}

	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Warning: failed to cleanup tmpDir: %v", err)
		}
		t.Fatalf("Failed to add file: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Warning: failed to cleanup tmpDir: %v", err)
		}
		t.Fatalf("Failed to commit: %v", err)
	}

	// Create and switch to specified branch if not main/master
	if branchName != "main" && branchName != "master" {
		cmd = exec.Command("git", "checkout", "-b", branchName)
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			if removeErr := os.RemoveAll(tmpDir); removeErr != nil {
				t.Logf("Warning: failed to cleanup temp dir: %v", removeErr)
			}
			t.Fatalf("Failed to create branch %s: %v", branchName, err)
		}
	} else if branchName == "master" {
		// If we want master, rename the default branch to master
		cmd = exec.Command("git", "branch", "-m", "main", "master")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			// Try renaming from whatever the default branch is
			cmd = exec.Command("git", "branch", "-m", "master")
			cmd.Dir = tmpDir
			if err := cmd.Run(); err != nil {
				if removeErr := os.RemoveAll(tmpDir); removeErr != nil {
					t.Logf("Warning: failed to cleanup temp dir: %v", removeErr)
				}
				t.Fatalf("Failed to rename branch to master: %v", err)
			}
		}
	}

	return tmpDir
}

// setupTestRepoWithBranches creates a test repo with multiple branches
func setupTestRepoWithBranches(t *testing.T, branches []string) string {
	t.Helper()
	
	tmpDir := setupTestRepo(t, "main")

	for _, branch := range branches {
		if branch == "main" || branch == "master" {
			continue // Skip if it's the default branch
		}
		
		cmd := exec.Command("git", "checkout", "-b", branch)
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			if removeErr := os.RemoveAll(tmpDir); removeErr != nil {
				t.Logf("Warning: failed to cleanup temp dir: %v", removeErr)
			}
			t.Fatalf("Failed to create branch %s: %v", branch, err)
		}

		// Go back to main
		cmd = exec.Command("git", "checkout", "main")
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			if removeErr := os.RemoveAll(tmpDir); removeErr != nil {
				t.Logf("Warning: failed to cleanup temp dir: %v", removeErr)
			}
			t.Fatalf("Failed to checkout main: %v", err)
		}
	}

	return tmpDir
}

func TestGetMainBranch(t *testing.T) {
	tests := []struct {
		name           string
		setupBranch    string
		expectedBranch string
	}{
		{
			name:           "Repository with main branch",
			setupBranch:    "main",
			expectedBranch: "main",
		},
		{
			name:           "Repository with master branch",
			setupBranch:    "master",
			expectedBranch: "master",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := setupTestRepo(t, tt.setupBranch)
			defer func() {
				if err := os.RemoveAll(tmpDir); err != nil {
					t.Logf("Warning: failed to cleanup tmpDir: %v", err)
				}
			}()

			branch, err := GetMainBranch(tmpDir)
			if err != nil {
				t.Errorf("GetMainBranch() error = %v", err)
				return
			}

			if branch != tt.expectedBranch {
				t.Errorf("GetMainBranch() = %v, want %v", branch, tt.expectedBranch)
			}
		})
	}
}

func TestGetDevelopBranch(t *testing.T) {
	tests := []struct {
		name           string
		branches       []string
		expectedBranch string
	}{
		{
			name:           "Repository with develop branch",
			branches:       []string{"develop"},
			expectedBranch: "develop",
		},
		{
			name:           "Repository with dev branch",
			branches:       []string{"dev"},
			expectedBranch: "dev",
		},
		{
			name:           "Repository with development branch",
			branches:       []string{"development"},
			expectedBranch: "development",
		},
		{
			name:           "Repository with multiple dev branches (prefer develop)",
			branches:       []string{"develop", "dev", "development"},
			expectedBranch: "develop",
		},
		{
			name:           "Repository with no dev branches",
			branches:       []string{},
			expectedBranch: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := setupTestRepoWithBranches(t, tt.branches)
			defer func() {
				if err := os.RemoveAll(tmpDir); err != nil {
					t.Logf("Warning: failed to cleanup tmpDir: %v", err)
				}
			}()

			branch, err := GetDevelopBranch(tmpDir)
			if err != nil {
				t.Errorf("GetDevelopBranch() error = %v", err)
				return
			}

			if branch != tt.expectedBranch {
				t.Errorf("GetDevelopBranch() = %v, want %v", branch, tt.expectedBranch)
			}
		})
	}
}

func TestGetCurrentBranch(t *testing.T) {
	tmpDir := setupTestRepoWithBranches(t, []string{"develop", "feature/test"})
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Warning: failed to cleanup tmpDir: %v", err)
		}
	}()

	// Test getting current branch when on main
	branch, err := GetCurrentBranch(tmpDir)
	if err != nil {
		t.Errorf("GetCurrentBranch() error = %v", err)
		return
	}
	if branch != "main" {
		t.Errorf("GetCurrentBranch() = %v, want %v", branch, "main")
	}

	// Switch to develop branch and test again
	cmd := exec.Command("git", "checkout", "develop")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to checkout develop: %v", err)
	}

	branch, err = GetCurrentBranch(tmpDir)
	if err != nil {
		t.Errorf("GetCurrentBranch() error = %v", err)
		return
	}
	if branch != "develop" {
		t.Errorf("GetCurrentBranch() = %v, want %v", branch, "develop")
	}
}

func TestBranchExists(t *testing.T) {
	tmpDir := setupTestRepoWithBranches(t, []string{"develop", "feature/test"})
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	tests := []struct {
		name       string
		branchName string
		expected   bool
	}{
		{
			name:       "Main branch exists",
			branchName: "main",
			expected:   true,
		},
		{
			name:       "Develop branch exists",
			branchName: "develop",
			expected:   true,
		},
		{
			name:       "Feature branch exists",
			branchName: "feature/test",
			expected:   true,
		},
		{
			name:       "Non-existent branch",
			branchName: "nonexistent",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := branchExists(tmpDir, tt.branchName)
			if result != tt.expected {
				t.Errorf("branchExists() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsDirty(t *testing.T) {
	tmpDir := setupTestRepo(t, "main")
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	// Test clean repository
	dirty, err := IsDirty(tmpDir)
	if err != nil {
		t.Errorf("IsDirty() error = %v", err)
		return
	}
	if dirty {
		t.Errorf("IsDirty() = %v, want %v for clean repo", dirty, false)
	}

	// Make repository dirty
	testFile := filepath.Join(tmpDir, "dirty.txt")
	if err := os.WriteFile(testFile, []byte("dirty content"), 0644); err != nil {
		t.Fatalf("Failed to create dirty file: %v", err)
	}

	// Test dirty repository
	dirty, err = IsDirty(tmpDir)
	if err != nil {
		t.Errorf("IsDirty() error = %v", err)
		return
	}
	if !dirty {
		t.Errorf("IsDirty() = %v, want %v for dirty repo", dirty, true)
	}
}

func TestEnsureOnDefaultBranch(t *testing.T) {
	tmpDir := setupTestRepoWithBranches(t, []string{"develop"})
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	// Switch to develop branch first
	cmd := exec.Command("git", "checkout", "develop")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to checkout develop: %v", err)
	}

	// Verify we're on develop
	currentBranch, err := GetCurrentBranch(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get current branch: %v", err)
	}
	if currentBranch != "develop" {
		t.Fatalf("Expected to be on develop, but on %s", currentBranch)
	}

	// Run EnsureOnDefaultBranch
	err = EnsureOnDefaultBranch(tmpDir)
	if err != nil {
		t.Errorf("EnsureOnDefaultBranch() error = %v", err)
		return
	}

	// Verify we're now on main
	currentBranch, err = GetCurrentBranch(tmpDir)
	if err != nil {
		t.Errorf("Failed to get current branch after EnsureOnDefaultBranch: %v", err)
		return
	}
	if currentBranch != "main" {
		t.Errorf("EnsureOnDefaultBranch() did not switch to main, still on %s", currentBranch)
	}
}

// Additional comprehensive tests for new functions

func TestGetMainBranch_Comprehensive(t *testing.T) {
	tests := []struct {
		name           string
		initialBranch  string
		additionalBranches []string
		expectedBranch string
	}{
		{
			name:           "Repository with main branch",
			initialBranch:  "main",
			additionalBranches: []string{},
			expectedBranch: "main",
		},
		{
			name:           "Repository with master branch",
			initialBranch:  "master",
			additionalBranches: []string{},
			expectedBranch: "master",
		},
		{
			name:           "Repository with both main and master - prefers main",
			initialBranch:  "master",
			additionalBranches: []string{"main"},
			expectedBranch: "main", // GetMainBranch prefers main over master
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoPath := setupTestRepo(t, tt.initialBranch)
			defer func() {
				_ = os.RemoveAll(repoPath)
			}()

			// Create additional branches if specified
			for _, branch := range tt.additionalBranches {
				// First check if we're trying to create a branch that already exists
				cmd := exec.Command("git", "rev-parse", "--verify", branch)
				cmd.Dir = repoPath
				if cmd.Run() == nil {
					// Branch already exists, skip creation
					t.Logf("Branch %s already exists, skipping creation", branch)
					continue
				}
				
				// Get current branch to return to it after creating the new branch
				currentBranch, err := getCurrentBranchInRepo(repoPath)
				if err != nil {
					t.Fatalf("Failed to get current branch: %v", err)
				}
				
				// Create the new branch from the current branch
				cmd = exec.Command("git", "checkout", "-b", branch)
				cmd.Dir = repoPath
				if err := cmd.Run(); err != nil {
					t.Fatalf("Failed to create branch %s: %v", branch, err)
				}
				
				// Return to the original branch
				cmd = exec.Command("git", "checkout", currentBranch)
				cmd.Dir = repoPath
				if err := cmd.Run(); err != nil {
					t.Fatalf("Failed to return to branch %s: %v", currentBranch, err)
				}
			}

			result, err := GetMainBranch(repoPath)
			if err != nil {
				t.Fatalf("GetMainBranch failed: %v", err)
			}

			if result != tt.expectedBranch {
				t.Errorf("GetMainBranch() = %v, want %v", result, tt.expectedBranch)
			}
		})
	}
}

// Helper function to get current branch for tests
func getCurrentBranchInRepo(repoPath string) (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func TestGetDevelopBranch_Comprehensive(t *testing.T) {
	tests := []struct {
		name           string
		branches       []string
		expectedBranch string
	}{
		{
			name:           "Repository with develop branch",
			branches:       []string{"develop"},
			expectedBranch: "develop",
		},
		{
			name:           "Repository with dev branch",
			branches:       []string{"dev"},
			expectedBranch: "dev",
		},
		{
			name:           "Repository with development branch",
			branches:       []string{"development"},
			expectedBranch: "development",
		},
		{
			name:           "Repository with multiple dev branches - prefers develop",
			branches:       []string{"dev", "develop", "development"},
			expectedBranch: "develop",
		},
		{
			name:           "Repository with no dev branches",
			branches:       []string{"feature/test"},
			expectedBranch: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoPath := setupTestRepo(t, "main")
			defer func() {
				_ = os.RemoveAll(repoPath)
			}()

			// Create test branches
			for _, branch := range tt.branches {
				cmd := exec.Command("git", "checkout", "-b", branch)
				cmd.Dir = repoPath
				if err := cmd.Run(); err != nil {
					t.Fatalf("Failed to create branch %s: %v", branch, err)
				}
			}

			result, err := GetDevelopBranch(repoPath)
			if err != nil {
				t.Fatalf("GetDevelopBranch failed: %v", err)
			}

			if result != tt.expectedBranch {
				t.Errorf("GetDevelopBranch() = %v, want %v", result, tt.expectedBranch)
			}
		})
	}
}

func TestBranchExists_Comprehensive(t *testing.T) {
	repoPath := setupTestRepo(t, "main")
	defer func() {
		_ = os.RemoveAll(repoPath)
	}()

	// Create test branches
	testBranches := []string{"test-branch", "feature/new-feature", "develop"}
	for _, branch := range testBranches {
		cmd := exec.Command("git", "checkout", "-b", branch)
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to create test branch %s: %v", branch, err)
		}
	}

	tests := []struct {
		name       string
		branchName string
		expected   bool
	}{
		{
			name:       "Existing main branch",
			branchName: "main",
			expected:   true,
		},
		{
			name:       "Existing test branch",
			branchName: "test-branch",
			expected:   true,
		},
		{
			name:       "Existing feature branch",
			branchName: "feature/new-feature",
			expected:   true,
		},
		{
			name:       "Existing develop branch",
			branchName: "develop",
			expected:   true,
		},
		{
			name:       "Non-existing branch",
			branchName: "non-existing",
			expected:   false,
		},
		{
			name:       "Non-existing feature branch",
			branchName: "feature/does-not-exist",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := branchExists(repoPath, tt.branchName)
			if result != tt.expected {
				t.Errorf("branchExists(%s) = %v, want %v", tt.branchName, result, tt.expected)
			}
		})
	}
}
