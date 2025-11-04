package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// setupTestWorkspace creates a temporary workspace with multiple git repositories
func setupTestWorkspace(tb testing.TB, repoNames []string) string {
	tb.Helper()

	// Create temporary workspace directory
	workspace, err := os.MkdirTemp("", "test-workspace-*")
	if err != nil {
		tb.Fatalf("Failed to create temp workspace: %v", err)
	}

	// Create multiple git repositories
	for _, repoName := range repoNames {
		repoPath := filepath.Join(workspace, repoName)
		if err := os.MkdirAll(repoPath, 0755); err != nil {
			_ = os.RemoveAll(workspace)
			tb.Fatalf("Failed to create repo directory %s: %v", repoName, err)
		}

		// Initialize git repo
		cmd := exec.Command("git", "init")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			_ = os.RemoveAll(workspace)
			tb.Fatalf("Failed to init git repo %s: %v", repoName, err)
		}

		// Configure git user (required for commits)
		cmd = exec.Command("git", "config", "user.name", "Test User")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			tb.Logf("Warning: failed to set git user.name: %v", err)
		}

		cmd = exec.Command("git", "config", "user.email", "test@example.com")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			tb.Logf("Warning: failed to set git user.email: %v", err)
		}

		// Disable commit signing for test repositories
		cmd = exec.Command("git", "config", "commit.gpgsign", "false")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			tb.Logf("Warning: failed to disable gpg signing: %v", err)
		}

		cmd = exec.Command("git", "config", "tag.gpgsign", "false")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			tb.Logf("Warning: failed to disable tag signing: %v", err)
		}

		// Create initial commit
		testFile := filepath.Join(repoPath, "README.md")
		content := "# " + repoName + "\n\nTest repository for " + repoName
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			_ = os.RemoveAll(workspace)
			tb.Fatalf("Failed to create test file in %s: %v", repoName, err)
		}

		cmd = exec.Command("git", "add", "README.md")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			_ = os.RemoveAll(workspace)
			tb.Fatalf("Failed to add file in %s: %v", repoName, err)
		}

		cmd = exec.Command("git", "commit", "-m", "Initial commit")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			_ = os.RemoveAll(workspace)
			tb.Fatalf("Failed to commit in %s: %v", repoName, err)
		}

		// Ensure we're on the main branch (rename default branch if needed)
		cmd = exec.Command("git", "branch", "-M", "main")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			tb.Logf("Warning: failed to rename branch to main in %s: %v", repoName, err)
		}
	}

	return workspace
}

// setupTestWorkspaceWithNonGitDirs creates a workspace with both git repos and non-git directories
func setupTestWorkspaceWithNonGitDirs(tb testing.TB, gitRepos []string, nonGitDirs []string) string {
	tb.Helper()

	workspace := setupTestWorkspace(tb, gitRepos)

	// Create non-git directories
	for _, dirName := range nonGitDirs {
		dirPath := filepath.Join(workspace, dirName)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			_ = os.RemoveAll(workspace)
			tb.Fatalf("Failed to create non-git directory %s: %v", dirName, err)
		}

		// Create a file in the directory to make it non-empty
		testFile := filepath.Join(dirPath, "file.txt")
		if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
			_ = os.RemoveAll(workspace)
			tb.Fatalf("Failed to create file in non-git dir %s: %v", dirName, err)
		}
	}

	return workspace
}

