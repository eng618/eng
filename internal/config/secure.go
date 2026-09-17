package config

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/bitwarden"
)

// Secret providers supported for token storage and lookup.
type SecretProvider string

const (
	ProviderEnv       SecretProvider = "env"
	ProviderBitwarden SecretProvider = "bitwarden"
	ProviderKeychain  SecretProvider = "keychain"
	ProviderConfig    SecretProvider = "config"
)

// SecretRef points at one secret value.
type SecretRef struct {
	Provider SecretProvider
	// Item names the vault item (bitwarden) or account (keychain).
	Item string
	// Field selects a custom field; empty means login password (bitwarden).
	Field string
	// EnvVar names the environment variable (env provider).
	EnvVar string
}

// Exec hooks are vars so tests can stub external CLIs.
var (
	bitwardenGetItem = defaultBitwardenGetItem
	keychainGet      = defaultKeychainGet
	keychainSet      = defaultKeychainSet
	secretStore      = defaultSecretStore
)

// defaultSecretStore persists secret via keychain or bitwarden item.
func defaultSecretStore(provider SecretProvider, item, secret string) error {
	switch provider {
	case ProviderKeychain:
		return keychainSet(KeychainService, item, secret)
	case ProviderBitwarden:
		_, err := bitwarden.SaveOrUpdateBitwardenSecret(item, secret, "GitLab token stored by eng")
		return err
	default:
		return fmt.Errorf("cannot store secret with provider %q (use keychain or bitwarden)", provider)
	}
}

func defaultBitwardenGetItem(name string) (password, tokenField string, err error) {
	if _, err := bitwarden.EnsureBitwardenSession(); err != nil {
		return "", "", fmt.Errorf("bitwarden session: %w", err)
	}
	item, err := bitwarden.GetBitwardenItem(name)
	if err != nil {
		return "", "", err
	}
	if item.Login != nil && item.Login.Password != "" {
		return item.Login.Password, "", nil
	}
	for _, f := range item.Fields {
		if strings.EqualFold(f.Name, "token") && f.Value != "" {
			return "", f.Value, nil
		}
	}
	return "", "", fmt.Errorf("bitwarden item %q has no login password or token field", name)
}

