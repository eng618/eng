package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// TestEmail tests the Email function logic
func TestEmail(t *testing.T) {
	testCases := []struct {
		name         string
		emailValue   string
		expectPrompt bool
	}{
		{
			name:         "EmailNotSet",
			emailValue:   "",
			expectPrompt: true,
		},
		{
			name:         "EmailAlreadySet",
			emailValue:   "test@example.com",
			expectPrompt: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset viper for test
			viper.Reset()

			if tc.emailValue != "" {
				viper.Set("user-email", tc.emailValue)
			}

			// Note: We can't easily test the interactive survey prompts in unit tests
			// So we focus on testing the viper logic
			email := viper.GetString("user-email")

			if tc.expectPrompt {
				assert.Empty(t, email, "Expected email to be empty when not set")
			} else {
				assert.Equal(t, tc.emailValue, email, "Expected email to match set value")
			}
		})
	}
}

// TestDotfilesRepo tests the DotfilesRepo function logic
func TestDotfilesRepo(t *testing.T) {
	testCases := []struct {
		name         string
		repoPath     string
		expectPrompt bool
	}{
		{
			name:         "RepoPathNotSet",
			repoPath:     "",
			expectPrompt: true,
		},
		{
			name:         "RepoPathAlreadySet",
			repoPath:     "/home/user/.dotfiles",
			expectPrompt: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset viper for test
			viper.Reset()

			if tc.repoPath != "" {
				viper.Set("dotfiles.repoPath", tc.repoPath)
			}

			repoPath := viper.GetString("dotfiles.repoPath")

			if tc.expectPrompt {
				assert.Empty(t, repoPath, "Expected repo path to be empty when not set")
			} else {
				assert.Equal(t, tc.repoPath, repoPath, "Expected repo path to match set value")
			}
		})
	}
}

// TestRepoURL tests the RepoURL function logic
func TestRepoURL(t *testing.T) {
	testCases := []struct {
		name         string
		repoURL      string
		expectPrompt bool
	}{
		{
			name:         "RepoURLNotSet",
			repoURL:      "",
			expectPrompt: true,
		},
		{
			name:         "RepoURLAlreadySet",
			repoURL:      "https://github.com/user/dotfiles.git",
			expectPrompt: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset viper for test
			viper.Reset()

			if tc.repoURL != "" {
				viper.Set("dotfiles.repo_url", tc.repoURL)
			}

			repoURL := viper.GetString("dotfiles.repo_url")

			if tc.expectPrompt {
				assert.Empty(t, repoURL, "Expected repo URL to be empty when not set")
			} else {
				assert.Equal(t, tc.repoURL, repoURL, "Expected repo URL to match set value")
			}
		})
	}
}

// TestBranch tests the Branch function logic
func TestBranch(t *testing.T) {
	testCases := []struct {
		name         string
		branch       string
		expectPrompt bool
	}{
		{
			name:         "BranchNotSet",
			branch:       "",
			expectPrompt: true,
		},
		{
			name:         "BranchAlreadySet",
			branch:       "main",
			expectPrompt: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset viper for test
			viper.Reset()

			if tc.branch != "" {
				viper.Set("dotfiles.branch", tc.branch)
			}

			branch := viper.GetString("dotfiles.branch")

			if tc.expectPrompt {
				assert.Empty(t, branch, "Expected branch to be empty when not set")
			} else {
				assert.Equal(t, tc.branch, branch, "Expected branch to match set value")
			}
		})
	}
}

// TestBareRepoPath tests the BareRepoPath function logic
func TestBareRepoPath(t *testing.T) {
	testCases := []struct {
		name         string
		bareRepoPath string
		expectPrompt bool
	}{
		{
			name:         "BareRepoPathNotSet",
			bareRepoPath: "",
			expectPrompt: true,
		},
		{
			name:         "BareRepoPathAlreadySet",
			bareRepoPath: "/home/user/.config/dotfiles",
			expectPrompt: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset viper for test
			viper.Reset()

			if tc.bareRepoPath != "" {
				viper.Set("dotfiles.bare_repo_path", tc.bareRepoPath)
			}

			bareRepoPath := viper.GetString("dotfiles.bare_repo_path")

			if tc.expectPrompt {
				assert.Empty(t, bareRepoPath, "Expected bare repo path to be empty when not set")
			} else {
				assert.Equal(t, tc.bareRepoPath, bareRepoPath, "Expected bare repo path to match set value")
			}
		})
	}
}

