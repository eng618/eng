package config

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui/theme"
)

// ProxyConfig represents a single proxy configuration.
type ProxyConfig struct {
	Title   string
	Value   string
	Enabled bool
	NoProxy string
}

// GetProxyConfigs checks for proxy settings in the configuration and returns the current proxies
// and the index of the active proxy (-1 if none are active).
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
				err := fmt.Errorf(
					"%s: %w",
					lipgloss.NewStyle().Foreground(theme.Destructive).Render("Error writing config file"),
					err,
				)
				cobra.CheckErr(err)
			}
			log.Success("Migration complete: old proxy configuration has been converted to the new format")
		} else {
			// No old format and no new format - initialize with empty array
			viper.Set("proxies", []ProxyConfig{})
			if err := viper.WriteConfig(); err != nil {
				err := fmt.Errorf(
					"%s: %w",
					lipgloss.NewStyle().Foreground(theme.Destructive).Render("Error writing config file"),
					err,
				)
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
// If no proxies exist, returns an empty string and false.
func GetActiveProxy() (string, bool) {
	proxies, activeIndex := GetProxyConfigs()

	if activeIndex >= 0 && activeIndex < len(proxies) {
		return proxies[activeIndex].Value, true
	} else if len(proxies) > 0 {
		return proxies[0].Value, false
	}

	return "", false
}

// SaveProxyConfigsFunc defines the function type for saving proxy configs.
type SaveProxyConfigsFunc func(proxies []ProxyConfig) error

// SaveProxyConfigsImpl is the actual implementation of saving proxy configurations to viper config.
func SaveProxyConfigsImpl(proxies []ProxyConfig) error {
	viper.Set("proxies", proxies)
	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf(
			"%s: %w",
			lipgloss.NewStyle().Foreground(theme.Destructive).Render("Error writing config file"),
			err,
		)
	}
	return nil
}

// SaveProxyConfigs is a variable that holds the function to save proxy configurations
// This can be overridden in tests.
var SaveProxyConfigs = SaveProxyConfigsImpl

// PromptProxyValuesFunc defines the function type for prompting for proxy values.
