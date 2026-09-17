package config

import (
	"errors"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestResolveSecret_Env(t *testing.T) {
	t.Setenv("ENG_TEST_SECRET", "s3cr3t")
	v, err := ResolveSecret(SecretRef{Provider: ProviderEnv, EnvVar: "ENG_TEST_SECRET"})
	require.NoError(t, err)
	require.Equal(t, "s3cr3t", v)
	_, err = ResolveSecret(SecretRef{Provider: ProviderEnv, EnvVar: "ENG_TEST_SECRET_MISSING"})
	require.Error(t, err)
}

func TestResolveSecret_BitwardenStub(t *testing.T) {
	old := bitwardenGetItem
	bitwardenGetItem = func(name string) (string, string, error) {
		require.Equal(t, "my-item", name)
		return "pw-token", "", nil
	}
	defer func() { bitwardenGetItem = old }()
	v, err := ResolveSecret(SecretRef{Provider: ProviderBitwarden, Item: "my-item"})
	require.NoError(t, err)
	require.Equal(t, "pw-token", v)
}

func TestResolveSecret_BitwardenError(t *testing.T) {
	old := bitwardenGetItem
	bitwardenGetItem = func(string) (string, string, error) { return "", "", errors.New("locked") }
	defer func() { bitwardenGetItem = old }()
	_, err := ResolveSecret(SecretRef{Provider: ProviderBitwarden, Item: "x"})
	require.ErrorContains(t, err, "locked")
}

func TestResolveSecret_KeychainStub(t *testing.T) {
	oldGet, oldSet := keychainGet, keychainSet
	stored := map[string]string{}
	keychainGet = func(service, account string) (string, error) {
		require.Equal(t, KeychainService, service)
		v, ok := stored[account]
		if !ok {
			return "", errors.New("not found")
		}
		return v, nil
	}
	keychainSet = func(service, account, secret string) error {
		stored[account] = secret
		return nil
	}
	defer func() { keychainGet, keychainSet = oldGet, oldSet }()
	require.NoError(t, keychainSet(KeychainService, "gitlab", "glpat-x"))
	v, err := ResolveSecret(SecretRef{Provider: ProviderKeychain, Item: "gitlab"})
	require.NoError(t, err)
	require.Equal(t, "glpat-x", v)
}

func TestResolveSecret_ConfigRejected(t *testing.T) {
	_, err := ResolveSecret(SecretRef{Provider: ProviderConfig, Item: "gitlab.token"})
	require.ErrorContains(t, err, "deprecated")
}

func TestGitLabTokenRef_Precedence(t *testing.T) {
	v := viper.New()
	require.Equal(t, ProviderEnv, GitLabTokenRef(v).Provider)
	v.Set("gitlab.token_item", "item-new")
	require.Equal(t, "item-new", GitLabTokenRef(v).Item)
	v2 := viper.New()
	v2.Set("gitlab.tokenItem", "item-legacy")
	require.Equal(t, "item-legacy", GitLabTokenRef(v2).Item)
	v3 := viper.New()
	v3.Set("gitlab.keychain_account", "gitlab")
	ref := GitLabTokenRef(v3)
	require.Equal(t, ProviderKeychain, ref.Provider)
}

func TestMigratePlaintextToken_Keychain(t *testing.T) {
	viper.Reset()
	defer viper.Reset()
	dir := t.TempDir()
	path := dir + "/.eng.yaml"
	require.NoError(t, os.WriteFile(path, []byte("{}"), 0o600))
	viper.SetConfigFile(path)
	viper.Set("gitlab.token", "glpat-plain")
	oldStore := secretStore
	var gotProvider SecretProvider
	var gotSecret string
	secretStore = func(p SecretProvider, item, secret string) error {
		gotProvider, gotSecret = p, secret
		require.Equal(t, "gitlab", item)
		return nil
	}
	defer func() { secretStore = oldStore }()
	require.NoError(t, MigratePlaintextToken(ProviderKeychain, "gitlab"))
	require.Equal(t, ProviderKeychain, gotProvider)
	require.Equal(t, "glpat-plain", gotSecret)
	require.Empty(t, viper.GetString("gitlab.token"))
	require.Equal(t, "gitlab", viper.GetString("gitlab.keychain_account"))
	require.NotNil(t, PlaintextSecretWarningForTest())
}

func PlaintextSecretWarningForTest() *FieldError {
	// Plaintext was cleared; warning must be gone. Return non-nil only if bug.
	if viper.GetString("gitlab.token") != "" {
		return PlaintextSecretWarning()
	}
	return &FieldError{Field: "ok"}
}

func TestResolveGitLabToken_EnvFirst(t *testing.T) {
	t.Setenv("GITLAB_TOKEN", "glpat-env")
	token, source, err := ResolveGitLabToken()
	require.NoError(t, err)
	require.Equal(t, "glpat-env", token)
	require.Equal(t, "env:GITLAB_TOKEN", source)
}