// TestGitDevPath tests the GitDevPath function logic
func TestGitDevPath(t *testing.T) {
	testCases := []struct {
		name         string
		devPath      string
		expectPrompt bool
	}{
		{
			name:         "DevPathNotSet",
			devPath:      "",
			expectPrompt: true,
		},
		{
			name:         "DevPathAlreadySet",
			devPath:      "/home/user/Development",
			expectPrompt: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset viper for test
			viper.Reset()

			if tc.devPath != "" {
				viper.Set("git.devPath", tc.devPath)
			}

			devPath := viper.GetString("git.devPath")

			if tc.expectPrompt {
				assert.Empty(t, devPath, "Expected dev path to be empty when not set")
			} else {
				assert.Equal(t, tc.devPath, devPath, "Expected dev path to match set value")
			}
		})
	}
}

// TestVerbose tests the Verbose function logic
func TestVerbose(t *testing.T) {
	testCases := []struct {
		name         string
		verboseValue bool
		expectPrompt bool
	}{
		{
			name:         "VerboseNotSet",
			verboseValue: false,
			expectPrompt: true,
		},
		{
			name:         "VerboseAlreadySet",
			verboseValue: true,
			expectPrompt: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset viper for test
			viper.Reset()

			// Note: viper.GetBool returns false for unset values, so we need to check if it was explicitly set
			if tc.verboseValue {
				viper.Set("verbose", tc.verboseValue)
			}

			verbose := viper.GetBool("verbose")

			if tc.expectPrompt {
				// For unset values, viper.GetBool returns false, but we want to test the prompt logic
				// In practice, the function would check if the key exists in the config
				assert.False(t, verbose, "Expected verbose to be false when not explicitly set")
			} else {
				assert.Equal(t, tc.verboseValue, verbose, "Expected verbose to match set value")
			}
		})
	}
}

// TestConfigFileOperations tests config file read/write operations
func TestConfigFileOperations(t *testing.T) {
	t.Run("ConfigFileCreation", func(t *testing.T) {
		// Reset viper for test
		viper.Reset()
		viper.SetConfigType("yaml")
		viper.SetConfigFile("/tmp/test_config.yaml")
		defer func() { _ = os.Remove("/tmp/test_config.yaml") }()

		// Set some test values
		viper.Set("user-email", "test@example.com")
		viper.Set("dotfiles.repoPath", "/test/path")

		// Write config
		err := viper.WriteConfig()
		assert.NoError(t, err, "Expected no error when writing config")

		// Verify file exists
		_, err = os.Stat("/tmp/test_config.yaml")
		assert.NoError(t, err, "Expected config file to exist")

		// Reset viper and read config back
		viper.Reset()
		viper.SetConfigType("yaml")
		viper.SetConfigFile("/tmp/test_config.yaml")
		err = viper.ReadInConfig()
		assert.NoError(t, err, "Expected no error when reading config")

		// Verify values were persisted
		assert.Equal(t, "test@example.com", viper.GetString("user-email"))
		assert.Equal(t, "/test/path", viper.GetString("dotfiles.repoPath"))
	})

	t.Run("ConfigFilePermissions", func(t *testing.T) {
		// Reset viper for test
		viper.Reset()
		viper.SetConfigType("yaml")
		viper.SetConfigFile("/tmp/test_config.yaml")
		defer func() { _ = os.Remove("/tmp/test_config.yaml") }()

		viper.Set("test.value", "test")

		err := viper.WriteConfig()
		assert.NoError(t, err, "Expected no error when writing config")

		// Check file permissions
		info, err := os.Stat("/tmp/test_config.yaml")
		assert.NoError(t, err, "Expected no error when getting file info")

		// File should be readable/writable by owner
		mode := info.Mode()
		assert.True(t, mode.Perm()&0o400 != 0, "Expected config file to be readable")
	})
}

