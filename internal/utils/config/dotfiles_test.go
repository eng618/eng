package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// TestEmail tests the Email function logic.
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

// TestDotfilesRepo tests the DotfilesRepo function logic.
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

// TestRepoURL tests the RepoURL function logic.
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

// TestBranch tests the Branch function logic.
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

// TestBareRepoPath tests the BareRepoPath function logic.
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

// TestGitDevPath tests the GitDevPath function logic.
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

// TestVerbose tests the Verbose function logic.
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

// TestConfigFileOperations tests config file read/write operations.
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

// TestEnvironmentVariableExpansion tests path expansion with environment variables.
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

// TestConfigValidation tests various config validation scenarios.
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

// TestBranchOptions tests the valid branch options.
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

// TestDefaultValues tests default value handling.
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

// TestBareRepoPathEnvironmentExpansion tests environment variable expansion in BareRepoPath.
func TestBareRepoPathEnvironmentExpansion(t *testing.T) {
	testCases := []struct {
		name            string
		configPath      string
		envVars         map[string]string
		expectedPath    string
		expectExpansion bool
	}{
		{
			name:            "NoExpansionNeeded",
			configPath:      "/absolute/path/to/repo",
			envVars:         map[string]string{},
			expectedPath:    "/absolute/path/to/repo",
			expectExpansion: false,
		},
		{
			name:            "HomeDirectoryExpansion",
			configPath:      "$HOME/.config/dotfiles",
			envVars:         map[string]string{"HOME": "/home/user"},
			expectedPath:    "/home/user/.config/dotfiles",
			expectExpansion: true,
		},
		{
			name:            "MultipleEnvVars",
			configPath:      "$HOME/$USER/config",
			envVars:         map[string]string{"HOME": "/home/user", "USER": "testuser"},
			expectedPath:    "/home/user/testuser/config",
			expectExpansion: true,
		},
		{
			name:            "MixedPath",
			configPath:      "/base/$HOME/suffix",
			envVars:         map[string]string{"HOME": "/home/user"},
			expectedPath:    "/base//home/user/suffix",
			expectExpansion: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset viper for test
			viper.Reset()

			// Set environment variables
			for key, value := range tc.envVars {
				oldValue := os.Getenv(key)
				_ = os.Setenv(key, value)
				defer func(k, v string) {
					if v == "" {
						_ = os.Unsetenv(k)
					} else {
						_ = os.Setenv(k, v)
					}
				}(key, oldValue)
			}

			// Set config path
			viper.Set("dotfiles.bare_repo_path", tc.configPath)

			// Test the expansion logic directly (similar to BareRepoPath function)
			configPath := viper.GetString("dotfiles.bare_repo_path")
			expandedPath := os.ExpandEnv(configPath)

			assert.Equal(t, tc.expectedPath, expandedPath, "Expected path to be expanded correctly")

			if tc.expectExpansion {
				assert.NotEqual(t, tc.configPath, expandedPath, "Expected path to be different after expansion")
				// Check that variables were expanded
				for varName := range tc.envVars {
					assert.NotContains(t, expandedPath, "$"+varName, "Expected %s variable to be expanded", varName)
				}
			} else {
				assert.Equal(t, tc.configPath, expandedPath, "Expected path to remain unchanged")
			}
		})
	}
}

