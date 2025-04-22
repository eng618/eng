package config

import (
	"errors"
	"fmt"

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
			log.Info("Migrated old proxy configuration format to the new format")
		}
	} else {
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

// SaveProxyConfigs saves the provided proxy configurations to viper config
func SaveProxyConfigs(proxies []ProxyConfig) error {
	viper.Set("proxies", proxies)
	if err := viper.WriteConfig(); err != nil {
		return errors.New(color.RedString("Error writing config file: %v", err))
	}
	return nil
}

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

	// Save the updated configurations
	if err := SaveProxyConfigs(proxies); err != nil {
		return proxies, err
	}

	log.Success("Proxy '%s' enabled", proxies[index].Title)
	return proxies, nil
}

// DisableAllProxies disables all proxy configurations
func DisableAllProxies() error {
	proxies, _ := GetProxyConfigs()

	// Disable all proxies
	for i := range proxies {
		proxies[i].Enabled = false
	}

	return SaveProxyConfigs(proxies)
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
	} else {
		// Add new proxy
		newProxy := ProxyConfig{
			Title:   title,
			Value:   value,
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
