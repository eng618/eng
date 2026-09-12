package editor

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func stubLookPath(t *testing.T, available map[string]bool) {
	t.Helper()
	old := LookPath
	LookPath = func(file string) (string, error) {
		if available[file] {
			return "/usr/bin/" + file, nil
		}
		return "", errors.New("not found")
	}
	t.Cleanup(func() { LookPath = old })
}

func stubStatNotFound(t *testing.T) {
	t.Helper()
	old := Stat
	Stat = func(string) (os.FileInfo, error) {
		return nil, errors.New("not found")
	}
	t.Cleanup(func() { Stat = old })
}

func TestDefaultCommandPrecedence(t *testing.T) {
	tests := []struct {
		name      string
		config    string
		visual    string
		editorEnv string
		available map[string]bool
		expected  string
	}{
		{
			name:      "ExplicitConfigWins",
			config:    "nvim",
			visual:    "code",
			editorEnv: "nano",
			available: map[string]bool{"agy-ide": true, "code": true},
			expected:  "nvim",
		},
		{
			name:      "VisualBeatsAutodetect",
			visual:    "emacs",
			available: map[string]bool{"agy-ide": true, "code": true},
			expected:  "emacs",
		},
		{
			name:      "EditorEnvBeatsAutodetect",
			editorEnv: "vim",
			available: map[string]bool{"agy-ide": true},
			expected:  "vim",
		},
		{name: "AgyIdeFirst", available: map[string]bool{"agy-ide": true, "code": true}, expected: "agy-ide"},
		{name: "CodeSecond", available: map[string]bool{"code": true}, expected: "code"},
		{name: "NanoFallback", available: map[string]bool{}, expected: "nano"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("VISUAL", tc.visual)
			t.Setenv("EDITOR", tc.editorEnv)
			stubLookPath(t, tc.available)
			assert.Equal(t, tc.expected, DefaultCommand(tc.config))
		})
	}
}

func TestResolveAppendsTargetAndSplitsFlags(t *testing.T) {
	stubLookPath(t, map[string]bool{})
	cmd := Resolve("code --wait", "/tmp/proj")
	require.NotNil(t, cmd)
	assert.Equal(t, "code", filepath.Base(cmd.Path))
	assert.Equal(t, []string{"code", "--wait", "/tmp/proj"}, cmd.Args)
}

func TestAvailablePreservesPrecedenceOrder(t *testing.T) {
	stubStatNotFound(t)
	stubLookPath(t, map[string]bool{"agy-ide": true, "code": true, "nano": true, "nvim": true})
	got := Commands()
	assert.Equal(t, []string{"agy-ide", "code", "nvim", "nano"}, got)
}

func TestFindByName(t *testing.T) {
	available := []Option{{Name: "agy-ide (CLI)", Command: "agy-ide"}, {Name: "Nano", Command: "nano"}}
	opt, ok := FindByName(available, "agy-ide (CLI)")
	assert.True(t, ok)
	assert.Equal(t, "agy-ide", opt.Command)
	_, ok = FindByName(available, "missing")
	assert.False(t, ok)
}
