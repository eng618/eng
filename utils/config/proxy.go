package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/utils/log"
)

// ProxyConfig represents a single proxy configuration
type ProxyConfig struct {
	Title   string
	Value   string
	Enabled bool
	NoProxy string
}

// GetProxyConfigs checks for proxy settings in the configuration and returns the current proxies
// and the index of the active proxy (-1 if none are active)
func GetProxyConfigs() ([]ProxyConfig, int) {
	log.Start("Checking for proxy configurations")

	var proxies []ProxyConfig
	activeIndex := -1

	// Read from config
	if !viper.IsSet("proxies") {
		// Handle migration from old format if there's a legacy proxy config
		if viper.IsSet("proxy.value") {
			log.Info("Migrating from old single proxy format to multi-proxy format...")

			title := "Default"
			value := viper.GetString("proxy.value")
			enabled := viper.GetBool("proxy.enabled")

			proxies = append(proxies, ProxyConfig{
				Title:   title,
				Value:   value,
				Enabled: enabled,
			})

			if enabled {
				activeIndex = 0
			}

			// Save in new format
			viper.Set("proxies", proxies)
			// Clean up old format
			viper.Set("proxy", nil)
			if err := viper.WriteConfig(); err != nil {
				err := errors.New(color.RedString("Error writing config file: %w", err))
				cobra.CheckErr(err)
			}
			log.Success("Migration complete: old proxy configuration has been converted to the new format")
		} else {
			// No old format and no new format - initialize with empty array
			viper.Set("proxies", []ProxyConfig{})
			if err := viper.WriteConfig(); err != nil {
				err := errors.New(color.RedString("Error writing config file: %v", err))
				cobra.CheckErr(err)
			}
			log.Info("Initialized empty proxy configurations array")
		}
	} else {
		// Load existing multi-proxy configuration
		err := viper.UnmarshalKey("proxies", &proxies)
		if err != nil {
			log.Error("Failed to unmarshal proxy configurations: %v", err)
			return []ProxyConfig{}, -1
		}

		// Find the active proxy index
		for i, proxy := range proxies {
			if proxy.Enabled {
				activeIndex = i
				break
			}
		}
	}

	log.Success("Proxy configurations loaded")
	return proxies, activeIndex
}

// GetActiveProxy returns the currently active proxy value and true if any proxy is enabled
// If no proxy is enabled, returns the first proxy value and false
// If no proxies exist, returns an empty string and false
func GetActiveProxy() (string, bool) {
	proxies, activeIndex := GetProxyConfigs()

	if activeIndex >= 0 && activeIndex < len(proxies) {
		return proxies[activeIndex].Value, true
	} else if len(proxies) > 0 {
		return proxies[0].Value, false
	}

	return "", false
}

// SaveProxyConfigsFunc defines the function type for saving proxy configs
type SaveProxyConfigsFunc func(proxies []ProxyConfig) error

// SaveProxyConfigsImpl is the actual implementation of saving proxy configurations to viper config
func SaveProxyConfigsImpl(proxies []ProxyConfig) error {
	viper.Set("proxies", proxies)
	if err := viper.WriteConfig(); err != nil {
		return errors.New(color.RedString("Error writing config file: %v", err))
	}
	return nil
}

// SaveProxyConfigs is a variable that holds the function to save proxy configurations
// This can be overridden in tests
var SaveProxyConfigs = SaveProxyConfigsImpl

// EnableProxy enables the proxy at the given index and disables all others
func EnableProxy(index int, proxies []ProxyConfig) ([]ProxyConfig, error) {
	if index < 0 || index >= len(proxies) {
		return proxies, errors.New("proxy index out of range")
	}

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

// DisableAllProxies disables all proxy configurations and unsets environment variables
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

// UnsetProxyEnvVars unsets all proxy-related environment variables
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
			log.Info("Unset environment variable: %s", v)
		}
	}

	log.Success("All proxy environment variables have been unset")
}

