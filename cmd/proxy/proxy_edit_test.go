package proxy

import (
	"errors"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/eng618/eng/internal/config"
)

var errTestAbort = errors.New("aborted")

func setupProxyViper(t *testing.T, proxies []config.ProxyConfig) {
	t.Helper()
	viper.Reset()
	viper.SetConfigType("json")
	path := t.TempDir() + "/config.json"
	require.NoError(t, os.WriteFile(path, []byte("{}"), 0o600))
	viper.SetConfigFile(path)
	viper.Set("proxies", proxies)
	// Cobra commands are package singletons: reset flags changed by tests.
	t.Cleanup(func() {
		editCmd.Flags().Set("title", "")
		editCmd.Flags().Set("new-title", "")
		editCmd.Flags().Set("value", "")
		editCmd.Flags().Set("url", "")
		editCmd.Flags().Set("no-proxy", "")
	})
}

func TestEditCmd_ConflictingTargets(t *testing.T) {
	setupProxyViper(t, []config.ProxyConfig{
		{Title: "corp", Value: "http://corp:8080"},
		{Title: "home", Value: "http://home:8080"},
	})
	cmd := *editCmd
	require.NoError(t, cmd.Flags().Set("title", "home"))
	err := cmd.RunE(&cmd, []string{"corp"})
	require.ErrorContains(t, err, "conflicting targets")
}

func TestEditCmd_Rename(t *testing.T) {
	setupProxyViper(t, []config.ProxyConfig{
		{Title: "corp", Value: "http://corp:8080", NoProxy: "internal.corp"},
	})
	cmd := *editCmd
	require.NoError(t, cmd.Flags().Set("title", "corp"))
	require.NoError(t, cmd.Flags().Set("new-title", "hq"))
	require.NoError(t, cmd.RunE(&cmd, []string{}))
	updated, _ := config.GetProxyConfigs()
	require.Len(t, updated, 1)
	require.Equal(t, "hq", updated[0].Title)
	// Rename preserves address and no-proxy.
	require.Equal(t, "http://corp:8080", updated[0].Value)
	require.Equal(t, "internal.corp", updated[0].NoProxy)
}

func TestEditCmd_RenameMissing(t *testing.T) {
	setupProxyViper(t, []config.ProxyConfig{
		{Title: "corp", Value: "http://corp:8080"},
	})
	cmd := *editCmd
	require.NoError(t, cmd.Flags().Set("title", "nope"))
	require.NoError(t, cmd.Flags().Set("new-title", "hq"))
	require.ErrorContains(t, cmd.RunE(&cmd, []string{}), "cannot rename")
}

func TestUseCmd_MissingTargetErrors(t *testing.T) {
	setupProxyViper(t, []config.ProxyConfig{
		{Title: "corp", Value: "http://corp:8080"},
	})
	cmd := *useCmd
	require.ErrorContains(t, cmd.RunE(&cmd, []string{"nope"}), "no proxy configuration found")
}

func TestRemoveCmd_MissingTargetErrors(t *testing.T) {
	setupProxyViper(t, []config.ProxyConfig{
		{Title: "corp", Value: "http://corp:8080"},
	})
	// Point SelectProxy at nothing by giving an unmatched title via flag is
	// not supported for remove selection; use unknown positional which falls
	// to interactive select — stub it to fail instead.
	old := config.SelectPrompt
	config.SelectPrompt = func(_ string, _ []string, _ string) (string, error) {
		return "", errTestAbort
	}
	defer func() { config.SelectPrompt = old }()
	cmd := *removeCmd
	require.Error(t, cmd.RunE(&cmd, []string{"nope"}))
}
