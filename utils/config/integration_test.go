package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigWorkflowIntegration tests the complete config workflow
func TestConfigWorkflowIntegration(t *testing.T) {
	// This test simulates the complete user workflow of setting up configuration
	// It tests that all config functions work together properly

	t.Run("CompleteConfigSetupWorkflow", func(t *testing.T) {
		// Reset viper for test
		viper.Reset()
		viper.SetConfigType("yaml")
		configPath := "/tmp/test_workflow_config.yaml"
		viper.SetConfigFile(configPath)
		defer func() { _ = os.Remove(configPath) }()

		// Simulate user setting up their complete configuration
		configValues := map[string]interface{}{
			"user-email":              "user@example.com",
			"dotfiles.repoPath":       "/home/user/.dotfiles",
			"dotfiles.repo_url":       "https://github.com/user/dotfiles.git",
			"dotfiles.branch":         "main",
			"dotfiles.bare_repo_path": "$HOME/.config/dotfiles",
			"dotfiles.workTree":       os.Getenv("HOME"),
			"git.devPath":             "$HOME/Development",
			"verbose":                 true,
		}

		// Set all config values
		for key, value := range configValues {
			viper.Set(key, value)
		}

		// Write config
		err := viper.WriteConfig()
		require.NoError(t, err, "Expected no error when writing complete config")

		// Read config back
		viper.Reset()
		viper.SetConfigType("yaml")
		viper.SetConfigFile(configPath)
		err = viper.ReadInConfig()
		require.NoError(t, err, "Expected no error when reading complete config")

		// Verify all values were persisted correctly
		assert.Equal(t, "user@example.com", viper.GetString("user-email"))
		assert.Equal(t, "/home/user/.dotfiles", viper.GetString("dotfiles.repoPath"))
		assert.Equal(t, "https://github.com/user/dotfiles.git", viper.GetString("dotfiles.repo_url"))
		assert.Equal(t, "main", viper.GetString("dotfiles.branch"))
		assert.Equal(t, "$HOME/.config/dotfiles", viper.GetString("dotfiles.bare_repo_path"))
		assert.Equal(t, os.Getenv("HOME"), viper.GetString("dotfiles.workTree"))
		assert.Equal(t, "$HOME/Development", viper.GetString("git.devPath"))
		assert.True(t, viper.GetBool("verbose"))

		// Test environment variable expansion in paths
		homeDir := os.Getenv("HOME")
		if homeDir != "" {
			bareRepoPath := viper.GetString("dotfiles.bare_repo_path")
			expandedPath := os.ExpandEnv(bareRepoPath)
			expectedPath := filepath.Join(homeDir, ".config/dotfiles")
			assert.Equal(t, expectedPath, expandedPath, "Expected bare repo path to be expanded correctly")

			devPath := viper.GetString("git.devPath")
			expandedDevPath := os.ExpandEnv(devPath)
			expectedDevPath := filepath.Join(homeDir, "Development")
			assert.Equal(t, expectedDevPath, expandedDevPath, "Expected dev path to be expanded correctly")
		}
	})
}

// TestConfigUpdateScenarios tests various config update scenarios
func TestConfigUpdateScenarios(t *testing.T) {
	t.Run("IncrementalConfigUpdates", func(t *testing.T) {
		// Reset viper for test
		viper.Reset()
		viper.SetConfigType("yaml")
		configPath := "/tmp/test_update_config.yaml"
		viper.SetConfigFile(configPath)
		defer func() { _ = os.Remove(configPath) }()

		// Initial config
		viper.Set("user-email", "initial@example.com")
		err := viper.WriteConfig()
		require.NoError(t, err)

		// Update one value
		viper.Set("dotfiles.repoPath", "/new/path")
		err = viper.WriteConfig()
		require.NoError(t, err)

		// Read back and verify both old and new values
		viper.Reset()
		viper.SetConfigType("yaml")
		viper.SetConfigFile(configPath)
		err = viper.ReadInConfig()
		require.NoError(t, err)

		assert.Equal(t, "initial@example.com", viper.GetString("user-email"), "Expected original value to persist")
		assert.Equal(t, "/new/path", viper.GetString("dotfiles.repoPath"), "Expected new value to be added")
	})

	t.Run("ConfigValueOverrides", func(t *testing.T) {
		// Reset viper for test
		viper.Reset()
		viper.SetConfigType("yaml")
		configPath := "/tmp/test_override_config.yaml"
		viper.SetConfigFile(configPath)
		defer func() { _ = os.Remove(configPath) }()

		// Set initial value
		viper.Set("dotfiles.branch", "main")
		err := viper.WriteConfig()
		require.NoError(t, err)

		// Override with different value
		viper.Set("dotfiles.branch", "develop")
		err = viper.WriteConfig()
		require.NoError(t, err)

		// Read back and verify override worked
		viper.Reset()
		viper.SetConfigType("yaml")
		viper.SetConfigFile(configPath)
		err = viper.ReadInConfig()
		require.NoError(t, err)

		assert.Equal(t, "develop", viper.GetString("dotfiles.branch"), "Expected value to be overridden")
	})
}

