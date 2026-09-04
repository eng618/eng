package config

import (
	"errors"
	"strings"

	"github.com/charmbracelet/huh"

	"github.com/eng618/eng/internal/ui/theme"
)

// ProxyConfig represents a single proxy configuration.
type PromptProxyValuesFunc func(initialTitle, initialValue, initialNoProxy string) (title, value, noProxy string, err error)

// PromptProxyValuesImpl uses huh.NewForm to collect proxy details interactively.
func PromptProxyValuesImpl(
	initialTitle, initialValue, initialNoProxy string,
) (title, value, noProxy string, err error) {
	title = initialTitle
	value = initialValue
	noProxy = initialNoProxy

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Title").
				Description("A unique name for this proxy configuration (e.g., Corp, Home).").
				Value(&title).
				Validate(func(s string) error {
					return validateTitle(s)
				}),
			huh.NewInput().
				Title("Proxy Address").
				Description("The address of the proxy (e.g., http://proxy:port).").
				Value(&value).
				Validate(func(s string) error {
					return validateProxyURL(s)
				}),
			huh.NewInput().
				Title("No Proxy").
				Description("Additional no_proxy values (comma-separated). Appended to defaults.").
				Placeholder("localhost,127.0.0.1,::1,.local").
				Value(&noProxy),
		),
	).WithTheme(theme.EngTheme())

	err = form.Run()
	return title, value, noProxy, err
}

// PromptProxyValues is a variable that holds the function to prompt for proxy values.
// This can be overridden in tests.
var PromptProxyValues = PromptProxyValuesImpl

// EnableProxy enables the proxy at the given index and disables all others.
func validateTitle(val interface{}) error {
	s, ok := val.(string)
	if !ok {
		return errors.New("invalid title input")
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return errors.New("title is required")
	}
	return nil
}

// validateProxyURL is a survey validator wrapper around ValidateProxyURLString.
func validateProxyURL(val interface{}) error {
	s, ok := val.(string)
	if !ok {
		return errors.New("invalid proxy input")
	}
	return ValidateProxyURLString(s)
}

// ValidateProxyURLString validates the proxy URL string for scheme and host:port.
