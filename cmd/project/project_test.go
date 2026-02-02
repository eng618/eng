package project

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eng618/eng/utils/config"
	"github.com/eng618/eng/utils/log"
)

// setupTestEnvironment creates a temporary workspace and config for testing.
func setupTestEnvironment(t *testing.T) (workspacePath string, configPath string, cleanup func()) {
	t.Helper()

	// Create temporary workspace directory
	workspace, err := os.MkdirTemp("", "project-test-workspace-*")
	require.NoError(t, err)

	// Create temporary config directory
	configDir, err := os.MkdirTemp("", "project-test-config-*")
	require.NoError(t, err)

	configPath = filepath.Join(configDir, ".eng.yaml")

	// Reset viper and configure for test
	viper.Reset()
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// Set git.dev_path to our test workspace
	viper.Set("git.dev_path", workspace)

	// Create empty config file
	err = os.WriteFile(configPath, []byte(""), 0o644)
	require.NoError(t, err)

	cleanup = func() {
		viper.Reset()
		os.RemoveAll(workspace)
		os.RemoveAll(configDir)
	}

	return workspace, configPath, cleanup
}

// setupTestRepo creates a bare git repository that can be cloned.
func setupTestRepo(t *testing.T, name string) string {
	t.Helper()

	repoDir, err := os.MkdirTemp("", "test-bare-repo-*")
	require.NoError(t, err)

	repoPath := filepath.Join(repoDir, name+".git")
	err = os.MkdirAll(repoPath, 0o755)
	require.NoError(t, err)

	// Initialize bare repo
	cmd := exec.Command("git", "init", "--bare")
	cmd.Dir = repoPath
	err = cmd.Run()
	require.NoError(t, err)

	return repoPath
}

func TestProjectCmd_Help(t *testing.T) {
	var buf bytes.Buffer
	log.SetWriters(&buf, &buf)
	defer log.ResetWriters()

	ProjectCmd.SetOut(&buf)
	ProjectCmd.SetErr(&buf)

	// Run without subcommand should show help
	ProjectCmd.Run(ProjectCmd, []string{})

	// The command should complete without error
	// (help is shown via cmd.Help() call)
}

func TestProjectCmd_Info(t *testing.T) {
	var buf bytes.Buffer
	log.SetWriters(&buf, &buf)
	defer log.ResetWriters()

	workspace, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	ProjectCmd.SetOut(&buf)
	ProjectCmd.SetErr(&buf)

	// Set info flag
	err := ProjectCmd.Flags().Set("info", "true")
	require.NoError(t, err)
	defer func() {
		_ = ProjectCmd.Flags().Set("info", "false")
	}()

	ProjectCmd.Run(ProjectCmd, []string{})

	out := buf.String()
	assert.Contains(t, out, "Development Path:")
	assert.Contains(t, out, workspace)
	assert.Contains(t, out, "Configured Projects:")
}

func TestListCmd_EmptyProjects(t *testing.T) {
	var buf bytes.Buffer
	log.SetWriters(&buf, &buf)
	defer log.ResetWriters()

	_, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	ListCmd.SetOut(&buf)
	ListCmd.SetErr(&buf)

	ListCmd.Run(ListCmd, []string{})

	out := buf.String()
	assert.Contains(t, out, "No projects configured")
}

func TestListCmd_WithProjects(t *testing.T) {
	var buf bytes.Buffer
	log.SetWriters(&buf, &buf)
	defer log.ResetWriters()

	workspace, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Add test projects
	testProjects := []config.Project{
		{
			Name: "TestProject",
			Repos: []config.ProjectRepo{
				{URL: "git@github.com:org/repo1.git"},
				{URL: "git@github.com:org/repo2.git"},
			},
		},
	}
	err := config.SaveProjects(testProjects)
	require.NoError(t, err)

	// Create one repo directory to simulate partial clone
	repoDir := filepath.Join(workspace, "TestProject", "repo1", ".git")
	err = os.MkdirAll(repoDir, 0o755)
	require.NoError(t, err)

	ListCmd.SetOut(&buf)
	ListCmd.SetErr(&buf)

	ListCmd.Run(ListCmd, []string{})

	out := buf.String()
	assert.Contains(t, out, "TestProject")
	assert.Contains(t, out, "1/2 repos cloned")
}