// defaultKeychainGet reads service/account from the OS credential store.
// macOS uses `security`; Linux uses `secret-tool` (libsecret).
func defaultKeychainGet(service, account string) (string, error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		cmd = exec.Command("security", "find-generic-password", "-s", service, "-a", account, "-w")
	} else {
		cmd = exec.Command("secret-tool", "lookup", "service", service, "account", account)
	}
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("keychain lookup %s/%s: %w", service, account, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// defaultKeychainSet stores service/account in the OS credential store.
func defaultKeychainSet(service, account, secret string) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		cmd = exec.Command("security", "add-generic-password", "-U", "-s", service, "-a", account, "-w", secret)
	} else {
		cmd = exec.Command("secret-tool", "store", "--label", "eng "+service,
			"service", service, "account", account)
		cmd.Stdin = strings.NewReader(secret)
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("keychain store %s/%s: %w: %s", service, account, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// ResolveSecret returns the secret value for ref.
func ResolveSecret(ref SecretRef) (string, error) {
	switch ref.Provider {
	case ProviderEnv:
		if ref.EnvVar == "" {
			return "", fmt.Errorf("env secret ref missing env var name")
		}
		if v, ok := os.LookupEnv(ref.EnvVar); ok {
			return v, nil
		}
		return "", fmt.Errorf("environment variable %s is not set", ref.EnvVar)
	case ProviderBitwarden:
		if ref.Item == "" {
			return "", fmt.Errorf("bitwarden secret ref missing item name")
		}
		password, tokenField, err := bitwardenGetItem(ref.Item)
		if err != nil {
			return "", err
		}
		if ref.Field != "" {
			if tokenField != "" {
				return tokenField, nil
			}
			return "", fmt.Errorf("bitwarden item %q has no field %q", ref.Item, ref.Field)
		}
		if password != "" {
			return password, nil
		}
		return tokenField, nil
	case ProviderKeychain:
		if ref.Item == "" {
			return "", fmt.Errorf("keychain secret ref missing account name")
		}
		return keychainGet(KeychainService, ref.Item)
	case ProviderConfig:
		return "", fmt.Errorf("plaintext config storage is deprecated: migrate with `eng config secrets migrate`")
	default:
		return "", fmt.Errorf("unknown secret provider %q", ref.Provider)
	}
}

// KeychainService is the OS keychain service name for eng secrets.
const KeychainService = "eng"

// GitLabTokenRef builds the token lookup for GitLab: explicit env first,
// then the configured bitwarden item (new token_item key, legacy tokenItem),
// then the keychain account, never plaintext.
func GitLabTokenRef(v *viper.Viper) SecretRef {
	if item := v.GetString("gitlab.token_item"); item != "" {
		return SecretRef{Provider: ProviderBitwarden, Item: item}
	}
	if item := v.GetString("gitlab.tokenItem"); item != "" {
		return SecretRef{Provider: ProviderBitwarden, Item: item}
	}
	if account := v.GetString("gitlab.keychain_account"); account != "" {
		return SecretRef{Provider: ProviderKeychain, Item: account}
	}
	return SecretRef{Provider: ProviderEnv, EnvVar: "GITLAB_TOKEN"}
}

// PlaintextSecretWarning reports whether gitlab.token is stored in plaintext.
func PlaintextSecretWarning() *FieldError {
	if viper.GetString("gitlab.token") == "" {
		return nil
	}
	return &FieldError{
		Field: "gitlab.token", Value: "(redacted)",
		Hint: "plaintext in config: run `eng config secrets migrate --to keychain|bitwarden`",
	}
}

// MigratePlaintextToken moves gitlab.token from plaintext config into the
// given store, clears the plaintext key, and records the reference.
// Bitwarden stores under item and sets gitlab.token_item; keychain stores
// under account and sets gitlab.keychain_account.
func MigratePlaintextToken(provider SecretProvider, item string) error {
	v := viper.GetViper()
	plain := v.GetString("gitlab.token")
	if plain == "" {
		return fmt.Errorf("no plaintext gitlab.token to migrate")
	}
	if item == "" {
		item = "gitlab"
	}
	if err := secretStore(provider, item, plain); err != nil {
		return err
	}
	v.Set("gitlab.token", "")
	switch provider {
	case ProviderBitwarden:
		v.Set("gitlab.token_item", item)
	case ProviderKeychain:
		v.Set("gitlab.keychain_account", item)
	}
	if file := v.ConfigFileUsed(); file != "" {
		if err := v.WriteConfig(); err != nil {
			return fmt.Errorf("write config: %w", err)
		}
	}
	return nil
}

// ResolveGitLabToken resolves the GitLab token with precedence:
// env GITLAB_TOKEN, bitwarden item, keychain account, legacy plaintext
// gitlab.token (returned with a migration warning source).
func ResolveGitLabToken() (token, source string, err error) {
	if v, ok := os.LookupEnv("GITLAB_TOKEN"); ok && v != "" {
		return v, "env:GITLAB_TOKEN", nil
	}
	v := viper.GetViper()
	ref := GitLabTokenRef(v)
	if ref.Provider != ProviderEnv {
		token, err := ResolveSecret(ref)
		if err == nil {
			return token, string(ref.Provider) + ":" + ref.Item, nil
		}
	}
	if plain := v.GetString("gitlab.token"); plain != "" {
		return plain, "config:gitlab.token (plaintext, run `eng config secrets migrate`)", nil
	}
	return "", "", fmt.Errorf(
		"no GitLab token: set GITLAB_TOKEN, configure a bitwarden item, or run `eng config secrets migrate`",
	)
}
