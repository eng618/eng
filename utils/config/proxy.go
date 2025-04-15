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

// ProxyConfig checks for the proxy settings in the configuration and prompts the user to confirm or update them.
// Returns the current proxy setting as a string and whether it is enabled.
func ProxyConfig() (string, bool) {
	log.Start("Checking for proxy configuration")

	proxy := viper.GetString("proxy.value")
	enabled := viper.GetBool("proxy.enabled")

	if proxy == "" {
		updateProxy()
		proxy = viper.GetString("proxy.value")
	}

	var confirm bool
	prompt := &survey.Confirm{
		Message: fmt.Sprintf("Proxy is currently set to: %s (enabled: %v). Is this correct?", color.CyanString(proxy), enabled),
	}
	prompt.Default = true
	err := survey.AskOne(prompt, &confirm)
	cobra.CheckErr(err)

	if !confirm {
		updateProxy()
		proxy = viper.GetString("proxy.value")
	}

	log.Success("Confirmed proxy configuration")
	return proxy, enabled
}

// SetProxyEnabled sets the proxy enabled/disabled in the config and saves it.
func SetProxyEnabled(enabled bool) {
	viper.Set("proxy.enabled", enabled)
	if err := viper.WriteConfig(); err != nil {
		err := errors.New(color.RedString("Error writing config file: %v", err))
		cobra.CheckErr(err)
	}
	log.Success("Proxy enabled set to: %v", enabled)
}

// updateProxy prompts the user to input their proxy address and updates the config.
func updateProxy() {
	var p string
	prompt := &survey.Input{
		Message: "What is your proxy address (e.g., http://proxy:port or leave blank to unset)?",
	}
	err := survey.AskOne(prompt, &p)
	cobra.CheckErr(err)

	viper.Set("proxy.value", p)
	if err := viper.WriteConfig(); err != nil {
		err := errors.New(color.RedString("Error writing config file: %v", err))
		cobra.CheckErr(err)
	}
	log.Success("Proxy address updated successfully")
}
