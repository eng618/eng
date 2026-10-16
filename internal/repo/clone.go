package repo

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// Clone clones a git repository to the specified path.
// It uses os/exec for the clone operation and provides informative error messages.
// On failure it removes a partially cloned destination (when this call created
// it) so retries start clean instead of tripping over half-written lock files.
func Clone(ctx context.Context, url, destPath string) error {
	created := false
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		created = true
	}
	cmd, cancel := execGitCloneCommand(ctx, "clone", "--progress", url, destPath)
	defer cancel()

	out, err := cmd.CombinedOutput()
	if err != nil {
		if created {
			_ = os.RemoveAll(destPath)
		}
		errStr := strings.ToLower(string(out))
		// Provide more helpful error messages for common issues
		switch {
		case strings.Contains(errStr, "already exists"):
			return fmt.Errorf("repository already exists at %s", destPath)
		case strings.Contains(errStr, "authentication") || strings.Contains(errStr, "permission denied"):
			return fmt.Errorf(
				"authentication failed - ensure your SSH keys are configured or use HTTPS with credentials: %w",
				err,
			)
		case strings.Contains(errStr, "could not read username"):
			return fmt.Errorf(
				"credentials required - for HTTPS URLs, configure git credential helper or use SSH: %w",
				err,
			)
		case strings.Contains(errStr, "ssh:"):
			return fmt.Errorf(
				"SSH error - ensure your SSH keys are loaded (ssh-add) and have access to the repository: %w",
				err,
			)
		default:
			return fmt.Errorf("git clone failed: %w\n%s", err, string(out))
		}
	}

	return nil
}
