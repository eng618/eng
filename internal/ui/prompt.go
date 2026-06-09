package ui

import (
	"github.com/charmbracelet/huh"
)

// Wrapper functions for testing.
var (
	Confirm     = ConfirmImpl
	Input       = InputImpl
	Select      = SelectImpl
	MultiSelect = MultiSelectImpl
	Password    = PasswordImpl
)

// ConfirmImpl prompts the user with a yes/no question using Huh.
func ConfirmImpl(message string, defaultVal bool) (bool, error) {
	var result bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(message).
				Value(&result).
				Affirmative("Yes").
				Negative("No"),
		),
	).WithTheme(EngTheme())

	err := form.Run()
	if err != nil {
		return defaultVal, err // return default if aborted
	}
	return result, nil
}

// InputImpl prompts the user for text input.
func InputImpl(message, defaultVal string) (string, error) {
	var result string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(message).
				Value(&result),
		),
	).WithTheme(EngTheme())

	err := form.Run()
	if err != nil {
		return defaultVal, err
	}
	if result == "" {
		return defaultVal, nil
	}
	return result, nil
}

// SelectImpl prompts the user to select one option from a list.
func SelectImpl(message string, options []string, defaultVal string) (string, error) {
	var result string
	var huhOptions []huh.Option[string]
	for _, opt := range options {
		huhOptions = append(huhOptions, huh.NewOption(opt, opt))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(message).
				Options(huhOptions...).
				Value(&result),
		),
	).WithTheme(EngTheme())

	err := form.Run()
	return result, err
}

// MultiSelectImpl prompts the user to select multiple options from a list.
func MultiSelectImpl(message string, options []string) ([]string, error) {
	var result []string
	var huhOptions []huh.Option[string]
	for _, opt := range options {
		huhOptions = append(huhOptions, huh.NewOption(opt, opt))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title(message).
				Options(huhOptions...).
				Value(&result),
		),
	).WithTheme(EngTheme())

	err := form.Run()
	return result, err
}

// PasswordImpl prompts the user for hidden text input.
func PasswordImpl(message string) (string, error) {
	var result string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(message).
				EchoMode(huh.EchoModePassword).
				Value(&result),
		),
	).WithTheme(EngTheme())

	err := form.Run()
	return result, err
}
