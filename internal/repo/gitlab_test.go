package repo

import (
	"context"
	"os"
	"testing"

	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
)

func TestParseGitLabRemote(t *testing.T) {
	tests := []struct {
		name         string
		remote       string
		expectedHost string
		expectedPath string
		expectErr    bool
	}{
		{
			name:         "Standard SSH URL with .git",
			remote:       "git@gitlab.com:group/project.git",
			expectedHost: "gitlab.com",
			expectedPath: "group/project",
			expectErr:    false,
		},
		{
			name:         "Standard SSH URL without .git",
			remote:       "git@gitlab.com:group/project",
			expectedHost: "gitlab.com",
			expectedPath: "group/project",
			expectErr:    false,
		},
		{
			name:         "SSH URL with custom domain and subgroups",
			remote:       "git@gitlab.company.org:group/subgroup/project.git",
			expectedHost: "gitlab.company.org",
			expectedPath: "group/subgroup/project",
			expectErr:    false,
		},
		{
			name:         "Standard HTTPS URL with .git",
			remote:       "https://gitlab.com/group/project.git",
			expectedHost: "gitlab.com",
			expectedPath: "group/project",
			expectErr:    false,
		},
		{
			name:         "Standard HTTPS URL without .git",
			remote:       "https://gitlab.com/group/project",
			expectedHost: "gitlab.com",
			expectedPath: "group/project",
			expectErr:    false,
		},
		{
			name:         "HTTPS URL with custom domain and subgroups",
			remote:       "https://my-gitlab.org/group/subgroup/sub/project.git",
			expectedHost: "my-gitlab.org",
			expectedPath: "group/subgroup/sub/project",
			expectErr:    false,
		},
		{
			name:         "Unsupported URL format",
			remote:       "/var/local/git/project.git",
			expectedHost: "",
			expectedPath: "",
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, path, err := parseGitLabRemote(tt.remote)
			if tt.expectErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if host != tt.expectedHost {
					t.Errorf("expected host %q, got %q", tt.expectedHost, host)
				}
				if path != tt.expectedPath {
					t.Errorf("expected path %q, got %q", tt.expectedPath, path)
				}
			}
		})
	}
}

func TestGetGitLabHostAndProjectPath(t *testing.T) {
	// Create temporary directory for git repo
	tmpDir, err := os.MkdirTemp("", "gitlab-test-repo-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test on non-initialized repo first (should fail)
	_, _, err = GetGitLabHostAndProjectPath(context.Background(), tmpDir)
	if err == nil {
		t.Error("expected error for non-initialized repo, got nil")
	}

	// Initialize git repo
	r, err := git.PlainInit(tmpDir, false)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	// Test on repo with no remotes (should fail)
	_, _, err = GetGitLabHostAndProjectPath(context.Background(), tmpDir)
	if err == nil {
		t.Error("expected error for repo with no remotes, got nil")
	}

	// Add origin remote
	_, err = r.CreateRemote(&gitconfig.RemoteConfig{
		Name: "origin",
		URLs: []string{"git@gitlab.com:myorg/myproject.git"},
	})
	if err != nil {
		t.Fatalf("failed to add remote: %v", err)
	}

	// Test with valid GitLab remote
	host, path, err := GetGitLabHostAndProjectPath(context.Background(), tmpDir)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if host != "gitlab.com" {
		t.Errorf("expected host %q, got %q", "gitlab.com", host)
	}
	if path != "myorg/myproject" {
		t.Errorf("expected path %q, got %q", "myorg/myproject", path)
	}
}
