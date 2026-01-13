package repo

import (
	"errors"
	"net/url"
	"os/exec"
	"regexp"
	"strings"
)

// GetGitLabHostAndProjectPath attempts to detect GitLab host and project path from the current repo remote.
// It returns host (e.g., gitlab.com) and project path (e.g., group/subgroup/repo).
func GetGitLabHostAndProjectPath(repoPath string) (string, string, error) {
	remoteURL, err := getOriginURL(repoPath)
	if err != nil || remoteURL == "" {
		return "", "", errors.New("no git remote origin url found")
	}
	return parseGitLabRemote(remoteURL)
}

func getOriginURL(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "config", "--get", "remote.origin.url")
	b, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func parseGitLabRemote(remote string) (string, string, error) {
	// Support SSH: git@gitlab.com:group/sub/repo.git
	sshPattern := regexp.MustCompile(`^(?:git|ssh)@([^:]+):(.+?)(?:\.git)?$`)
	if matches := sshPattern.FindStringSubmatch(remote); len(matches) == 3 {
		host := matches[1]
		path := strings.TrimSuffix(matches[2], ".git")
		return host, path, nil
	}
	// Support HTTPS: https://gitlab.com/group/sub/repo.git
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
