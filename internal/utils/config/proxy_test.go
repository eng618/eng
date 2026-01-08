package config

import (
	"errors"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

const (
	testConfigPath    = "/tmp/test_config.json"
	invalidConfigPath = "/invalid/path/test_config.json"
)

var (
	// Standard test proxies to reduce duplication.
	testProxy1 = ProxyConfig{Title: "Proxy1", Value: "http://proxy1:8080", Enabled: false, NoProxy: ""}
	testProxy2 = ProxyConfig{Title: "Proxy2", Value: "http://proxy2:8080", Enabled: true, NoProxy: "internal.example.com"}

	// Environment variables that get modified during tests.
	proxyEnvVars = []string{
		"ALL_PROXY", "HTTP_PROXY", "HTTPS_PROXY", "GLOBAL_AGENT_HTTP_PROXY",
		"NO_PROXY", "http_proxy", "https_proxy", "no_proxy",
	}
)

// TestMain handles global setup/teardown for all tests.
func TestMain(m *testing.M) {
	// Run tests
	exitCode := m.Run()

	// Always clean up environment variables after tests
	cleanupEnvVars()

	os.Exit(exitCode)
}

// Helper function to set up viper for testing.
func setupViper(configPath string) {
	viper.Reset()
	viper.SetConfigType("json")
	viper.SetConfigFile(configPath)
}

// Helper function to setup proxies configuration.
func setupProxies(proxies []ProxyConfig) {
	viper.Set("proxies", proxies)
}

// Helper function to clean up environment variables.
func cleanupEnvVars() {
	for _, envVar := range proxyEnvVars {
		os.Unsetenv(envVar)
	}
}

// Helper to set environment variables for testing.
func setTestEnvVars(proxyValue string) {
	for _, envVar := range proxyEnvVars[:6] { // All except NO_PROXY vars
		os.Setenv(envVar, proxyValue)
	}
	os.Setenv("NO_PROXY", "localhost,127.0.0.1")
	os.Setenv("no_proxy", "localhost,127.0.0.1")
}

// Group 1: GetProxyConfigs tests.
func TestGetProxyConfigs(t *testing.T) {
	testCases := []struct {
		name            string
		setup           func()
		expectedProxies int
		expectedActive  int
	}{
		{
			name: "NoProxiesSet",
			setup: func() {
				setupViper(testConfigPath)
				viper.Set("proxies", nil)
			},
			expectedProxies: 0,
			expectedActive:  -1,
		},
		{
			name: "LegacyProxyMigration",
			setup: func() {
				setupViper(testConfigPath)
				viper.Set("proxy.value", "http://legacy-proxy:8080")
				viper.Set("proxy.enabled", true)
			},
			expectedProxies: 1,
			expectedActive:  0,
		},
		{
			name: "ExistingProxies",
			setup: func() {
				setupViper(testConfigPath)
				setupProxies([]ProxyConfig{testProxy1, testProxy2})
			},
			expectedProxies: 2,
			expectedActive:  1,
		},
		{
			name: "NoActiveProxy",
			setup: func() {
				setupViper(testConfigPath)
				setupProxies([]ProxyConfig{
					{Title: "Proxy1", Value: "http://proxy1:8080", Enabled: false, NoProxy: ""},
					{Title: "Proxy2", Value: "http://proxy2:8080", Enabled: false, NoProxy: ""},
				})
			},
			expectedProxies: 2,
			expectedActive:  -1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup()
			proxies, activeIndex := GetProxyConfigs()
			assert.Equal(t, tc.expectedProxies, len(proxies), "Expected proxy count to match")
			assert.Equal(t, tc.expectedActive, activeIndex, "Expected active index to match")

			if tc.name == "LegacyProxyMigration" && len(proxies) > 0 {
				assert.Equal(t, "Default", proxies[0].Title, "Expected default title for migrated proxy")
				assert.Equal(t, "http://legacy-proxy:8080", proxies[0].Value, "Expected legacy proxy value to be migrated")
				assert.True(t, proxies[0].Enabled, "Expected migrated proxy to be enabled")
			}
		})
	}
}