func TestSetupCmd_NoDevPath(t *testing.T) {
	var buf bytes.Buffer
	log.SetWriters(&buf, &buf)
	defer log.ResetWriters()

	viper.Reset()
	// Don't set git.dev_path

	SetupCmd.SetOut(&buf)
	SetupCmd.SetErr(&buf)

	SetupCmd.Run(SetupCmd, []string{})

	out := buf.String()
	assert.Contains(t, out, "Development folder path is not set")
}

func TestSetupCmd_NoProjects(t *testing.T) {
	var buf bytes.Buffer
	log.SetWriters(&buf, &buf)
	defer log.ResetWriters()

	_, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	SetupCmd.SetOut(&buf)
	SetupCmd.SetErr(&buf)

	SetupCmd.Run(SetupCmd, []string{})

	out := buf.String()
	assert.Contains(t, out, "No projects configured")
}

func TestSetupCmd_DryRun(t *testing.T) {
	var buf bytes.Buffer
	log.SetWriters(&buf, &buf)
	defer log.ResetWriters()

	_, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Add test project
	testProjects := []config.Project{
		{
			Name: "DryRunProject",
			Repos: []config.ProjectRepo{
				{URL: "git@github.com:org/repo1.git"},
			},
		},
	}
	err := config.SaveProjects(testProjects)
	require.NoError(t, err)

	// Set dry-run flag on parent command (persistent flag)
	err = ProjectCmd.PersistentFlags().Set("dry-run", "true")
	require.NoError(t, err)
	defer func() {
		_ = ProjectCmd.PersistentFlags().Set("dry-run", "false")
	}()

	SetupCmd.SetOut(&buf)
	SetupCmd.SetErr(&buf)

	SetupCmd.Run(SetupCmd, []string{})

	out := buf.String()
	assert.Contains(t, out, "Dry run mode")
	assert.Contains(t, out, "[DRY RUN]")
	assert.Contains(t, out, "DryRunProject")
}

func TestSetupCmd_ProjectFilter_NotFound(t *testing.T) {
	var buf bytes.Buffer
	log.SetWriters(&buf, &buf)
	defer log.ResetWriters()

	_, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Add test project
	testProjects := []config.Project{
		{
			Name:  "ExistingProject",
			Repos: []config.ProjectRepo{{URL: "git@github.com:org/repo.git"}},
		},
	}
	err := config.SaveProjects(testProjects)
	require.NoError(t, err)

	// Set project filter to non-existent project (persistent flag on parent)
	err = ProjectCmd.PersistentFlags().Set("project", "NonExistent")
	require.NoError(t, err)
	defer func() {
		_ = ProjectCmd.PersistentFlags().Set("project", "")
	}()

	SetupCmd.SetOut(&buf)
	SetupCmd.SetErr(&buf)

	SetupCmd.Run(SetupCmd, []string{})

	out := buf.String()
	assert.Contains(t, out, "not found")
}

func TestSetupCmd_SkipsExistingRepos(t *testing.T) {
	var buf bytes.Buffer
	log.SetWriters(&buf, &buf)
	defer log.ResetWriters()

	workspace, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Add test project
	testProjects := []config.Project{
		{
			Name: "TestProject",
			Repos: []config.ProjectRepo{
				{URL: "git@github.com:org/existing-repo.git"},
			},
		},
	}
	err := config.SaveProjects(testProjects)
	require.NoError(t, err)

	// Create the repo directory with .git to simulate already cloned
	repoDir := filepath.Join(workspace, "TestProject", "existing-repo", ".git")
	err = os.MkdirAll(repoDir, 0o755)
	require.NoError(t, err)

	SetupCmd.SetOut(&buf)
	SetupCmd.SetErr(&buf)

	// Reset any lingering flags
	_ = SetupCmd.Flags().Set("dry-run", "false")
	_ = SetupCmd.Flags().Set("project", "")

	SetupCmd.Run(SetupCmd, []string{})

	out := buf.String()
	assert.Contains(t, out, "Already present: 1")
}

