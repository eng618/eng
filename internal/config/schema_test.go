package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestMigrateMap_V0LegacyKeys(t *testing.T) {
	in := map[string]any{
		"dotfiles":   map[string]any{"repoPath": "/old/repo", "worktree": "/old/wt"},
		"git":        map[string]any{"devPath": "/old/dev"},
		"user-email": "a@b.c",
		"gitlab":     map[string]any{"tokenItem": "op://x"},
	}
	out, changed, err := MigrateMap(in)
	require.NoError(t, err)
	require.True(t, changed)
	require.Equal(t, CurrentVersion, out["version"])
	require.Equal(t, "/old/repo", out["dotfiles"].(map[string]any)["bare_repo_path"])
	require.Equal(t, "/old/wt", out["dotfiles"].(map[string]any)["worktree_path"])
	require.Equal(t, "/old/dev", out["git"].(map[string]any)["dev_path"])
	require.Equal(t, "a@b.c", out["email"])
	require.Equal(t, "op://x", out["gitlab"].(map[string]any)["token_item"])
	// Old keys removed.
	_, ok := out["dotfiles"].(map[string]any)["repoPath"]
	require.False(t, ok)
	_, ok = out["user-email"]
	require.False(t, ok)
	_, ok = out["gitlab"].(map[string]any)["tokenItem"]
	require.False(t, ok)
	require.Contains(t, out, "projects")
}

func TestMigrateMap_NewKeyWins(t *testing.T) {
	in := map[string]any{
		"git": map[string]any{"devPath": "/old", "dev_path": "/new"},
	}
	out, changed, err := MigrateMap(in)
	require.NoError(t, err)
	require.True(t, changed)
	require.Equal(t, "/new", out["git"].(map[string]any)["dev_path"])
	_, ok := out["git"].(map[string]any)["devPath"]
	require.False(t, ok)
}

func TestMigrateMap_LegacyProxy(t *testing.T) {
	in := map[string]any{
		"proxy": map[string]any{"value": "http://proxy:8080", "enabled": true},
	}
	out, changed, err := MigrateMap(in)
	require.NoError(t, err)
	require.True(t, changed)
	require.NotContains(t, out, "proxy")
	proxies, ok := out["proxies"].([]any)
	require.True(t, ok)
	require.Len(t, proxies, 1)
	first := proxies[0].(map[string]any)
	require.Equal(t, "Default", first["title"])
	require.Equal(t, "http://proxy:8080", first["value"])
	require.Equal(t, true, first["enabled"])
}

func TestMigrateMap_MultiEnabledRepair(t *testing.T) {
	in := map[string]any{
		"proxies": []any{
			map[string]any{"title": "a", "value": "http://a:8080", "enabled": true},
			map[string]any{"title": "b", "value": "http://b:8080", "enabled": true},
		},
		"version":  CurrentVersion,
		"projects": []any{},
	}
	out, changed, err := MigrateMap(in)
	require.NoError(t, err)
	require.True(t, changed)
	proxies := out["proxies"].([]any)
	require.Equal(t, true, proxies[0].(map[string]any)["enabled"])
	require.Equal(t, false, proxies[1].(map[string]any)["enabled"])
}

func TestMigrateFile_BackupAndIdempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".eng.yaml")
	legacy := "dotfiles:\n  repoPath: /old/repo\nemail: x@y.z\n"
	require.NoError(t, os.WriteFile(path, []byte(legacy), 0o600))

	backup, changed, err := MigrateFile(path)
	require.NoError(t, err)
	require.True(t, changed)
	require.NotEmpty(t, backup)
	require.FileExists(t, backup)

	// Migrated content parses and carries the version stamp.
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, yaml.Unmarshal(data, &decoded))
	require.Equal(t, CurrentVersion, decoded["version"])

	// Second run is a no-op.
	backup2, changed2, err := MigrateFile(path)
	require.NoError(t, err)
	require.False(t, changed2)
	require.Empty(t, backup2)
}