func TestFindGitRepositories(t *testing.T) {
	tests := []struct {
		name        string
		gitRepos    []string
		nonGitDirs  []string
		expectedLen int
	}{
		{
			name:        "Multiple git repositories",
			gitRepos:    []string{"repo1", "repo2", "repo3"},
			nonGitDirs:  []string{},
			expectedLen: 3,
		},
		{
			name:        "Mixed git repos and non-git directories",
			gitRepos:    []string{"project1", "project2"},
			nonGitDirs:  []string{"documents", "images"},
			expectedLen: 2,
		},
		{
			name:        "No git repositories",
			gitRepos:    []string{},
			nonGitDirs:  []string{"folder1", "folder2"},
			expectedLen: 0,
		},
		{
			name:        "Single git repository",
			gitRepos:    []string{"single-repo"},
			nonGitDirs:  []string{},
			expectedLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workspace := setupTestWorkspaceWithNonGitDirs(t, tt.gitRepos, tt.nonGitDirs)
			defer func() {
				_ = os.RemoveAll(workspace)
			}()

			repos, err := findGitRepositories(workspace)
			if err != nil {
				t.Errorf("findGitRepositories() error = %v", err)
				return
			}

			if len(repos) != tt.expectedLen {
				t.Errorf("findGitRepositories() found %d repos, want %d", len(repos), tt.expectedLen)
			}

			// Verify that all found repositories are actually git repositories
			for _, repoPath := range repos {
				gitDir := filepath.Join(repoPath, ".git")
				if _, err := os.Stat(gitDir); os.IsNotExist(err) {
					t.Errorf("Repository %s does not contain .git directory", repoPath)
				}
			}

			// Verify that all expected git repos are found
			for _, expectedRepo := range tt.gitRepos {
				expectedPath := filepath.Join(workspace, expectedRepo)
				found := false
				for _, foundRepo := range repos {
					if foundRepo == expectedPath {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected git repository %s not found in results", expectedRepo)
				}
			}
		})
	}
}

func TestFindGitRepositoriesNonExistentPath(t *testing.T) {
	nonExistentPath := "/this/path/does/not/exist"

	repos, err := findGitRepositories(nonExistentPath)
	if err == nil {
		t.Errorf("findGitRepositories() with non-existent path should return error, got nil")
	}

	if repos != nil {
		t.Errorf("findGitRepositories() with non-existent path should return nil repos, got %v", repos)
	}
}

func TestFindGitRepositoriesEmptyDirectory(t *testing.T) {
	// Create empty temporary directory
	tmpDir, err := os.MkdirTemp("", "empty-workspace-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	repos, err := findGitRepositories(tmpDir)
	if err != nil {
		t.Errorf("findGitRepositories() on empty directory error = %v", err)
		return
	}

	if len(repos) != 0 {
		t.Errorf("findGitRepositories() on empty directory found %d repos, want 0", len(repos))
	}
}

// Integration test that tests the workflow of git commands
func TestGitWorkflow(t *testing.T) {
	// Setup test workspace
	workspace := setupTestWorkspace(t, []string{"test-repo1", "test-repo2"})
	defer func() {
		_ = os.RemoveAll(workspace)
	}()

	// Test that we can find the repositories
	repos, err := findGitRepositories(workspace)
	if err != nil {
		t.Fatalf("Failed to find git repositories: %v", err)
	}

	if len(repos) != 2 {
		t.Fatalf("Expected 2 repositories, found %d", len(repos))
	}

	// Test that all repos are clean initially
	for _, repoPath := range repos {
		dirty, err := isDirty(repoPath)
		if err != nil {
			t.Errorf("Failed to check if repository %s is dirty: %v", repoPath, err)
			continue
		}

		if dirty {
			t.Errorf("Repository %s should be clean initially", repoPath)
		}
	}
}

// Helper function to check if repository is dirty (for testing)
func isDirty(repoPath string) (bool, error) {
	cmd := exec.Command("git", "-C", repoPath, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return len(output) > 0, nil
}

// Benchmark test for finding git repositories
func BenchmarkFindGitRepositories(b *testing.B) {
	// Setup workspace with multiple repositories
	repoNames := make([]string, 50) // Create 50 test repos
	for i := 0; i < 50; i++ {
		repoNames[i] = filepath.Join("repo", string(rune('A'+i%26)), string(rune('0'+i/26)))
	}

	workspace := setupTestWorkspace(b, repoNames)
	defer func() {
		_ = os.RemoveAll(workspace)
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := findGitRepositories(workspace)
		if err != nil {
			b.Fatalf("findGitRepositories() error = %v", err)
		}
	}
}
