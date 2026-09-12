package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eng618/eng/internal/editor"
)

func setupEditorTestViper(t *testing.T) {
	t.Helper()
	viper.Reset()
	viper.SetConfigType("yaml")
	path := filepath.Join(t.TempDir(), ".eng.yaml")
	require.NoError(t, os.WriteFile(path, []byte("{}"), 0o644))
	viper.SetConfigFile(path)
}

func stubEditorPrompts(t *testing.T) {
	t.Helper()
	oldConfirm, oldInput, oldSelect := ConfirmPrompt, InputPrompt, SelectPrompt
	ConfirmPrompt = func(_ string, def bool) (bool, error) { return def, nil }
	InputPrompt = func(_, def string) (string, error) { return def, nil }
	SelectPrompt = func(_ string, _ []string, def string) (string, error) { return def, nil }
	t.Cleanup(func() {
		ConfirmPrompt, InputPrompt, SelectPrompt = oldConfirm, oldInput, oldSelect
	})
}

func stubEditorDetection(t *testing.T, available map[string]bool) {
	t.Helper()
	oldLook, oldStat := editor.LookPath, editor.Stat
	editor.LookPath = func(file string) (string, error) {
		if available[file] {
			return "/usr/bin/" + file, nil
		}
		return "", os.ErrNotExist
	}
	editor.Stat = func(string) (os.FileInfo, error) { return nil, os.ErrNotExist }
	t.Cleanup(func() {
		editor.LookPath, editor.Stat = oldLook, oldStat
	})
}

func TestSetGitEditorPersists(t *testing.T) {
	setupEditorTestViper(t)
	SetGitEditor("agy-ide")
	assert.Equal(t, "agy-ide", viper.GetString("git.editor"))
}

func TestGitEditorDirectArg(t *testing.T) {
	setupEditorTestViper(t)
	stubEditorPrompts(t)
	got := GitEditor("code")
	assert.Equal(t, "code", got)
	assert.Equal(t, "code", viper.GetString("git.editor"))
}

func TestGitEditorConfirmKeepsCurrent(t *testing.T) {
	setupEditorTestViper(t)
	stubEditorPrompts(t)
	viper.Set("git.editor", "nvim")
	got := GitEditor()
	assert.Equal(t, "nvim", got)
	assert.Equal(t, "nvim", viper.GetString("git.editor"))
}

func TestUpdateGitEditorSelectsDetected(t *testing.T) {
	setupEditorTestViper(t)
	stubEditorDetection(t, map[string]bool{"agy-ide": true, "code": true})
	SelectPrompt = func(_ string, options []string, _ string) (string, error) {
		assert.Contains(t, options, "agy-ide (CLI)")
		return "agy-ide (CLI)", nil
	}
	t.Cleanup(func() { SelectPrompt = selectUnwired })
	got := UpdateGitEditor()
	assert.Equal(t, "agy-ide", got)
	assert.Equal(t, "agy-ide", viper.GetString("git.editor"))
}

func TestUpdateGitEditorCustomInput(t *testing.T) {
	setupEditorTestViper(t)
	stubEditorDetection(t, map[string]bool{"code": true})
	SelectPrompt = func(_ string, _ []string, _ string) (string, error) { return customEditorLabel, nil }
	InputPrompt = func(_, _ string) (string, error) { return "hx", nil }
	t.Cleanup(func() {
		SelectPrompt = selectUnwired
		InputPrompt = inputUnwired
	})
	got := UpdateGitEditor()
	assert.Equal(t, "hx", got)
	assert.Equal(t, "hx", viper.GetString("git.editor"))
}
