package config

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
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

// GetProxyConfigs reads proxy settings and returns the current proxies
// and the index of the active proxy (-1 if none are active).
// It is a pure read: it never writes the config file. Legacy singular
// `proxy:` configs are converted in memory with a warning pointing at
// `eng config migrate`, which owns persistent migration.
func GetProxyConfigs() ([]ProxyConfig, int) {
	log.Start("Checking for proxy configurations")

	var proxies []ProxyConfig
	activeIndex := -1

	// Read from config
	if !viper.IsSet("proxies") {
		// Legacy singular format: convert in memory only.
		if viper.IsSet("proxy.value") {
			log.Info("Legacy single proxy format detected; run `eng config migrate` to convert it persistently")

			proxies = append(proxies, ProxyConfig{
				Title:   "Default",
				Value:   viper.GetString("proxy.value"),
				Enabled: viper.GetBool("proxy.enabled"),
			})

			if proxies[0].Enabled {
				activeIndex = 0
			}
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
