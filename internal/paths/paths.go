// Package paths centralizes home-directory resolution and path expansion.
//
// All home-dir lookups should go through this package instead of calling
// os.UserHomeDir directly, so tests can stub UserHomeDir and behavior
// stays consistent (tilde + env expansion) across commands.
package paths

import (
	"os"
	"path/filepath"
	"strings"
)

// UserHomeDir is mockable for testing. Defaults to os.UserHomeDir.
var UserHomeDir = os.UserHomeDir

// Home returns the current user's home directory.
func Home() (string, error) {
	return UserHomeDir()
}

// MustHome returns the home directory or an empty string on error.
// Use when a best-effort fallback path is acceptable (e.g. display defaults).
func MustHome() string {
	home, _ := UserHomeDir()
	return home
}

// JoinHome joins elem onto the home directory.
func JoinHome(elem ...string) (string, error) {
	home, err := UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(append([]string{home}, elem...)...), nil
}

// Expand expands a leading "~" and environment variables in path.
// "~" alone maps to home; "~/x" maps to $HOME/x. Other values pass
// through os.ExpandEnv unchanged. On home lookup failure the input
// is returned with env expansion only.
func Expand(path string) string {
	if path == "" {
		return ""
	}
	if path == "~" || strings.HasPrefix(path, "~/") {
		if home, err := UserHomeDir(); err == nil {
			if path == "~" {
				return home
			}
			return filepath.Join(home, strings.TrimPrefix(path, "~/"))
		}
	}
	return os.ExpandEnv(path)
}