// Group 2: GetActiveProxy tests.
func TestGetActiveProxy(t *testing.T) {
	testCases := []struct {
		name         string
		setup        func()
		expectedVal  string
		expectedBool bool
	}{
		{
			name: "NoProxies",
			setup: func() {
				setupViper(testConfigPath)
				viper.Set("proxies", nil)
			},
			expectedVal:  "",
			expectedBool: false,
		},
		{
			name: "ActiveProxyExists",
			setup: func() {
				setupViper(testConfigPath)
				setupProxies([]ProxyConfig{testProxy1, testProxy2})
			},
			expectedVal:  testProxy2.Value,
			expectedBool: true,
		},
		{
			name: "NoActiveProxy",
			setup: func() {
				setupViper(testConfigPath)
				setupProxies([]ProxyConfig{
					{Title: "Proxy1", Value: "http://proxy1:8080", Enabled: false, NoProxy: ""},
					{Title: "Proxy2", Value: "http://proxy2:8080", Enabled: false, NoProxy: ""},
				})
			},
			expectedVal:  "http://proxy1:8080",
			expectedBool: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup()
			value, isActive := GetActiveProxy()
			assert.Equal(t, tc.expectedVal, value, "Expected active proxy value to match")
			assert.Equal(t, tc.expectedBool, isActive, "Expected active proxy state to match")
		})
	}
}

// Group 3: SaveProxyConfigs tests.
func TestSaveProxyConfigs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		setupViper(testConfigPath)
		proxies := []ProxyConfig{testProxy1, testProxy2}

		err := SaveProxyConfigs(proxies)
		assert.NoError(t, err, "Expected no error when saving proxies")

		var savedProxies []ProxyConfig
		err = viper.UnmarshalKey("proxies", &savedProxies)
		assert.NoError(t, err, "Expected no error when unmarshaling saved proxies")
		assert.Equal(t, proxies, savedProxies, "Expected saved proxies to match input proxies")
	})

	t.Run("WriteConfigError", func(t *testing.T) {
		setupViper(invalidConfigPath)
		err := SaveProxyConfigs([]ProxyConfig{testProxy1})
		assert.Error(t, err, "Expected an error when saving to invalid path")
		assert.Contains(t, err.Error(), "Error writing config file", "Expected error message to indicate write config failure")
	})
}

// Mock function for saving configs that simulates a failure.
func mockSaveProxyConfigsFailure(proxies []ProxyConfig) error {
	return errors.New("Error writing config file: mock error")
}

// Group 4: EnableProxy tests.
func TestEnableProxy(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		setupViper(testConfigPath)
		proxies := []ProxyConfig{
			testProxy1,
			{Title: "Proxy2", Value: "http://proxy2:8080", Enabled: false, NoProxy: ""},
		}

		updatedProxies, err := EnableProxy(1, proxies)
		assert.NoError(t, err, "Expected no error when enabling a proxy")
		assert.False(t, updatedProxies[0].Enabled, "Expected first proxy to be disabled")
		assert.True(t, updatedProxies[1].Enabled, "Expected second proxy to be enabled")
	})

	t.Run("IndexOutOfRange", func(t *testing.T) {
		setupViper(testConfigPath)
		proxies := []ProxyConfig{testProxy1}

		_, err := EnableProxy(5, proxies)
		assert.Error(t, err, "Expected error for out-of-range index")
		assert.Equal(t, "proxy index out of range", err.Error(), "Expected specific error message")
	})

	// Use a temporary function override for testing SaveProxyConfigs failure
	t.Run("SaveConfigError", func(t *testing.T) {
		setupViper(testConfigPath) // Use valid path but mock the save function
		proxies := []ProxyConfig{
			testProxy1,
			{Title: "Proxy2", Value: "http://proxy2:8080", Enabled: false, NoProxy: ""},
		}

		// Store original function
		originalSaveFunc := SaveProxyConfigs
		// Override with mock
		SaveProxyConfigs = mockSaveProxyConfigsFailure

		// Test with mocked function
		_, err := EnableProxy(1, proxies)
		assert.Error(t, err, "Expected error when saving")
		assert.Contains(t, err.Error(), "Error writing config file", "Expected error message for config write failure")

		// Restore original function
		SaveProxyConfigs = originalSaveFunc
	})
}

