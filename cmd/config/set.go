package config

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui/theme"
)

// settableKeys maps user-facing dotted keys to viper keys.
var settableKeys = map[string]string{
	"email":                        "email",
	"verbose":                      "verbose",
	"git.dev_path":                 "git.dev_path",
	"git.editor":                   "git.editor",
	"dotfiles.repo_url":            "dotfiles.repo_url",
	"dotfiles.branch":              "dotfiles.branch",
	"dotfiles.bare_repo_path":      "dotfiles.bare_repo_path",
	"dotfiles.worktree_path":       "dotfiles.worktree_path",
	"dotfiles.target_repo_path":    "dotfiles.target_repo_path",
	"containers.path":              "containers.path",
	"antigravity.ide_download_url": "antigravity.ide_download_url",
}

// SetCmd sets a single config value, validating the field before write.
var SetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a single configuration value",
	Long: `Sets one configuration key and validates it before writing.

Keys: email, verbose, git.dev_path, git.editor, dotfiles.repo_url,
dotfiles.branch, dotfiles.bare_repo_path, dotfiles.worktree_path,
dotfiles.target_repo_path, containers.path, antigravity.ide_download_url.

Example: eng config set email you@example.com`,
	Args: cobra.ExactArgs(2),
	RunE: func(_ *cobra.Command, args []string) error {
		key, value := args[0], args[1]
		viperKey, ok := settableKeys[key]
		if !ok {
			return fmt.Errorf("unknown config key %q (run `eng config list` for known settings)", key)
		}
		// Single-field validation before write.
		if err := validateField(key, value); err != nil {
			return err
		}
		var stored any = value
		if viperKey == "verbose" {
			b, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("invalid verbose %q: use true or false", value)
			}
			stored = b
		}
		viper.Set(viperKey, stored)
		path := config.ResolveConfigPath(viper.ConfigFileUsed())
		if err := viper.WriteConfigAs(path); err != nil {
			return fmt.Errorf("write config: %w", err)
		}
		theme.SuccessMessage(fmt.Sprintf("Set %s", key))
		log.Info("Stored in %s", path)
		return nil
	},
}

// validateField runs the matching central-validation rule for one key.
func validateField(key, value string) error {
	rc := &config.ResolvedConfig{DotfilesBranch: "main"}
	switch key {
	case "email":
		rc.Email = value
	case "dotfiles.repo_url":
		rc.DotfilesRepo = value
	case "dotfiles.branch":
		rc.DotfilesBranch = value
	default:
		return nil
	}
	for _, fe := range rc.Validate() {
		return fmt.Errorf("%s", fe.Error())
	}
	return nil
}
