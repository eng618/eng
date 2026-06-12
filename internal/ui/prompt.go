package ui

import (
	"errors"

	"github.com/charmbracelet/huh"

	"github.com/eng618/eng/internal/ui/theme"
)

// Wrapper functions for testing.
var (
	Confirm     = ConfirmImpl
	Input       = InputImpl
	Select      = SelectImpl
	MultiSelect = MultiSelectImpl
	Password    = PasswordImpl
)

// ConfirmImpl prompts the user with a yes/no question using the default theme.
func ConfirmImpl(message string, defaultVal bool) (bool, error) {
	var val bool
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(message).
				Value(&val).
				Affirmative("Yes!").
				Negative("No."),
		),
	).WithTheme(theme.EngTheme()).Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return false, err
		}
		return defaultVal, err
	}
	return val, nil
}

// InputImpl prompts the user for text input using the default theme.
func InputImpl(message, defaultVal string) (string, error) {
	var val string
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(message).
				Value(&val).
				Placeholder(defaultVal),
		),
	).WithTheme(theme.EngTheme()).Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", err
		}
		return defaultVal, err
	}
	if val == "" {
		return defaultVal, nil
	}
	return val, nil
}

// SelectImpl prompts the user to select an option from a list using the default theme.
func SelectImpl(message string, options []string, defaultVal string) (string, error) {
	var val string

	huhOptions := make([]huh.Option[string], len(options))
	for i, opt := range options {
		huhOptions[i] = huh.NewOption(opt, opt)
	}

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(message).
				Options(huhOptions...).
				Value(&val),
		),
	).WithTheme(theme.EngTheme()).Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", err
		}
		return defaultVal, err
	}
	return val, nil
}

// MultiSelectImpl prompts the user to select multiple options from a list using the default theme.
func MultiSelectImpl(message string, options, defaultSelected []string) ([]string, error) {
	var val []string

	huhOptions := make([]huh.Option[string], len(options))
	for i, opt := range options {
		selected := false
		for _, def := range defaultSelected {
			if opt == def {
				selected = true
				break
			}
		}
		huhOptions[i] = huh.NewOption(opt, opt).Selected(selected)
	}

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title(message).
				Options(huhOptions...).
				Value(&val),
		),
	).WithTheme(theme.EngTheme()).Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil, err
		}
		return nil, err
	}
	return val, nil
}

// PasswordImpl prompts the user for a secret text input using the default theme.
func PasswordImpl(message string) (string, error) {
	var val string
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(message).
				Value(&val).
				EchoMode(huh.EchoModePassword),
		),
	).WithTheme(theme.EngTheme()).Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", err
		}
		return "", err
	}
	return val, nil
}