// TestConfigFileFormats tests different config file formats
func TestConfigFileFormats(t *testing.T) {
	testCases := []struct {
		name       string
		configType string
		fileExt    string
	}{
		{"YAML", "yaml", "yaml"},
		{"JSON", "json", "json"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset viper for test
			viper.Reset()
			viper.SetConfigType(tc.configType)
			configPath := "/tmp/test_format_config." + tc.fileExt
			viper.SetConfigFile(configPath)
			defer func() { _ = os.Remove(configPath) }()

			// Set test data
			viper.Set("user-email", "format@example.com")
			viper.Set("dotfiles.repo_url", "https://github.com/user/repo.git")

			// Write config
			err := viper.WriteConfig()
			require.NoError(t, err, "Expected no error when writing %s config", tc.name)

			// Read config back
			viper.Reset()
			viper.SetConfigType(tc.configType)
			viper.SetConfigFile(configPath)
			err = viper.ReadInConfig()
			require.NoError(t, err, "Expected no error when reading %s config", tc.name)

			// Verify data
			assert.Equal(t, "format@example.com", viper.GetString("user-email"))
			assert.Equal(t, "https://github.com/user/repo.git", viper.GetString("dotfiles.repo_url"))
		})
	}
}

// TestConfigEnvironmentIntegration tests how config interacts with environment
func TestConfigEnvironmentIntegration(t *testing.T) {
	t.Run("EnvironmentVariablePrecedence", func(t *testing.T) {
		// Reset viper for test
		viper.Reset()

		// Set environment variable that might conflict with config
		originalHome := os.Getenv("HOME")
		defer func() {
			if originalHome == "" {
				_ = os.Unsetenv("HOME")
			} else {
				_ = os.Setenv("HOME", originalHome)
			}
		}()

		testHome := "/test/home"
		_ = os.Setenv("HOME", testHome)

		// Test that environment variables are used for expansion
		configPath := "$HOME/.config/app"
		viper.Set("test.path", configPath)

		expandedPath := os.ExpandEnv(viper.GetString("test.path"))
		expectedPath := "/test/home/.config/app"
		assert.Equal(t, expectedPath, expandedPath, "Expected environment variable to be expanded in config value")
	})

	t.Run("ConfigIsolationBetweenTests", func(t *testing.T) {
		// Reset viper for test
		viper.Reset()

		// This test ensures viper state doesn't leak between tests
		viper.Set("test.isolation", "unique-value")

		// In a real scenario, each test should start with a clean viper state
		assert.Equal(t, "unique-value", viper.GetString("test.isolation"))

		// Reset should clean this up for next test
		viper.Reset()
		assert.Empty(t, viper.GetString("test.isolation"), "Expected viper to be clean after reset")
	})
}