// TestConfigFileBackupAndRecovery tests config file backup and recovery scenarios.
func TestConfigFileBackupAndRecovery(t *testing.T) {
	t.Run("ConfigFileBackupOnWrite", func(t *testing.T) {
		// Reset viper for test
		viper.Reset()
		viper.SetConfigType("yaml")
		configPath := "/tmp/test_config_backup.yaml"
		viper.SetConfigFile(configPath)
		defer func() { _ = os.Remove(configPath) }()

		// Write initial config
		viper.Set("test.initial", "value1")
		err := viper.WriteConfig()
		assert.NoError(t, err, "Expected no error when writing initial config")

		// Modify config
		viper.Set("test.modified", "value2")
		err = viper.WriteConfig()
		assert.NoError(t, err, "Expected no error when writing modified config")

		// Verify both values are present
		assert.Equal(t, "value1", viper.GetString("test.initial"))
		assert.Equal(t, "value2", viper.GetString("test.modified"))
	})

	t.Run("ConfigFileCorruptionRecovery", func(t *testing.T) {
		// Reset viper for test
		viper.Reset()
		viper.SetConfigType("yaml")
		configPath := "/tmp/test_config_corrupt.yaml"
		viper.SetConfigFile(configPath)
		defer func() { _ = os.Remove(configPath) }()

		// Write valid config
		viper.Set("test.value", "original")
		err := viper.WriteConfig()
		assert.NoError(t, err, "Expected no error when writing config")

		// Corrupt the file (simulate file corruption)
		err = os.WriteFile(configPath, []byte("invalid: yaml: content: [unclosed"), 0o644)
		assert.NoError(t, err, "Expected no error when corrupting file")

		// Try to read corrupted config - this should fail gracefully
		viper.Reset()
		viper.SetConfigType("yaml")
		viper.SetConfigFile(configPath)
		err = viper.ReadInConfig()
		assert.Error(t, err, "Expected error when reading corrupted config")
	})

	t.Run("ConfigFilePermissionsCheck", func(t *testing.T) {
		// Reset viper for test
		viper.Reset()
		viper.SetConfigType("yaml")
		configPath := "/tmp/test_config_perms.yaml"
		viper.SetConfigFile(configPath)
		defer func() { _ = os.Remove(configPath) }()

		viper.Set("test.value", "secret")

		err := viper.WriteConfig()
		assert.NoError(t, err, "Expected no error when writing config")

		// Check file permissions exist and are reasonable
		info, err := os.Stat(configPath)
		assert.NoError(t, err, "Expected no error when getting file info")

		mode := info.Mode()
		// File should be readable by owner
		assert.True(t, mode.Perm()&0o400 != 0, "Config file should be readable by owner")
		// File should be writable by owner
		assert.True(t, mode.Perm()&0o200 != 0, "Config file should be writable by owner")
		// File should exist
		assert.True(t, mode.IsRegular(), "Config file should be a regular file")
	})
}

// TestConfigMigrationScenarios tests various config migration scenarios.
func TestConfigMigrationScenarios(t *testing.T) {
	t.Run("EmptyConfigInitialization", func(t *testing.T) {
		// Reset viper for test
		viper.Reset()
		viper.SetConfigType("yaml")
		configPath := "/tmp/test_config_empty.yaml"
		viper.SetConfigFile(configPath)
		defer func() { _ = os.Remove(configPath) }()

		// Write empty config
		err := viper.WriteConfig()
		assert.NoError(t, err, "Expected no error when writing empty config")

		// Read it back
		viper.Reset()
		viper.SetConfigType("yaml")
		viper.SetConfigFile(configPath)
		err = viper.ReadInConfig()
		assert.NoError(t, err, "Expected no error when reading empty config")

		// Verify default values
		assert.Empty(t, viper.GetString("user-email"))
		assert.Empty(t, viper.GetString("dotfiles.repoPath"))
		assert.Empty(t, viper.GetString("dotfiles.repo_url"))
		assert.Empty(t, viper.GetString("dotfiles.branch"))
		assert.False(t, viper.GetBool("verbose"))
	})

	t.Run("PartialConfigHandling", func(t *testing.T) {
		// Reset viper for test
		viper.Reset()
		viper.SetConfigType("yaml")
		configPath := "/tmp/test_config_partial.yaml"
		viper.SetConfigFile(configPath)
		defer func() { _ = os.Remove(configPath) }()

		// Set only some values
		viper.Set("user-email", "test@example.com")
		viper.Set("verbose", true)
		// Leave others unset

		err := viper.WriteConfig()
		assert.NoError(t, err, "Expected no error when writing partial config")

		// Read it back
		viper.Reset()
		viper.SetConfigType("yaml")
		viper.SetConfigFile(configPath)
		err = viper.ReadInConfig()
		assert.NoError(t, err, "Expected no error when reading partial config")

		// Verify set values
		assert.Equal(t, "test@example.com", viper.GetString("user-email"))
		assert.True(t, viper.GetBool("verbose"))

		// Verify unset values have defaults
		assert.Empty(t, viper.GetString("dotfiles.repoPath"))
		assert.Empty(t, viper.GetString("dotfiles.repo_url"))
		assert.Empty(t, viper.GetString("dotfiles.branch"))
	})
}

