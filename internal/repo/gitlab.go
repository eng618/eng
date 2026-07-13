package repo

import (
	"context"
	"errors"
	"net/url"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
)

// GetGitLabHostAndProjectPath attempts to detect GitLab host and project path from the current repo remote.
// It returns host (e.g., gitlab.com) and project path (e.g., group/subgroup/repo).
func GetGitLabHostAndProjectPath(ctx context.Context, repoPath string) (string, string, error) {
	remoteURL, err := getOriginURL(ctx, repoPath)
	if err != nil || remoteURL == "" {
		return "", "", errors.New("no git remote origin url found")
	}
	return parseGitLabRemote(remoteURL)
}

func getOriginURL(ctx context.Context, repoPath string) (string, error) {
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		return "", err
	}
	cfg, err := r.Config()
	if err != nil {
		return "", err
	}
	remote, ok := cfg.Remotes["origin"]
	if !ok || len(remote.URLs) == 0 {
		return "", errors.New("no git remote origin url found")
	}
	return remote.URLs[0], nil
}

func parseGitLabRemote(remote string) (string, string, error) {
	// Support SSH: git@gitlab.com:group/sub/git.git
	sshPattern := regexp.MustCompile(`^(?:git|ssh)@([^:]+):(.+?)(?:\.git)?$`)
	if matches := sshPattern.FindStringSubmatch(remote); len(matches) == 3 {
		host := matches[1]
		path := strings.TrimSuffix(matches[2], ".git")
		return host, path, nil
	}
	// Support HTTPS: https://gitlab.com/group/sub/git.git
	if strings.HasPrefix(remote, "http://") || strings.HasPrefix(remote, "https://") {
		u, err := url.Parse(remote)
		if err != nil {
			return "", "", err
		}
		path := strings.TrimPrefix(u.Path, "/")
		path = strings.TrimSuffix(path, ".git")
		return u.Host, path, nil
	}
	return "", "", errors.New("unsupported remote url format")
}
