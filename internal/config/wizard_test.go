package config

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func stubWizardPrompts(inputs map[string]string, confirms map[string]bool) func() {
	oldInput, oldConfirm := InputPrompt, ConfirmPrompt
	InputPrompt = func(msg, def string) (string, error) {
		for k, v := range inputs {
			if containsFold(msg, k) {
				return v, nil
			}
		}
		return def, nil
	}
	ConfirmPrompt = func(msg string, def bool) (bool, error) {
		for k, v := range confirms {
			if containsFold(msg, k) {
				return v, nil
			}
		}
		return def, nil
	}
	return func() { InputPrompt, ConfirmPrompt = oldInput, oldConfirm }
}

func containsFold(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i+len(sub) <= len(s); i++ {
				if equalFold(s[i:i+len(sub)], sub) {
					return true
				}
			}
			return false
		}())
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		c, d := a[i], b[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		if d >= 'A' && d <= 'Z' {
			d += 'a' - 'A'
		}
		if c != d {
			return false
		}
	}
	return true
}

func TestParseWizardStep(t *testing.T) {
	s, err := ParseWizardStep("")
	require.NoError(t, err)
	require.Equal(t, WizardProfile, s)
	s, err = ParseWizardStep("dotfiles")
	require.NoError(t, err)
	require.Equal(t, WizardDotfiles, s)
	_, err = ParseWizardStep("nope")
	require.Error(t, err)
}

func TestRunWizard_Full(t *testing.T) {
	restore := stubWizardPrompts(
		map[string]string{
			"email":    "me@x.io",
			"dev path": "/tmp/dev",
			"editor":   "nvim",
			"repo url": "https://x/y.git",
			"branch":   "main",
		},
		map[string]bool{"verbose": true, "bare repo": true, "telemetry": false},
	)
	defer restore()
	changes, err := RunWizard(WizardAnswers{GitDevPath: "/old", DotfilesBranch: "main"}, WizardProfile)
	require.NoError(t, err)
	require.Equal(t, "me@x.io", changes["email"])
	require.Equal(t, "/tmp/dev", changes["git.dev_path"])
	require.Equal(t, "https://x/y.git", changes["dotfiles.repo_url"])
	require.Equal(t, false, changes["telemetry.enabled"])
}

func TestRunWizard_ResumeFromStep(t *testing.T) {
	restore := stubWizardPrompts(nil, map[string]bool{"telemetry": true})
	defer restore()
	changes, err := RunWizard(WizardAnswers{}, WizardTelemetry)
	require.NoError(t, err)
	require.Len(t, changes, 1)
	require.Equal(t, true, changes["telemetry.enabled"])
}

func TestRunWizard_SkipDotfiles(t *testing.T) {
	restore := stubWizardPrompts(nil, map[string]bool{"bare repo": false})
	defer restore()
	changes, err := RunWizard(WizardAnswers{}, WizardDotfiles)
	require.NoError(t, err)
	require.NotContains(t, changes, "dotfiles.repo_url")
}

func TestRunWizard_AbortPropagates(t *testing.T) {
	old := InputPrompt
	InputPrompt = func(_, _ string) (string, error) { return "", errors.New("aborted") }
	defer func() { InputPrompt = old }()
	_, err := RunWizard(WizardAnswers{}, WizardProfile)
	require.Error(t, err)
}

func TestDefaultsFromResolved(t *testing.T) {
	d := DefaultsFromResolved(nil)
	require.Equal(t, "main", d.DotfilesBranch)
	d2 := DefaultsFromResolved(&ResolvedConfig{Email: "a@b.c", DotfilesBranch: ""})
	require.Equal(t, "a@b.c", d2.Email)
	require.Equal(t, "main", d2.DotfilesBranch)
}
