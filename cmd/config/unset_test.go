package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestUnsetCmd_RemovesTopLevelKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".eng.yaml")
	require.NoError(t, os.WriteFile(path, []byte("email: a@b.c\nproject_filter: foo\n"), 0o600))
	viper.Reset()
	defer viper.Reset()
	viper.SetConfigFile(path)

	require.NoError(t, UnsetCmd.RunE(UnsetCmd, []string{"project_filter"}))
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, yaml.Unmarshal(data, &decoded))
	require.NotContains(t, decoded, "project_filter")
	require.Equal(t, "a@b.c", decoded["email"])
}

func TestUnsetCmd_MissingKeyErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".eng.yaml")
	require.NoError(t, os.WriteFile(path, []byte("email: a@b.c\n"), 0o600))
	viper.Reset()
	defer viper.Reset()
	viper.SetConfigFile(path)

	require.Error(t, UnsetCmd.RunE(UnsetCmd, []string{"nope"}))
}

func TestDeleteKey_Dotted(t *testing.T) {
	m := map[string]any{"git": map[string]any{"dev_path": "/x", "editor": "nvim"}}
	require.True(t, deleteKey(m, "git.editor"))
	require.NotContains(t, m["git"], "editor")
	require.False(t, deleteKey(m, "git.missing"))
	require.False(t, deleteKey(m, "nope.key"))
}
