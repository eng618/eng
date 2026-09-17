package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/log"
)

// ProxyConfig represents a single proxy configuration.
// EnableProxy enables the proxy at index. It operates on a copy of proxies
// and returns the updated slice; the caller's slice is never mutated.
func EnableProxy(index int, proxies []ProxyConfig) ([]ProxyConfig, error) {
	if index < 0 || index >= len(proxies) {
		return proxies, errors.New("proxy index out of range")
	}
	out := make([]ProxyConfig, len(proxies))
	copy(out, proxies)

	// Validate proxy URL before enabling
	// Normalize scheme-less values (default to http)
	normalized := NormalizeProxyURLString(out[index].Value)
	if err := ValidateProxyURLString(normalized); err != nil {
		return proxies, fmt.Errorf("invalid proxy URL '%s': %w", out[index].Value, err)
	}

	// Persist normalized value
	out[index].Value = normalized

	// Disable all proxies first
	for i := range out {
		out[i].Enabled = false
	}

	// Enable the selected proxy
	out[index].Enabled = true

	// Set environment variables from the just-saved proxy, so its custom
	// NoProxy applies immediately instead of a pre-save snapshot.
	SetProxyEnvVars(out[index])

	// Save the updated configurations
	if err := SaveProxyConfigs(out); err != nil {
		return proxies, err
	}

	log.Success("Proxy '%s' enabled", out[index].Title)
	return out, nil
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

// SetProxyEnvVars sets all proxy-related environment variables from the
// given proxy, including its custom NoProxy hosts.
func SetProxyEnvVars(proxy ProxyConfig) {
	// List of proxy environment variables to set
	vars := []string{
		"ALL_PROXY",
		"HTTP_PROXY",
		"HTTPS_PROXY",
		"GLOBAL_AGENT_HTTP_PROXY",
	}

	if proxy.Value == "" {
		// If proxy value is empty, just unset
		UnsetProxyEnvVars()
		return
	}

	// Set the environment variables
	for _, v := range vars {
		if err := os.Setenv(v, proxy.Value); err != nil {
			log.Warn("Failed to set environment variable %s=%s: %v", v, proxy.Value, err)
		} else {
			log.Verbose(viper.GetBool("verbose"), "Set environment variable: %s=%s", v, proxy.Value)
		}
	}

	// Also set lowercase versions
	if err := os.Setenv("http_proxy", proxy.Value); err != nil {
		log.Warn("Failed to set environment variable http_proxy=%s: %v", proxy.Value, err)
	} else {
		log.Verbose(viper.GetBool("verbose"), "Set environment variable: http_proxy=%s", proxy.Value)
	}

	if err := os.Setenv("https_proxy", proxy.Value); err != nil {
		log.Warn("Failed to set environment variable https_proxy=%s: %v", proxy.Value, err)
	} else {
		log.Verbose(viper.GetBool("verbose"), "Set environment variable: https_proxy=%s", proxy.Value)
	}

	// Set the NO_PROXY variable with default values and any custom values
	noProxyValue := "localhost,127.0.0.1,::1,.local"

	// Add custom no_proxy settings from the given proxy
	if proxy.NoProxy != "" {
		noProxyValue = noProxyValue + "," + proxy.NoProxy
		log.Verbose(viper.GetBool("verbose"), "Adding custom no_proxy values: %s", proxy.NoProxy)
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
