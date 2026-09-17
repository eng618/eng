package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/eng618/eng/internal/paths"
)

// EnvPrefix is the environment variable prefix for config overrides.
// Dots in keys map to underscores: ENG_GIT_DEV_PATH overrides git.dev_path.
const EnvPrefix = "ENG"

// ResolvedConfig is the fully-loaded, expanded, validated view of configuration.
type ResolvedConfig struct {
	Version        int
	Email          string
	Verbose        bool
	GitDevPath     string
	GitEditor      string
	DotfilesRepo   string
	DotfilesBranch string
	ContainersPath string
	Source         map[string]string
}

// DefaultConfigPath returns the primary config file path ($HOME/.eng.yaml).
func DefaultConfigPath() string {
	return filepath.Join(paths.MustHome(), ".eng.yaml")
}

// XDGConfigPath returns the XDG fallback path, or "" when XDG_CONFIG_HOME is unset.
func XDGConfigPath() string {
	xdg := os.Getenv("XDG_CONFIG_HOME")
	if xdg == "" {
		return ""
	}
	return filepath.Join(xdg, "eng", "config.yaml")
}

// ResolveConfigPath returns the config file to use: explicit flag path,
// then XDG fallback if it exists, else the default home path.
func ResolveConfigPath(explicit string) string {
	if explicit != "" {
		return explicit
	}
	if xdg := XDGConfigPath(); xdg != "" {
		if _, err := os.Stat(xdg); err == nil {
			return xdg
		}
	}
	return DefaultConfigPath()
}

// NewLoader returns an isolated viper instance with defaults, ENG_ env
// bindings, and the given config file preselected. It never touches the
// global viper singleton, so tests can load configs in parallel.
func NewLoader(configFile string) *viper.Viper {
	v := viper.New()
	v.SetConfigType("yaml")
	if configFile != "" {
		v.SetConfigFile(configFile)
	}
	v.SetEnvPrefix(EnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	v.SetDefault("version", CurrentVersion)
	v.SetDefault("verbose", false)
	v.SetDefault("email", "")
	v.SetDefault("git.dev_path", filepath.Join(paths.MustHome(), "Development"))
	v.SetDefault("git.editor", "")
	v.SetDefault("dotfiles.branch", "main")
	v.SetDefault("dotfiles.bare_repo_path", filepath.Join(paths.MustHome(), ".eng-cfg"))
	v.SetDefault("dotfiles.worktree_path", paths.MustHome())
	v.SetDefault("containers.path", "")
	v.SetDefault("projects", []any{})
	return v
}

// LoadResolved reads the config file (if present), unmarshals exactly into
// known keys, expands all paths once, and records the value source per key.
func LoadResolved(v *viper.Viper) (*ResolvedConfig, error) {
	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return nil, err
		}
	}
	var raw struct {
		Version int    `mapstructure:"version"`
		Email   string `mapstructure:"email"`
		Verbose bool   `mapstructure:"verbose"`
		Git     struct {
			DevPath string `mapstructure:"dev_path"`
			Editor  string `mapstructure:"editor"`
		} `mapstructure:"git"`
		Dotfiles struct {
			RepoURL        string `mapstructure:"repo_url"`
			Branch         string `mapstructure:"branch"`
			BareRepoPath   string `mapstructure:"bare_repo_path"`
			WorktreePath   string `mapstructure:"worktree_path"`
			TargetRepoPath string `mapstructure:"target_repo_path"`
		} `mapstructure:"dotfiles"`
		Containers struct {
			Path string `mapstructure:"path"`
		} `mapstructure:"containers"`
		Projects    []any          `mapstructure:"projects"`
		Telemetry   map[string]any `mapstructure:"telemetry"`
		Gitlab      map[string]any `mapstructure:"gitlab"`
		Proxy       map[string]any `mapstructure:"proxy"`
		Antigravity map[string]any `mapstructure:"antigravity"`
	}
	if err := v.Unmarshal(&raw); err != nil {
		return nil, err
	}
	if err := rejectUnknownTopLevelKeys(v); err != nil {
		return nil, err
	}
	rc := &ResolvedConfig{
		Version:        raw.Version,
		Email:          raw.Email,
		Verbose:        raw.Verbose,
		GitDevPath:     paths.Expand(raw.Git.DevPath),
		GitEditor:      raw.Git.Editor,
		DotfilesRepo:   raw.Dotfiles.RepoURL,
		DotfilesBranch: raw.Dotfiles.Branch,
		ContainersPath: paths.Expand(raw.Containers.Path),
		Source:         map[string]string{},
	}
	fileKeys := fileKeySet(v.ConfigFileUsed())
	for _, key := range []string{
		"email", "verbose", "git.dev_path", "git.editor",
		"dotfiles.repo_url", "dotfiles.branch", "dotfiles.bare_repo_path",
		"containers.path",
	} {
		rc.Source[key] = valueSource(fileKeys, key)
	}
	return rc, nil
}

// knownTopLevelKeys is the allowlist for typo detection in config files.
var knownTopLevelKeys = map[string]bool{
	"version": true, "email": true, "verbose": true, "git": true,
	"dotfiles": true, "containers": true, "projects": true,
	"telemetry": true, "gitlab": true, "proxy": true, "proxies": true,
	"antigravity": true,
	"user-email":  true, // legacy, migrated by MigrateMap
}

// rejectUnknownTopLevelKeys fails fast on typos like `gti:` instead of `git:`.
// Legacy keys pass; only truly unknown keys are rejected.
func rejectUnknownTopLevelKeys(v *viper.Viper) error {
	for _, key := range v.AllKeys() {
		top := key
		if i := strings.Index(key, "."); i >= 0 {
			top = key[:i]
		}
		if !knownTopLevelKeys[top] {
			return fmt.Errorf(
				"unknown config key %q (top-level %q not recognized): "+
					"remove it with `eng config unset %s` or delete it from the config file",
				key,
				top,
				top,
			)
		}
	}
	return nil
}

func valueSource(fileKeys map[string]bool, key string) string {
	envKey := EnvPrefix + "_" + strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
	if _, ok := os.LookupEnv(envKey); ok {
		return "env"
	}
	if fileKeys[key] {
		return "file"
	}
	return "default"
}

// fileKeySet returns the set of dotted keys present in the YAML config file.
func fileKeySet(path string) map[string]bool {
	out := map[string]bool{}
	if path == "" {
		return out
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return out
	}
	var decoded map[string]any
	if err := yaml.Unmarshal(data, &decoded); err != nil {
		return out
	}
	flattenKeys(decoded, "", out)
	return out
}

func flattenKeys(m map[string]any, prefix string, out map[string]bool) {
	for k, v := range m {
		dotted := k
		if prefix != "" {
			dotted = prefix + "." + k
		}
		out[dotted] = true
		if nested, ok := v.(map[string]any); ok {
			flattenKeys(nested, dotted, out)
		}
	}
}