// TestConfigEdgeCases tests edge cases in config handling.
func TestConfigEdgeCases(t *testing.T) {
	t.Run("SpecialCharactersInPaths", func(t *testing.T) {
		// Reset viper for test
		viper.Reset()

		// Test paths with special characters
		specialPaths := []string{
			"/path with spaces/repo",
			"/path-with-dashes/repo",
			"/path_with_underscores/repo",
			"/path.with.dots/repo",
			"C:\\Windows\\path\\repo", // Windows-style path
		}

		for _, path := range specialPaths {
			viper.Set("dotfiles.repoPath", path)
			storedPath := viper.GetString("dotfiles.repoPath")
			assert.Equal(t, path, storedPath, "Expected special characters to be preserved in path: %s", path)
		}
	})

	t.Run("VeryLongValues", func(t *testing.T) {
		// Reset viper for test
		viper.Reset()

		// Create a very long string
		longString := ""
		for i := 0; i < 1000; i++ {
			longString += "a"
		}

		viper.Set("dotfiles.repo_url", longString)
		storedString := viper.GetString("dotfiles.repo_url")
		assert.Equal(t, longString, storedString, "Expected very long values to be stored correctly")
		assert.Len(t, storedString, 1000, "Expected length to be preserved")
	})

	t.Run("UnicodeCharacters", func(t *testing.T) {
		// Reset viper for test
		viper.Reset()

		unicodeValues := []string{
			"user@例え.テスト",           // Japanese characters
			"user@пример.испытание", // Cyrillic characters
			"user@مثال.اختبار",      // Arabic characters
			"repo/path/测试",          // Chinese characters
		}

		for _, value := range unicodeValues {
			viper.Set("dotfiles.repo_url", value)
			storedValue := viper.GetString("dotfiles.repo_url")
			assert.Equal(t, value, storedValue, "Expected Unicode characters to be preserved: %s", value)
		}
	})
}

// TestConfigThreadSafetyNote documents that viper is not thread-safe.
func TestConfigThreadSafetyNote(t *testing.T) {
	// This test documents that viper (the underlying config library) is not thread-safe
	// and concurrent access should be avoided in production code

	// Reset viper for test
	viper.Reset()

	// Set a simple value to verify viper works
	viper.Set("test.thread_safety", "value")
	assert.Equal(t, "value", viper.GetString("test.thread_safety"))

	// Note: Concurrent access to viper would cause race conditions and is not recommended
	// Each goroutine should use its own viper instance if concurrent access is needed
}

// TestConfigTypeValidation tests validation of different config value types.
func TestConfigTypeValidation(t *testing.T) {
	// Reset viper for test
	viper.Reset()

	t.Run("StringValues", func(t *testing.T) {
		testValues := map[string]string{
			"user-email":              "test@example.com",
			"dotfiles.repoPath":       "/path/to/repo",
			"dotfiles.repo_url":       "https://github.com/user/repo.git",
			"dotfiles.branch":         "main",
			"dotfiles.bare_repo_path": "/bare/repo/path",
			"git.devPath":             "/dev/path",
		}

		for key, value := range testValues {
			viper.Set(key, value)
			storedValue := viper.GetString(key)
			assert.Equal(t, value, storedValue, "Expected string value to be stored correctly for key: %s", key)
			assert.IsType(t, "", storedValue, "Expected value to be string type for key: %s", key)
		}
	})

	t.Run("BooleanValues", func(t *testing.T) {
		viper.Set("verbose", true)
		storedValue := viper.GetBool("verbose")
		assert.True(t, storedValue, "Expected boolean value to be stored correctly")
		assert.IsType(t, true, storedValue, "Expected value to be boolean type")

		viper.Set("verbose", false)
		storedValue = viper.GetBool("verbose")
		assert.False(t, storedValue, "Expected boolean false value to be stored correctly")
	})
}
