package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), ".eng.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

func TestLoadResolved_PrecedenceFileOverDefault(t *testing.T) {
	path := writeConfig(t, "email: file@example.com\ngit:\n  dev_path: /tmp/dev\n")
	v := NewLoader(path)
	rc, err := LoadResolved(v)
	require.NoError(t, err)
	require.Equal(t, "file@example.com", rc.Email)
	require.Equal(t, "file", rc.Source["email"])
	require.Equal(t, "default", rc.Source["containers.path"])
	require.Equal(t, "main", rc.DotfilesBranch)
}

func TestLoadResolved_EnvBeatsFile(t *testing.T) {
	path := writeConfig(t, "email: file@example.com\n")
	t.Setenv("ENG_EMAIL", "env@example.com")
	v := NewLoader(path)
	rc, err := LoadResolved(v)
	require.NoError(t, err)
	require.Equal(t, "env@example.com", rc.Email)
	require.Equal(t, "env", rc.Source["email"])
}

func TestLoadResolved_EnvDottedKey(t *testing.T) {
	path := writeConfig(t, "git:\n  dev_path: /tmp/file\n")
	t.Setenv("ENG_GIT_DEV_PATH", "/tmp/env")
	v := NewLoader(path)
	rc, err := LoadResolved(v)
	require.NoError(t, err)
	require.Equal(t, "/tmp/env", rc.GitDevPath)
	require.Equal(t, "env", rc.Source["git.dev_path"])
}

func TestLoadResolved_ExpandsTilde(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	path := writeConfig(t, "git:\n  dev_path: ~/dev\ncontainers:\n  path: $HOME/containers\n")
	v := NewLoader(path)
	rc, err := LoadResolved(v)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(home, "dev"), rc.GitDevPath)
	require.Equal(t, filepath.Join(home, "containers"), rc.ContainersPath)
}

func TestLoadResolved_RejectsUnknownKeys(t *testing.T) {
	path := writeConfig(t, "bogus_key: 1\n")
	v := NewLoader(path)
	_, err := LoadResolved(v)
	require.Error(t, err)
}

func TestLoadResolved_AcceptsKnownKeys(t *testing.T) {
	path := writeConfig(
		t,
		"proxies:\n  - title: Default\n    value: http://proxy:8080\n    enabled: false\nproxy:\n  value: http://old:8080\n  enabled: false\n",
	)
	v := NewLoader(path)
	_, err := LoadResolved(v)
	require.NoError(t, err)
}

func TestFindUnknownKeys(t *testing.T) {
	path := writeConfig(t, "email: a@b.c\ndry_run: true\nproject_filter: x\n")
	require.Equal(t, []string{"dry_run", "project_filter"}, FindUnknownKeys(path))
	clean := writeConfig(t, "email: a@b.c\n")
	require.Empty(t, FindUnknownKeys(clean))
	require.Empty(t, FindUnknownKeys("/nonexistent/path.yaml"))
}

func TestResolveConfigPath(t *testing.T) {
	require.Equal(t, "/custom/path.yaml", ResolveConfigPath("/custom/path.yaml"))
	require.Equal(t, DefaultConfigPath(), ResolveConfigPath(""))
	xdgHome := t.TempDir()
	xdg := filepath.Join(xdgHome, "eng", "config.yaml")
	require.NoError(t, os.MkdirAll(filepath.Dir(xdg), 0o755))
	require.NoError(t, os.WriteFile(xdg, []byte("email: a@b.c\n"), 0o600))
	t.Setenv("XDG_CONFIG_HOME", xdgHome)
	require.Equal(t, xdg, ResolveConfigPath(""))
}