// SetProxyEnvVars sets all proxy-related environment variables to the provided value
// and handles custom no_proxy settings
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
			log.Info("Set environment variable: %s=%s", v, proxyValue)
		}
	}

	// Also set lowercase versions
	if err := os.Setenv("http_proxy", proxyValue); err != nil {
		log.Warn("Failed to set environment variable http_proxy=%s: %v", proxyValue, err)
	} else {
		log.Info("Set environment variable: http_proxy=%s", proxyValue)
	}

	if err := os.Setenv("https_proxy", proxyValue); err != nil {
		log.Warn("Failed to set environment variable https_proxy=%s: %v", proxyValue, err)
	} else {
		log.Info("Set environment variable: https_proxy=%s", proxyValue)
	}

	// Set the NO_PROXY variable with default values and any custom values
	noProxyValue := "localhost,127.0.0.1,::1,.local"
	
	// Add custom no_proxy settings if available for the active proxy
	if activeIndex >= 0 && activeIndex < len(proxies) && proxies[activeIndex].NoProxy != "" {
		noProxyValue = noProxyValue + "," + proxies[activeIndex].NoProxy
		log.Info("Adding custom no_proxy values: %s", proxies[activeIndex].NoProxy)
	}
	
	if err := os.Setenv("NO_PROXY", noProxyValue); err != nil {
		log.Warn("Failed to set environment variable NO_PROXY=%s: %v", noProxyValue, err)
	} else {
		log.Info("Set environment variable: NO_PROXY=%s", noProxyValue)
	}

	if err := os.Setenv("no_proxy", noProxyValue); err != nil {
		log.Warn("Failed to set environment variable no_proxy=%s: %v", noProxyValue, err)
	} else {
		log.Info("Set environment variable: no_proxy=%s", noProxyValue)
	}

	log.Success("All proxy environment variables have been set")
}

// AddOrUpdateProxy adds a new proxy or updates an existing one
func AddOrUpdateProxy() ([]ProxyConfig, int) {
	proxies, _ := GetProxyConfigs()

	var title string
	prompt := &survey.Input{
		Message: "Enter a title for this proxy configuration:",
	}
	err := survey.AskOne(prompt, &title)
	cobra.CheckErr(err)

	var value string
	prompt2 := &survey.Input{
		Message: "Enter the proxy address (e.g., http://proxy:port):",
	}
	err = survey.AskOne(prompt2, &value)
	cobra.CheckErr(err)

	var noProxy string
	prompt3 := &survey.Input{
		Message: "Enter additional no_proxy values (comma-separated, leave empty for defaults only):",
		Help:    "These values will be appended to the default no_proxy list: localhost,127.0.0.1,::1,.local",
	}
	err = survey.AskOne(prompt3, &noProxy)
	cobra.CheckErr(err)

	// Check if we're updating an existing proxy
	index := -1
	for i, proxy := range proxies {
		if proxy.Title == title {
			index = i
			break
		}
	}

	if index >= 0 {
		// Update existing proxy
		proxies[index].Value = value
		proxies[index].NoProxy = noProxy
	} else {
		// Add new proxy
		newProxy := ProxyConfig{
			Title:   title,
			Value:   value,
			NoProxy: noProxy,
			Enabled: false,
		}
		proxies = append(proxies, newProxy)
		index = len(proxies) - 1
	}

	// Save configurations
	if err := SaveProxyConfigs(proxies); err != nil {
		cobra.CheckErr(err)
	}

	log.Success("Proxy '%s' added/updated successfully", title)
	return proxies, index
}

// SelectProxy prompts the user to select a proxy from the list and returns the index
func SelectProxy(proxies []ProxyConfig) (int, error) {
	if len(proxies) == 0 {
		return -1, errors.New("no proxy configurations found")
	}

	var options []string
	for _, proxy := range proxies {
		status := " "
		if proxy.Enabled {
			status = "*"
		}
		options = append(options, fmt.Sprintf("[%s] %s (%s)", status, proxy.Title, proxy.Value))
	}

	var selectedIndex int
	prompt := &survey.Select{
		Message: "Select a proxy configuration:",
		Options: options,
	}
	err := survey.AskOne(prompt, &selectedIndex)
	if err != nil {
		return -1, err
	}

	return selectedIndex, nil
}
