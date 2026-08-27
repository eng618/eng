package config

import (
	"crypto/rand"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

var (
	// DefaultAPIURL is the default OpenPanel API endpoint URL (overridable at build time via ldflags).
	DefaultAPIURL = "https://openpanel.gventureshq.com/api"
	// DefaultClientID is the default OpenPanel client ID (injected at build time via ldflags).
	DefaultClientID = ""
	// DefaultClientSecret is the default OpenPanel client secret (injected at build time via ldflags).
	DefaultClientSecret = ""
)

// GenerateUUIDv4 generates a cryptographically secure RFC 4122 version 4 UUID.
func GenerateUUIDv4() string {
	var uuid [16]byte
	_, err := rand.Read(uuid[:])
	if err != nil {
		// Fallback to pseudo-random based on process ID
		return fmt.Sprintf("anon-%d", os.Getpid())
	}
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant RFC 4122
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
}

// IsTelemetryEnabled checks if telemetry is enabled, honoring environment variables
// (DO_NOT_TRACK, ENG_TELEMETRY_DISABLED, ENG_TELEMETRY_ENABLED) and config settings.
func IsTelemetryEnabled() bool {
	// 1. Check industry-standard DO_NOT_TRACK (https://consoledonottrack.com)
	if dnt := os.Getenv("DO_NOT_TRACK"); dnt == "1" ||
		strings.EqualFold(dnt, "true") ||
		strings.EqualFold(dnt, "yes") {
		return false
	}

	// 2. Check ENG_TELEMETRY_DISABLED
	if disabled := os.Getenv("ENG_TELEMETRY_DISABLED"); disabled == "1" ||
		strings.EqualFold(disabled, "true") ||
		strings.EqualFold(disabled, "yes") {
		return false
	}

	// 3. Check ENG_TELEMETRY_ENABLED env var
	if enabledEnv := os.Getenv("ENG_TELEMETRY_ENABLED"); enabledEnv != "" {
		return enabledEnv == "1" ||
			strings.EqualFold(enabledEnv, "true") ||
			strings.EqualFold(enabledEnv, "yes")
	}

	// 4. Check config file if explicitly set
	if viper.IsSet("telemetry.enabled") {
		return viper.GetBool("telemetry.enabled")
	}

	// 5. Default to true if credentials are built-in or configured
	return true
}

// IsTelemetryConfigured returns true if ClientID and ClientSecret are populated.
func IsTelemetryConfigured() bool {
	clientID := DefaultClientID
	if clientID == "" {
		clientID = os.Getenv("OPENPANEL_CLIENT_ID")
	}
	clientSecret := DefaultClientSecret
	if clientSecret == "" {
		clientSecret = os.Getenv("OPENPANEL_CLIENT_SECRET")
	}
	return clientID != "" && clientSecret != ""
}

// GetEffectiveTelemetryConfig returns the active telemetry configuration.
// Telemetry endpoint and credentials are built-in and fixed to the project servers.
func GetEffectiveTelemetryConfig() TelemetryConfig {
	clientID := DefaultClientID
	if clientID == "" {
		clientID = os.Getenv("OPENPANEL_CLIENT_ID")
	}
	clientSecret := DefaultClientSecret
	if clientSecret == "" {
		clientSecret = os.Getenv("OPENPANEL_CLIENT_SECRET")
	}

	enabled := IsTelemetryEnabled()
	// If credentials are not injected at build time or provided via env, disable quietly
	if clientID == "" || clientSecret == "" {
		enabled = false
	}

	cfg := TelemetryConfig{
		Enabled:      enabled,
		APIURL:       DefaultAPIURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		ProfileID:    viper.GetString("telemetry.profile_id"),
	}

	// Ensure an anonymous Profile ID exists
	if cfg.ProfileID == "" {
		cfg.ProfileID = EnsureProfileID()
	}

	return cfg
}

// EnsureProfileID returns the configured anonymous Profile ID, or generates and saves a new one.
func EnsureProfileID() string {
	profileID := viper.GetString("telemetry.profile_id")
	if profileID != "" {
		return profileID
	}

	profileID = GenerateUUIDv4()
	viper.Set("telemetry.profile_id", profileID)

	// Attempt to save to config file quietly (ignore error if config file isn't writable)
	_ = viper.WriteConfig()
	return profileID
}

// SetTelemetryEnabled updates the enabled state of telemetry in the config file.
func SetTelemetryEnabled(enabled bool) error {
	viper.Set("telemetry.enabled", enabled)
	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("failed to save telemetry configuration: %w", err)
	}
	return nil
}

// Telemetry prompts the user to verify or update their telemetry preferences.
func Telemetry() bool {
	log.Start("Checking telemetry configuration")

	enabled := IsTelemetryEnabled()
	confirm, err := ui.Confirm(
		fmt.Sprintf("Telemetry enabled: %s?", theme.PrimaryText.Render(fmt.Sprintf("%t", enabled))),
		enabled,
	)
	cobra.CheckErr(err)

	if confirm != enabled {
		if err := SetTelemetryEnabled(confirm); err != nil {
			cobra.CheckErr(
				fmt.Errorf(
					"%s: %w",
					lipgloss.NewStyle().Foreground(theme.Destructive).Render("Error writing config file"),
					err,
				),
			)
		}
		enabled = confirm
	}

	log.Success("Confirmed telemetry setting")
	return enabled
}