// Group 5: DisableAllProxies tests.
func TestDisableAllProxies(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		setupViper(testConfigPath)
		setupProxies([]ProxyConfig{
			{Title: "Proxy1", Value: "http://proxy1:8080", Enabled: true, NoProxy: ""},
			{Title: "Proxy2", Value: "http://proxy2:8080", Enabled: false, NoProxy: ""},
		})

		err := DisableAllProxies()
		assert.NoError(t, err, "Expected no error when disabling all proxies")

		// Verify proxies are disabled
		updatedProxies, _ := GetProxyConfigs()
		for _, proxy := range updatedProxies {
			assert.False(t, proxy.Enabled, "Expected all proxies to be disabled")
		}

		// Check environment variables are unset
		for _, envVar := range proxyEnvVars {
			value := os.Getenv(envVar)
			assert.Empty(t, value, "Expected environment variable %s to be unset", envVar)
		}
	})

	t.Run("SaveConfigError", func(t *testing.T) {
		setupViper(testConfigPath) // Use valid path but mock the save function
		setupProxies([]ProxyConfig{{Title: "Proxy1", Value: "http://proxy1:8080", Enabled: true, NoProxy: ""}})

		// Store original function
		originalSaveFunc := SaveProxyConfigs
		// Override with mock
		SaveProxyConfigs = mockSaveProxyConfigsFailure

		err := DisableAllProxies()
		assert.Error(t, err, "Expected error for invalid config path")
		assert.Contains(t, err.Error(), "Error writing config file", "Expected config write error message")

		// Restore original function
		SaveProxyConfigs = originalSaveFunc
	})
}

// Group 6: Environment Variable Management tests.
func TestProxyEnvironmentVariables(t *testing.T) {
	t.Run("UnsetProxyEnvVars", func(t *testing.T) {
		// Setup env vars
		setTestEnvVars("http://proxy1:8080")

		// Call function
		UnsetProxyEnvVars()

		// Verify all are unset
		for _, envVar := range proxyEnvVars {
			value := os.Getenv(envVar)
			assert.Empty(t, value, "Expected environment variable %s to be unset", envVar)
		}
	})

	t.Run("SetProxyEnvVars_WithProxyValue", func(t *testing.T) {
		// Setup test environment with an active proxy that has a custom NoProxy value
		setupViper(testConfigPath)
		setupProxies([]ProxyConfig{testProxy1, testProxy2})

		proxyValue := "http://proxy1:8080"
		SetProxyEnvVars(proxyValue)

		// Check standard proxy vars
		for _, envVar := range []string{"ALL_PROXY", "HTTP_PROXY", "HTTPS_PROXY", "GLOBAL_AGENT_HTTP_PROXY", "http_proxy", "https_proxy"} {
			assert.Equal(t, proxyValue, os.Getenv(envVar), "Expected %s to be set to proxy value", envVar)
		}

		// Check no_proxy vars - should include both default and custom values from active proxy
		expectedNoProxy := "localhost,127.0.0.1,::1,.local,internal.example.com"
		assert.Equal(t, expectedNoProxy, os.Getenv("NO_PROXY"), "Expected NO_PROXY to include both default and custom values")
		assert.Equal(t, expectedNoProxy, os.Getenv("no_proxy"), "Expected no_proxy to include both default and custom values")
	})

	t.Run("SetProxyEnvVars_EmptyValue", func(t *testing.T) {
		// First set some values
		setTestEnvVars("http://proxy1:8080")

		// Then unset with empty string
		SetProxyEnvVars("")

		// Verify all are unset
		for _, envVar := range proxyEnvVars {
			assert.Empty(t, os.Getenv(envVar), "Expected environment variable %s to be unset", envVar)
		}
	})
}