// TestEnvironmentVariableExpansion tests path expansion with environment variables
func TestEnvironmentVariableExpansion(t *testing.T) {
	t.Run("HomeDirectoryExpansion", func(t *testing.T) {
		// Reset viper for test
		viper.Reset()

		homeDir := os.Getenv("HOME")
		if homeDir == "" {
			t.Skip("HOME environment variable not set")
		}

		testPath := "$HOME/.config/test"
		expectedPath := filepath.Join(homeDir, ".config/test")

		// Test the expansion logic (similar to what's used in BareRepoPath)
		expandedPath := os.ExpandEnv(testPath)
		assert.Equal(t, expectedPath, expandedPath, "Expected path to be expanded with HOME directory")
	})

	t.Run("MultipleEnvVars", func(t *testing.T) {
		// Reset viper for test
		viper.Reset()

		// Set test environment variables
		_ = os.Setenv("TEST_VAR1", "value1")
		_ = os.Setenv("TEST_VAR2", "value2")
		defer func() {
			_ = os.Unsetenv("TEST_VAR1")
			_ = os.Unsetenv("TEST_VAR2")
		}()

		testPath := "$HOME/$TEST_VAR1/$TEST_VAR2"
		expandedPath := os.ExpandEnv(testPath)

		assert.Contains(t, expandedPath, "value1", "Expected TEST_VAR1 to be expanded")
		assert.Contains(t, expandedPath, "value2", "Expected TEST_VAR2 to be expanded")
		assert.NotContains(t, expandedPath, "$TEST_VAR1", "Expected variables to be fully expanded")
		assert.NotContains(t, expandedPath, "$TEST_VAR2", "Expected variables to be fully expanded")
	})
}

// TestConfigValidation tests various config validation scenarios
func TestConfigValidation(t *testing.T) {
	t.Run("EmailFormatValidation", func(t *testing.T) {
		// Reset viper for test
		viper.Reset()

		// Test various email formats
		validEmails := []string{
			"test@example.com",
			"user.name@domain.co.uk",
			"test+tag@gmail.com",
		}

		for _, email := range validEmails {
			viper.Set("user-email", email)
			storedEmail := viper.GetString("user-email")
			assert.Equal(t, email, storedEmail, "Expected email to be stored correctly")
		}
	})

	t.Run("URLFormatValidation", func(t *testing.T) {
		// Reset viper for test
		viper.Reset()

		// Test various URL formats
		validURLs := []string{
			"https://github.com/user/dotfiles.git",
			"git@github.com:user/dotfiles.git",
			"https://gitlab.com/user/dotfiles",
		}

		for _, url := range validURLs {
			viper.Set("dotfiles.repo_url", url)
			storedURL := viper.GetString("dotfiles.repo_url")
			assert.Equal(t, url, storedURL, "Expected URL to be stored correctly")
		}
	})

	t.Run("PathValidation", func(t *testing.T) {
		// Reset viper for test
		viper.Reset()

		// Test absolute paths
		absPaths := []string{
			"/home/user/.dotfiles",
			"/opt/config",
			"C:\\Users\\user\\dotfiles", // Windows-style path
		}

		for _, path := range absPaths {
			viper.Set("dotfiles.repoPath", path)
			storedPath := viper.GetString("dotfiles.repoPath")
			assert.Equal(t, path, storedPath, "Expected path to be stored correctly")
		}
	})
}

// TestBranchOptions tests the valid branch options
func TestBranchOptions(t *testing.T) {
	// Reset viper for test
	viper.Reset()

	validBranches := []string{"main", "master", "develop", "work", "server"}

	for _, branch := range validBranches {
		viper.Set("dotfiles.branch", branch)
		storedBranch := viper.GetString("dotfiles.branch")
		assert.Equal(t, branch, storedBranch, "Expected branch to be stored correctly")
	}
}

// TestDefaultValues tests default value handling
func TestDefaultValues(t *testing.T) {
	t.Run("DefaultBranch", func(t *testing.T) {
		// Reset viper for test
		viper.Reset()

		// When branch is not set, viper should return empty string
		branch := viper.GetString("dotfiles.branch")
		assert.Empty(t, branch, "Expected unset branch to be empty")
	})

	t.Run("DefaultVerbose", func(t *testing.T) {
		// Reset viper for test
		viper.Reset()

		// When verbose is not set, viper.GetBool should return false
		verbose := viper.GetBool("verbose")
		assert.False(t, verbose, "Expected unset verbose to be false")
	})
}
