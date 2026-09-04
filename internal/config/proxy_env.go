package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/log"
)

// ProxyConfig represents a single proxy configuration.
func EnableProxy(index int, proxies []ProxyConfig) ([]ProxyConfig, error) {
	if index < 0 || index >= len(proxies) {
		return proxies, errors.New("proxy index out of range")
	}

	// Validate proxy URL before enabling
	// Normalize scheme-less values (default to http)
	normalized := NormalizeProxyURLString(proxies[index].Value)
	if err := ValidateProxyURLString(normalized); err != nil {
		return proxies, fmt.Errorf("invalid proxy URL '%s': %w", proxies[index].Value, err)
	}

	// Persist normalized value
	proxies[index].Value = normalized

	// Disable all proxies first
	for i := range proxies {
		proxies[i].Enabled = false
	}

	// Enable the selected proxy
	proxies[index].Enabled = true

	// Set environment variables for the enabled proxy
	SetProxyEnvVars(proxies[index].Value)

	// Save the updated configurations
	if err := SaveProxyConfigs(proxies); err != nil {
		return proxies, err
	}

	log.Success("Proxy '%s' enabled", proxies[index].Title)
	return proxies, nil
}

// DisableAllProxies disables all proxy configurations and unsets environment variables.
func DisableAllProxies() error {
	proxies, _ := GetProxyConfigs()

	// Disable all proxies
	for i := range proxies {
		proxies[i].Enabled = false
	}

	// Unset environment variables
	UnsetProxyEnvVars()

	return SaveProxyConfigs(proxies)
}

// UnsetProxyEnvVars unsets all proxy-related environment variables.
func UnsetProxyEnvVars() {
	// List of proxy environment variables to unset
	vars := []string{
		"ALL_PROXY",
		"HTTP_PROXY",
		"HTTPS_PROXY",
		"GLOBAL_AGENT_HTTP_PROXY",
		"NO_PROXY",
		"http_proxy",
		"https_proxy",
		"no_proxy",
	}

	for _, v := range vars {
		if err := os.Unsetenv(v); err != nil {
			log.Warn("Failed to unset environment variable %s: %v", v, err)
		} else {
			log.Verbose(viper.GetBool("verbose"), "Unset environment variable: %s", v)
		}
	}

	log.Success("All proxy environment variables have been unset")
}

// SetProxyEnvVars sets all proxy-related environment variables to the provided value
// and handles custom no_proxy settings.
func SetProxyEnvVars(proxyValue string) {
	// Get the active proxy configuration to access custom NoProxy settings
	proxies, activeIndex := GetProxyConfigs()

	// List of proxy environment variables to set
	vars := []string{
		"ALL_PROXY",
		"HTTP_PROXY",
		"HTTPS_PROXY",
		"GLOBAL_AGENT_HTTP_PROXY",
	}

	if proxyValue == "" {
		// If proxy value is empty, just unset
		UnsetProxyEnvVars()
		return
	}

	// Set the environment variables
	for _, v := range vars {
		if err := os.Setenv(v, proxyValue); err != nil {
			log.Warn("Failed to set environment variable %s=%s: %v", v, proxyValue, err)
		} else {
			log.Verbose(viper.GetBool("verbose"), "Set environment variable: %s=%s", v, proxyValue)
		}
	}

	// Also set lowercase versions
	if err := os.Setenv("http_proxy", proxyValue); err != nil {
		log.Warn("Failed to set environment variable http_proxy=%s: %v", proxyValue, err)
	} else {
		log.Verbose(viper.GetBool("verbose"), "Set environment variable: http_proxy=%s", proxyValue)
	}

	if err := os.Setenv("https_proxy", proxyValue); err != nil {
		log.Warn("Failed to set environment variable https_proxy=%s: %v", proxyValue, err)
	} else {
		log.Verbose(viper.GetBool("verbose"), "Set environment variable: https_proxy=%s", proxyValue)
	}

	// Set the NO_PROXY variable with default values and any custom values
	noProxyValue := "localhost,127.0.0.1,::1,.local"

	// Add custom no_proxy settings if available for the active proxy
	if activeIndex >= 0 && activeIndex < len(proxies) && proxies[activeIndex].NoProxy != "" {
		noProxyValue = noProxyValue + "," + proxies[activeIndex].NoProxy
		log.Verbose(viper.GetBool("verbose"), "Adding custom no_proxy values: %s", proxies[activeIndex].NoProxy)
	}

	if err := os.Setenv("NO_PROXY", noProxyValue); err != nil {
		log.Warn("Failed to set environment variable NO_PROXY=%s: %v", noProxyValue, err)
	} else {
		log.Verbose(viper.GetBool("verbose"), "Set environment variable: NO_PROXY=%s", noProxyValue)
	}

	if err := os.Setenv("no_proxy", noProxyValue); err != nil {
		log.Warn("Failed to set environment variable no_proxy=%s: %v", noProxyValue, err)
	} else {
		log.Verbose(viper.GetBool("verbose"), "Set environment variable: no_proxy=%s", noProxyValue)
	}

	log.Success("All proxy environment variables have been set")
}

// AddOrUpdateProxy adds a new proxy or updates an existing one.
