package config

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
)

// GetCmd prints a single config value with its source.
var GetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Print a single configuration value and where it comes from",
	Long: `Prints one configuration value with its source (env, file, or default).

Example: eng config get git.dev_path`,
	Args: cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		key := args[0]
		if _, ok := settableKeys[key]; !ok {
			return fmt.Errorf("unknown config key %q (run `eng config list` for known settings)", key)
		}
		v := config.NewLoader(config.ResolveConfigPath(viper.ConfigFileUsed()))
		rc, err := config.LoadResolved(v)
		if err != nil {
			return err
		}
		log.Message("%s=%s (source: %s)", key, resolvedValue(rc, key), rc.Source[key])
		return nil
	},
}

func resolvedValue(rc *config.ResolvedConfig, key string) string {
	switch key {
	case "email":
		return rc.Email
	case "verbose":
		if rc.Verbose {
			return "true"
		}
		return "false"
	case "git.dev_path":
		return rc.GitDevPath
	case "git.editor":
		return rc.GitEditor
	case "dotfiles.repo_url":
		return rc.DotfilesRepo
	case "dotfiles.branch":
		return rc.DotfilesBranch
	case "containers.path":
		return rc.ContainersPath
	default:
		return viper.GetString(settableKeys[key])
	}
}