func TestFetchCmd_NoDevPath(t *testing.T) {
	var buf bytes.Buffer
	log.SetWriters(&buf, &buf)
	defer log.ResetWriters()

	viper.Reset()

	FetchCmd.SetOut(&buf)
	FetchCmd.SetErr(&buf)

	FetchCmd.Run(FetchCmd, []string{})

	out := buf.String()
	assert.Contains(t, out, "Development folder path is not set")
}

func TestPullCmd_NoDevPath(t *testing.T) {
	var buf bytes.Buffer
	log.SetWriters(&buf, &buf)
	defer log.ResetWriters()

	viper.Reset()

	PullCmd.SetOut(&buf)
	PullCmd.SetErr(&buf)

	PullCmd.Run(PullCmd, []string{})

	out := buf.String()
	assert.Contains(t, out, "Development folder path is not set")
}

func TestSyncCmd_NoDevPath(t *testing.T) {
	var buf bytes.Buffer
	log.SetWriters(&buf, &buf)
	defer log.ResetWriters()

	viper.Reset()

	SyncCmd.SetOut(&buf)
	SyncCmd.SetErr(&buf)

	SyncCmd.Run(SyncCmd, []string{})

	out := buf.String()
	assert.Contains(t, out, "Development folder path is not set")
}

func TestIsRepoCloned(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-is-cloned-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Test non-existent path
	assert.False(t, isRepoCloned(filepath.Join(tmpDir, "nonexistent")))

	// Test directory without .git
	noGitDir := filepath.Join(tmpDir, "no-git")
	err = os.MkdirAll(noGitDir, 0o755)
	require.NoError(t, err)
	assert.False(t, isRepoCloned(noGitDir))

	// Test directory with .git file (not directory)
	gitFileDir := filepath.Join(tmpDir, "git-file")
	err = os.MkdirAll(gitFileDir, 0o755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(gitFileDir, ".git"), []byte("gitdir: ../other"), 0o644)
	require.NoError(t, err)
	assert.False(t, isRepoCloned(gitFileDir)) // .git must be a directory

	// Test directory with .git directory
	gitDirPath := filepath.Join(tmpDir, "has-git")
	err = os.MkdirAll(filepath.Join(gitDirPath, ".git"), 0o755)
	require.NoError(t, err)
	assert.True(t, isRepoCloned(gitDirPath))
}

func TestCloneRepository_InvalidURL(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-clone-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	destPath := filepath.Join(tmpDir, "test-repo")

	// Try to clone with invalid URL - should fail with helpful error
	err = cloneRepository("not-a-valid-url", destPath)
	assert.Error(t, err)
}

func TestFetchCmd_DryRun(t *testing.T) {
	var buf bytes.Buffer
	log.SetWriters(&buf, &buf)
	defer log.ResetWriters()

	workspace, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Add test project
	testProjects := []config.Project{
		{
			Name: "FetchProject",
			Repos: []config.ProjectRepo{
				{URL: "git@github.com:org/repo1.git"},
			},
		},
	}
	err := config.SaveProjects(testProjects)
	require.NoError(t, err)

	// Create the repo directory with .git to simulate cloned
	repoDir := filepath.Join(workspace, "FetchProject", "repo1", ".git")
	err = os.MkdirAll(repoDir, 0o755)
	require.NoError(t, err)

	// Set dry-run flag on parent command (persistent flag)
	err = ProjectCmd.PersistentFlags().Set("dry-run", "true")
	require.NoError(t, err)
	defer func() {
		_ = ProjectCmd.PersistentFlags().Set("dry-run", "false")
	}()

	FetchCmd.SetOut(&buf)
	FetchCmd.SetErr(&buf)

	FetchCmd.Run(FetchCmd, []string{})

	out := buf.String()
	assert.Contains(t, out, "Dry run mode")
	assert.Contains(t, out, "[DRY RUN]")
}

