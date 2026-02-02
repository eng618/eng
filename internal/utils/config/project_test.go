package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestConfig creates a temporary config file for testing.
func setupTestConfig(t *testing.T) (string, func()) {
	t.Helper()

	// Create a temporary directory for the config file
	tmpDir, err := os.MkdirTemp("", "project-config-test-*")
	require.NoError(t, err)

	configPath := filepath.Join(tmpDir, ".eng.yaml")

	// Reset viper and configure for test
	viper.Reset()
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// Create empty config file
	err = os.WriteFile(configPath, []byte(""), 0o644)
	require.NoError(t, err)

	cleanup := func() {
		viper.Reset()
		os.RemoveAll(tmpDir)
	}

	return configPath, cleanup
}

func TestRepoNameFromURL(t *testing.T) {
	testCases := []struct {
		name        string
		url         string
		expected    string
		expectError bool
	}{
		{
			name:     "SSH format with .git",
			url:      "git@github.com:org/my-repo.git",
			expected: "my-repo",
		},
		{
			name:     "SSH format without .git",
			url:      "git@github.com:org/my-repo",
			expected: "my-repo",
		},
		{
			name:     "SSH format with nested path",
			url:      "git@gitlab.com:group/subgroup/my-repo.git",
			expected: "my-repo",
		},
		{
			name:     "HTTPS format with .git",
			url:      "https://github.com/org/my-repo.git",
			expected: "my-repo",
		},
		{
			name:     "HTTPS format without .git",
			url:      "https://github.com/org/my-repo",
			expected: "my-repo",
		},
		{
			name:     "HTTPS format with nested path",
			url:      "https://gitlab.com/group/subgroup/my-repo.git",
			expected: "my-repo",
		},
		{
			name:     "HTTP format",
			url:      "http://github.com/org/my-repo.git",
			expected: "my-repo",
		},
		{
			name:        "Invalid format",
			url:         "not-a-valid-url",
			expectError: true,
		},
		{
			name:        "Empty string",
			url:         "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := RepoNameFromURL(tc.url)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestProjectRepo_GetEffectivePath(t *testing.T) {
	testCases := []struct {
		name        string
		repo        ProjectRepo
		expected    string
		expectError bool
	}{
		{
			name: "Custom path set",
			repo: ProjectRepo{
				URL:  "git@github.com:org/my-repo.git",
				Path: "custom-name",
			},
			expected: "custom-name",
		},
		{
			name: "No custom path - derive from URL",
			repo: ProjectRepo{
				URL:  "git@github.com:org/my-repo.git",
				Path: "",
			},
			expected: "my-repo",
		},
		{
			name: "Invalid URL with no custom path",
			repo: ProjectRepo{
				URL:  "invalid",
				Path: "",
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.repo.GetEffectivePath()

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestGetProjects_EmptyConfig(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	projects := GetProjects()
	assert.Empty(t, projects)
}

func TestSaveAndGetProjects(t *testing.T) {
	configPath, cleanup := setupTestConfig(t)
	defer cleanup()

	// Read the config file
	err := viper.ReadInConfig()
	if err != nil {
		// File might be empty, that's ok
		t.Logf("Note: %v", err)
	}

	// Create test projects
	testProjects := []Project{
		{
			Name: "ProjectA",
			Repos: []ProjectRepo{
				{URL: "git@github.com:org/repo1.git"},
				{URL: "git@github.com:org/repo2.git", Path: "custom"},
			},
		},
		{
			Name: "ProjectB",
			Repos: []ProjectRepo{
				{URL: "https://github.com/org/repo3.git"},
			},
		},
	}

	// Save projects
	err = SaveProjects(testProjects)
	require.NoError(t, err)

	// Verify file was written
	_, err = os.Stat(configPath)
	require.NoError(t, err)

	// Read back projects
	projects := GetProjects()
	assert.Len(t, projects, 2)
	assert.Equal(t, "ProjectA", projects[0].Name)
	assert.Len(t, projects[0].Repos, 2)
	assert.Equal(t, "ProjectB", projects[1].Name)
}

func TestGetProjectByName(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	testProjects := []Project{
		{Name: "Alpha", Repos: []ProjectRepo{{URL: "git@github.com:org/alpha.git"}}},
		{Name: "Beta", Repos: []ProjectRepo{{URL: "git@github.com:org/beta.git"}}},
	}

	err := SaveProjects(testProjects)
	require.NoError(t, err)

	// Test finding existing project
	project := GetProjectByName("Alpha")
	require.NotNil(t, project)
	assert.Equal(t, "Alpha", project.Name)

	// Test case-insensitive search
	project = GetProjectByName("alpha")
	require.NotNil(t, project)
	assert.Equal(t, "Alpha", project.Name)

	// Test non-existent project
	project = GetProjectByName("NonExistent")
	assert.Nil(t, project)
}

func TestAddProject(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	// Add first project
	project1 := Project{
		Name:  "First",
		Repos: []ProjectRepo{{URL: "git@github.com:org/first.git"}},
	}
	err := AddProject(project1)
	require.NoError(t, err)

	projects := GetProjects()
	assert.Len(t, projects, 1)

	// Add second project
	project2 := Project{
		Name:  "Second",
		Repos: []ProjectRepo{{URL: "git@github.com:org/second.git"}},
	}
	err = AddProject(project2)
	require.NoError(t, err)

	projects = GetProjects()
	assert.Len(t, projects, 2)

	// Update existing project (same name)
	project1Updated := Project{
		Name: "First",
		Repos: []ProjectRepo{
			{URL: "git@github.com:org/first.git"},
			{URL: "git@github.com:org/first-new.git"},
		},
	}
	err = AddProject(project1Updated)
	require.NoError(t, err)

	projects = GetProjects()
	assert.Len(t, projects, 2) // Still 2 projects
	first := GetProjectByName("First")
	require.NotNil(t, first)
	assert.Len(t, first.Repos, 2) // But first now has 2 repos
}

func TestAddRepoToProject(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	// Setup initial project
	project := Project{
		Name:  "TestProject",
		Repos: []ProjectRepo{{URL: "git@github.com:org/repo1.git"}},
	}
	err := AddProject(project)
	require.NoError(t, err)

	// Add repo to existing project
	newRepo := ProjectRepo{URL: "git@github.com:org/repo2.git", Path: "custom-path"}
	err = AddRepoToProject("TestProject", newRepo)
	require.NoError(t, err)

	p := GetProjectByName("TestProject")
	require.NotNil(t, p)
	assert.Len(t, p.Repos, 2)
	assert.Equal(t, "custom-path", p.Repos[1].Path)

	// Try to add duplicate repo
	err = AddRepoToProject("TestProject", newRepo)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Try to add to non-existent project
	err = AddRepoToProject("NonExistent", newRepo)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRemoveProject(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	// Setup projects
	testProjects := []Project{
		{Name: "Keep", Repos: []ProjectRepo{{URL: "git@github.com:org/keep.git"}}},
		{Name: "Remove", Repos: []ProjectRepo{{URL: "git@github.com:org/remove.git"}}},
	}
	err := SaveProjects(testProjects)
	require.NoError(t, err)

	// Remove project
	err = RemoveProject("Remove")
	require.NoError(t, err)

	projects := GetProjects()
	assert.Len(t, projects, 1)
	assert.Equal(t, "Keep", projects[0].Name)

	// Try to remove non-existent project
	err = RemoveProject("NonExistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRemoveRepoFromProject(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	// Setup project with multiple repos
	project := Project{
		Name: "TestProject",
		Repos: []ProjectRepo{
			{URL: "git@github.com:org/repo1.git"},
			{URL: "git@github.com:org/repo2.git"},
			{URL: "git@github.com:org/repo3.git"},
		},
	}
	err := AddProject(project)
	require.NoError(t, err)

	// Remove middle repo
	err = RemoveRepoFromProject("TestProject", "git@github.com:org/repo2.git")
	require.NoError(t, err)

	p := GetProjectByName("TestProject")
	require.NotNil(t, p)
	assert.Len(t, p.Repos, 2)
	assert.Equal(t, "git@github.com:org/repo1.git", p.Repos[0].URL)
	assert.Equal(t, "git@github.com:org/repo3.git", p.Repos[1].URL)

	// Try to remove non-existent repo
	err = RemoveRepoFromProject("TestProject", "git@github.com:org/nonexistent.git")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Try to remove from non-existent project
	err = RemoveRepoFromProject("NonExistent", "git@github.com:org/repo1.git")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetProjectNames(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	// Empty initially
	names := GetProjectNames()
	assert.Empty(t, names)

	// Add projects
	testProjects := []Project{
		{Name: "Zebra", Repos: []ProjectRepo{{URL: "git@github.com:org/z.git"}}},
		{Name: "Alpha", Repos: []ProjectRepo{{URL: "git@github.com:org/a.git"}}},
		{Name: "Beta", Repos: []ProjectRepo{{URL: "git@github.com:org/b.git"}}},
	}
	err := SaveProjects(testProjects)
	require.NoError(t, err)

	names = GetProjectNames()
	assert.Len(t, names, 3)
	// Names should be in the order they were saved
	assert.Equal(t, []string{"Zebra", "Alpha", "Beta"}, names)
}