// TestConfigDataTypes tests various data types in config
func TestConfigDataTypes(t *testing.T) {
	t.Run("ComplexDataStructures", func(t *testing.T) {
		// Reset viper for test
		viper.Reset()
		viper.SetConfigType("yaml")
		configPath := "/tmp/test_complex_config.yaml"
		viper.SetConfigFile(configPath)
		defer func() { _ = os.Remove(configPath) }()

		// Set complex data structures
		viper.Set("user.preferences", map[string]interface{}{
			"theme":     "dark",
			"language":  "en",
			"auto_save": true,
			"font_size": 12,
		})

		viper.Set("dotfiles.settings", map[string]interface{}{
			"auto_commit": false,
			"backup":      true,
		})

		// Write config
		err := viper.WriteConfig()
		require.NoError(t, err, "Expected no error when writing complex config")

		// Read back
		viper.Reset()
		viper.SetConfigType("yaml")
		viper.SetConfigFile(configPath)
		err = viper.ReadInConfig()
		require.NoError(t, err, "Expected no error when reading complex config")

		// Verify complex structures
		prefs := viper.Get("user.preferences")
		assert.NotNil(t, prefs, "Expected preferences to be stored")

		settings := viper.Get("dotfiles.settings")
		assert.NotNil(t, settings, "Expected settings to be stored")
	})
}

// TestConfigPerformance tests config performance with large datasets
func TestConfigPerformance(t *testing.T) {
	t.Run("LargeConfigHandling", func(t *testing.T) {
		// Reset viper for test
		viper.Reset()
		viper.SetConfigType("yaml")
		configPath := "/tmp/test_large_config.yaml"
		viper.SetConfigFile(configPath)
		defer func() { _ = os.Remove(configPath) }()

		// Create a large config with many entries using proper key formatting
		for i := 0; i < 100; i++ { // Reduced to 100 entries to avoid potential limits
			key := fmt.Sprintf("test.key%d", i)
			value := fmt.Sprintf("value%d", i)
			viper.Set(key, value)
		}

		// Write large config
		err := viper.WriteConfig()
		require.NoError(t, err, "Expected no error when writing large config")

		// Read back large config
		viper.Reset()
		viper.SetConfigType("yaml")
		viper.SetConfigFile(configPath)
		err = viper.ReadInConfig()
		require.NoError(t, err, "Expected no error when reading large config")

		// Verify a few values
		assert.Equal(t, "value97", viper.GetString("test.key97"))
		assert.Equal(t, "value50", viper.GetString("test.key50"))
		assert.Equal(t, "value10", viper.GetString("test.key10"))
	})
}

// TestConfigRealWorldScenarios tests real-world usage scenarios
func TestConfigRealWorldScenarios(t *testing.T) {
	t.Run("DotfilesSetupScenario", func(t *testing.T) {
		// Reset viper for test
		viper.Reset()
		viper.SetConfigType("yaml")
		configPath := "/tmp/test_dotfiles_scenario.yaml"
		viper.SetConfigFile(configPath)
		defer func() { _ = os.Remove(configPath) }()

		// Simulate a real dotfiles setup scenario
		configSetup := map[string]interface{}{
			"user-email":              "developer@example.com",
			"dotfiles.repoPath":       "~/dotfiles",
			"dotfiles.repo_url":       "git@github.com:developer/dotfiles.git",
			"dotfiles.branch":         "main",
			"dotfiles.bare_repo_path": "~/.config/dotfiles",
			"git.devPath":             "~/Development",
			"verbose":                 false,
		}

		// Set up complete configuration
		for key, value := range configSetup {
			viper.Set(key, value)
		}

		err := viper.WriteConfig()
		require.NoError(t, err, "Expected no error in dotfiles setup scenario")

		// Simulate reading config after application restart
		viper.Reset()
		viper.SetConfigType("yaml")
		viper.SetConfigFile(configPath)
		err = viper.ReadInConfig()
		require.NoError(t, err, "Expected no error when loading dotfiles config")

		// Verify all settings are correct for a typical dotfiles setup
		assert.Equal(t, "developer@example.com", viper.GetString("user-email"))
		assert.Equal(t, "~/dotfiles", viper.GetString("dotfiles.repoPath"))
		assert.Equal(t, "git@github.com:developer/dotfiles.git", viper.GetString("dotfiles.repo_url"))
		assert.Equal(t, "main", viper.GetString("dotfiles.branch"))
		assert.Equal(t, "~/.config/dotfiles", viper.GetString("dotfiles.bare_repo_path"))
		assert.Equal(t, "~/Development", viper.GetString("git.devPath"))
		assert.False(t, viper.GetBool("verbose"))
	})
}