func TestPullCmd_DryRun(t *testing.T) {
	var buf bytes.Buffer
	log.SetWriters(&buf, &buf)
	defer log.ResetWriters()

	workspace, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Add test project
	testProjects := []config.Project{
		{
			Name: "PullProject",
			Repos: []config.ProjectRepo{
				{URL: "git@github.com:org/repo1.git"},
			},
		},
	}
	err := config.SaveProjects(testProjects)
	require.NoError(t, err)

	// Create the repo directory with .git to simulate cloned
	repoDir := filepath.Join(workspace, "PullProject", "repo1", ".git")
	err = os.MkdirAll(repoDir, 0o755)
	require.NoError(t, err)

	// Set dry-run flag on parent command (persistent flag)
	err = ProjectCmd.PersistentFlags().Set("dry-run", "true")
	require.NoError(t, err)
	defer func() {
		_ = ProjectCmd.PersistentFlags().Set("dry-run", "false")
	}()

	PullCmd.SetOut(&buf)
	PullCmd.SetErr(&buf)

	PullCmd.Run(PullCmd, []string{})

	out := buf.String()
	assert.Contains(t, out, "Dry run mode")
	assert.Contains(t, out, "[DRY RUN]")
}

func TestSyncCmd_DryRun(t *testing.T) {
	var buf bytes.Buffer
	log.SetWriters(&buf, &buf)
	defer log.ResetWriters()

	workspace, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Add test project
	testProjects := []config.Project{
		{
			Name: "SyncProject",
			Repos: []config.ProjectRepo{
				{URL: "git@github.com:org/repo1.git"},
			},
		},
	}
	err := config.SaveProjects(testProjects)
	require.NoError(t, err)

	// Create the repo directory with .git to simulate cloned
	repoDir := filepath.Join(workspace, "SyncProject", "repo1", ".git")
	err = os.MkdirAll(repoDir, 0o755)
	require.NoError(t, err)

	// Set dry-run flag on parent command (persistent flag)
	err = ProjectCmd.PersistentFlags().Set("dry-run", "true")
	require.NoError(t, err)
	defer func() {
		_ = ProjectCmd.PersistentFlags().Set("dry-run", "false")
	}()

	SyncCmd.SetOut(&buf)
	SyncCmd.SetErr(&buf)

	SyncCmd.Run(SyncCmd, []string{})

	out := buf.String()
	assert.Contains(t, out, "Dry run mode")
	assert.Contains(t, out, "[DRY RUN]")
}

func TestListCmd_VerboseOutput(t *testing.T) {
	var buf bytes.Buffer
	log.SetWriters(&buf, &buf)
	defer log.ResetWriters()

	workspace, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Add test projects
	testProjects := []config.Project{
		{
			Name: "VerboseProject",
			Repos: []config.ProjectRepo{
				{URL: "git@github.com:org/repo1.git"},
				{URL: "git@github.com:org/repo2.git", Path: "custom-name"},
			},
		},
	}
	err := config.SaveProjects(testProjects)
	require.NoError(t, err)

	// Create one repo directory
	repoDir := filepath.Join(workspace, "VerboseProject", "repo1", ".git")
	err = os.MkdirAll(repoDir, 0o755)
	require.NoError(t, err)

	// Enable verbose via viper (simulating -v flag)
	viper.Set("verbose", true)
	defer viper.Set("verbose", false)

	ListCmd.SetOut(&buf)
	ListCmd.SetErr(&buf)

	ListCmd.Run(ListCmd, []string{})

	out := buf.String()
	assert.Contains(t, out, "VerboseProject")
	assert.Contains(t, out, "Repositories")
	assert.Contains(t, out, "git@github.com:org/repo1.git")
	assert.Contains(t, out, "Custom path: custom-name")
	// Check for status indicators
	assert.True(t, strings.Contains(out, "✓") || strings.Contains(out, "✗"))
}

func TestRemoveCmd_NoProjects(t *testing.T) {
	var buf bytes.Buffer
	log.SetWriters(&buf, &buf)
	defer log.ResetWriters()

	_, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	RemoveCmd.SetOut(&buf)
	RemoveCmd.SetErr(&buf)

	RemoveCmd.Run(RemoveCmd, []string{})

	out := buf.String()
	assert.Contains(t, out, "No projects configured")
}
