package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui/theme"
)

// MigrateCmd migrates the config file to the current schema version.
var MigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate the config file to the current schema version",
	Long: `Migrates legacy keys to the current schema, deletes renamed keys,
stamps the version, and writes a timestamped backup before changing anything.

Use --check to preview without writing.

Example: eng config migrate --check`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		check, _ := cmd.Flags().GetBool("check")
		path := config.ResolveConfigPath(viper.ConfigFileUsed())
		if check {
			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("read config: %w", err)
			}
			var decoded map[string]any
			if err := yaml.Unmarshal(data, &decoded); err != nil {
				return fmt.Errorf("parse config: %w", err)
			}
			_, changed, err := config.MigrateMap(decoded)
			if err != nil {
				return err
			}
			if unknown := config.FindUnknownKeys(path); len(unknown) > 0 {
				log.Warn("Unknown config keys (not read by any command): %s", strings.Join(unknown, ", "))
				log.Info("Drop them with `eng config unset <key>` after confirming nothing depends on them")
			}
			if changed {
				log.Info("Migration needed for %s (run without --check to apply)", path)
			} else {
				theme.SuccessMessage("Configuration is already at the latest schema")
			}
			return nil
		}
		backup, changed, err := config.MigrateFile(path)
		if err != nil {
			return err
		}
		if !changed {
			theme.SuccessMessage("Configuration is already at the latest schema")
			return nil
		}
		theme.SuccessMessage(fmt.Sprintf("Migrated to schema v%d", config.CurrentVersion))
		log.Info("Backup written to %s", backup)
		return nil
	},
}

func init() {
	MigrateCmd.Flags().Bool("check", false, "preview migration without writing")
}
