package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/ui/theme"
)

// UnsetCmd removes a configuration key from the config file.
var UnsetCmd = &cobra.Command{
	Use:   "unset <key>",
	Short: "Remove a configuration key from the config file",
	Long: `Removes a top-level or dotted key from the config file entirely.
Useful for dropping orphan keys flagged by validate, e.g. project_filter.

Example: eng config unset project_filter`,
	Args: cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		key := args[0]
		path := config.ResolveConfigPath(viper.ConfigFileUsed())
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read config: %w", err)
		}
		var decoded map[string]any
		if err := yaml.Unmarshal(data, &decoded); err != nil {
			return fmt.Errorf("parse config: %w", err)
		}
		if !deleteKey(decoded, key) {
			return fmt.Errorf("key %q not present in %s", key, path)
		}
		out, err := yaml.Marshal(decoded)
		if err != nil {
			return fmt.Errorf("encode config: %w", err)
		}
		if err := os.WriteFile(path, out, 0o600); err != nil {
			return fmt.Errorf("write config: %w", err)
		}
		theme.SuccessMessage(fmt.Sprintf("Removed %s", key))
		return nil
	},
}

// deleteKey removes a dotted key from m, pruning empty parent maps.
func deleteKey(m map[string]any, dotted string) bool {
	parts := strings.Split(dotted, ".")
	cur := m
	for _, p := range parts[:len(parts)-1] {
		next, ok := cur[p].(map[string]any)
		if !ok {
			return false
		}
		cur = next
	}
	last := parts[len(parts)-1]
	if _, ok := cur[last]; !ok {
		return false
	}
	delete(cur, last)
	return true
}
