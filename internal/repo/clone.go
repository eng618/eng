package repo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"

	"github.com/eng618/eng/internal/log"
)

// Clone clones a git repository to the specified path.
// It uses go-git for the clone operation and provides informative error messages.
func Clone(ctx context.Context, url, destPath string) error {
	_, err := git.PlainCloneContext(ctx, destPath, false, &git.CloneOptions{
		URL:      url,
		Progress: log.Writer(),
	})
	if err != nil {
		errStr := strings.ToLower(err.Error())
		// Provide more helpful error messages for common issues
		switch {
		case errors.Is(err, git.ErrRepositoryAlreadyExists):
			return fmt.Errorf("repository already exists at %s", destPath)
		case strings.Contains(errStr, "authentication"):
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
			return err
		}
	}

	return nil
}
