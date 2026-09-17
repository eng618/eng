package config

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui/theme"
)

// SecretsCmd manages secret storage for tokens.
var SecretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "Manage secret storage (bitwarden, keychain, env)",
	Long: `Inspect and migrate secret storage. Tokens resolve with precedence:
env vars first, then Bitwarden items, then OS keychain accounts.
Plaintext config storage is deprecated.

Example: eng config secrets migrate --to keychain --item gitlab`,
}

// SecretsMigrateCmd moves plaintext tokens into a secure store.
var SecretsMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Move plaintext tokens from config into keychain or bitwarden",
	Long: `Moves gitlab.token from plaintext config into the chosen store,
clears the plaintext key, and records the reference.

Example: eng config secrets migrate --to keychain --item gitlab`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		to, _ := cmd.Flags().GetString("to")
		item, _ := cmd.Flags().GetString("item")
		provider := config.SecretProvider(to)
		if provider != config.ProviderKeychain && provider != config.ProviderBitwarden {
			return fmt.Errorf("invalid --to %q (use keychain or bitwarden)", to)
		}
		if viper.GetString("gitlab.token") == "" {
			theme.SuccessMessage("No plaintext secrets to migrate")
			return nil
		}
		if err := config.MigratePlaintextToken(provider, item); err != nil {
			return err
		}
		theme.SuccessMessage(fmt.Sprintf("Migrated gitlab.token to %s", provider))
		log.Info("Plaintext key cleared from config")
		return nil
	},
}

func init() {
	SecretsMigrateCmd.Flags().String("to", "keychain", "secure store: keychain or bitwarden")
	SecretsMigrateCmd.Flags().String("item", "gitlab", "item/account name in the store")
	SecretsCmd.AddCommand(SecretsMigrateCmd)
}
