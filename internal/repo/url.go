package repo

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

// RepoNameFromURL extracts the repository name from a git URL.
// Supports both SSH (git@host:path/git.git) and HTTPS (https://host/path/git.git) formats.
func RepoNameFromURL(repoURL string) (string, error) {
	// SSH format: git@host:path/git.git or ssh://git@host/path/git.git
	sshPattern := regexp.MustCompile(`^(?:git|ssh)@[^:]+:(.+?)(?:\.git)?$`)
	if matches := sshPattern.FindStringSubmatch(repoURL); len(matches) == 2 {
		path := strings.TrimSuffix(matches[1], ".git")
		return filepath.Base(path), nil
	}

	// SSH with protocol: ssh://git@host/path/git.git
	sshProtoPattern := regexp.MustCompile(`^ssh://[^/]+/(.+?)(?:\.git)?$`)
	if matches := sshProtoPattern.FindStringSubmatch(repoURL); len(matches) == 2 {
		path := strings.TrimSuffix(matches[1], ".git")
		return filepath.Base(path), nil
	}

	// HTTPS format: https://host/path/git.git
	if strings.HasPrefix(repoURL, "http://") || strings.HasPrefix(repoURL, "https://") {
		u, err := url.Parse(repoURL)
		if err != nil {
			return "", fmt.Errorf("failed to parse URL: %w", err)
		}
		path := strings.TrimPrefix(u.Path, "/")
		path = strings.TrimSuffix(path, ".git")
		return filepath.Base(path), nil
	}

	return "", fmt.Errorf("unsupported URL format: %s", repoURL)
}

// EffectivePath returns the directory name for a repository: customPath when
// set, otherwise the name derived from the URL.
func EffectivePath(repoURL, customPath string) (string, error) {
	if customPath != "" {
		return customPath, nil
	}
	return RepoNameFromURL(repoURL)
}
