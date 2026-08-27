package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eng618/eng/internal/config"
)

func setupTestConfig(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, ".eng.yaml")
	viper.Reset()
	viper.SetConfigFile(configFile)
	return configFile
}

func TestGenerateUUIDv4(t *testing.T) {
	uuid1 := config.GenerateUUIDv4()
	uuid2 := config.GenerateUUIDv4()

	assert.NotEmpty(t, uuid1)
	assert.NotEmpty(t, uuid2)
	assert.NotEqual(t, uuid1, uuid2)
	assert.Len(t, uuid1, 36)
	assert.Equal(t, byte('4'), uuid1[14], "should be UUID version 4")
}

func TestIsTelemetryEnabled(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		viperVal *bool
		expected bool
	}{
		{
			name:     "default enabled",
			expected: true,
		},
		{
			name:     "disabled via DO_NOT_TRACK=1",
			envVars:  map[string]string{"DO_NOT_TRACK": "1"},
			expected: false,
		},
		{
			name:     "disabled via DO_NOT_TRACK=true",
			envVars:  map[string]string{"DO_NOT_TRACK": "true"},
			expected: false,
		},
		{
			name:     "disabled via ENG_TELEMETRY_DISABLED=1",
			envVars:  map[string]string{"ENG_TELEMETRY_DISABLED": "1"},
			expected: false,
		},
		{
			name:     "enabled via ENG_TELEMETRY_ENABLED=1",
			envVars:  map[string]string{"ENG_TELEMETRY_ENABLED": "1"},
			expected: true,
		},
		{
			name:     "disabled via ENG_TELEMETRY_ENABLED=0",
			envVars:  map[string]string{"ENG_TELEMETRY_ENABLED": "0"},
			expected: false,
		},
		{
			name:     "disabled via viper config",
			viperVal: boolPtr(false),
			expected: false,
		},
		{
			name:     "enabled via viper config",
			viperVal: boolPtr(true),
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			setupTestConfig(t)

			// Clear all test env vars
			os.Unsetenv("DO_NOT_TRACK")
			os.Unsetenv("ENG_TELEMETRY_DISABLED")
			os.Unsetenv("ENG_TELEMETRY_ENABLED")

			// Set test env vars
			for k, v := range tc.envVars {
				t.Setenv(k, v)
			}

			if tc.viperVal != nil {
				viper.Set("telemetry.enabled", *tc.viperVal)
			}

			actual := config.IsTelemetryEnabled()
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestGetEffectiveTelemetryConfig(t *testing.T) {
	setupTestConfig(t)
	t.Setenv("OPENPANEL_CLIENT_ID", "")
	t.Setenv("OPENPANEL_CLIENT_SECRET", "")

	cfg := config.GetEffectiveTelemetryConfig()
	assert.Equal(t, config.DefaultAPIURL, cfg.APIURL)
	assert.Equal(t, config.DefaultClientID, cfg.ClientID)
	assert.Equal(t, config.DefaultClientSecret, cfg.ClientSecret)
	assert.NotEmpty(t, cfg.ProfileID)
}

func TestSetTelemetryConfig(t *testing.T) {
	setupTestConfig(t)

	err := config.SetTelemetryEnabled(true)
	require.NoError(t, err)
	assert.True(t, viper.GetBool("telemetry.enabled"))

	err = config.SetTelemetryEnabled(false)
	require.NoError(t, err)
	assert.False(t, viper.GetBool("telemetry.enabled"))
}

func boolPtr(b bool) *bool {
	return &b
}
