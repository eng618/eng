package repo

import (
	"testing"
)

func TestParseGitLabRemote(t *testing.T) {
	tests := []struct {
		name        string
		remote      string
		wantHost    string
		wantPath    string
		wantErr     bool
		wantErrText string
	}{
		// SSH Cases
		{
			name:     "SSH standard git user",
			remote:   "git@gitlab.com:group/sub/repo.git",
			wantHost: "gitlab.com",
			wantPath: "group/sub/repo",
			wantErr:  false,
		},
		{
			name:     "SSH ssh user",
			remote:   "ssh@gitlab.com:group/sub/repo.git",
			wantHost: "gitlab.com",
			wantPath: "group/sub/repo",
			wantErr:  false,
		},
		{
			name:     "SSH without .git suffix",
			remote:   "git@gitlab.com:group/sub/repo",
			wantHost: "gitlab.com",
			wantPath: "group/sub/repo",
			wantErr:  false,
		},
		{
			name:     "SSH with custom host",
			remote:   "git@custom.gitlab.com:group/repo.git",
			wantHost: "custom.gitlab.com",
			wantPath: "group/repo",
			wantErr:  false,
		},

		// HTTPS Cases
		{
			name:     "HTTPS standard",
			remote:   "https://gitlab.com/group/sub/repo.git",
			wantHost: "gitlab.com",
			wantPath: "group/sub/repo",
			wantErr:  false,
		},
		{
			name:     "HTTP standard",
			remote:   "http://gitlab.com/group/sub/repo.git",
			wantHost: "gitlab.com",
			wantPath: "group/sub/repo",
			wantErr:  false,
		},
		{
			name:     "HTTPS without .git suffix",
			remote:   "https://gitlab.com/group/sub/repo",
			wantHost: "gitlab.com",
			wantPath: "group/sub/repo",
			wantErr:  false,
		},
		{
			name:     "HTTPS with custom host",
			remote:   "https://custom.gitlab.com/group/repo.git",
			wantHost: "custom.gitlab.com",
			wantPath: "group/repo",
			wantErr:  false,
		},

		// Invalid Cases
		{
			name:        "Empty string",
			remote:      "",
			wantErr:     true,
			wantErrText: "unsupported remote url format",
		},
		{
			name:        "Unsupported protocol",
			remote:      "ftp://gitlab.com/group/repo.git",
			wantErr:     true,
			wantErrText: "unsupported remote url format",
		},
		{
			name:        "Invalid URL format",
			remote:      "https://::invalid",
			wantErr:     true,
			wantErrText: "parse \"https://::invalid\": invalid port \"::invalid\" after host",
		},
		{
			name:        "Random string",
			remote:      "just some random text",
			wantErr:     true,
			wantErrText: "unsupported remote url format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHost, gotPath, err := parseGitLabRemote(tt.remote)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseGitLabRemote() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.wantErrText != "" {
				if err.Error() != tt.wantErrText {
					t.Errorf("parseGitLabRemote() error text = %q, wantErrText %q", err.Error(), tt.wantErrText)
				}
			}

			if gotHost != tt.wantHost {
				t.Errorf("parseGitLabRemote() gotHost = %v, want %v", gotHost, tt.wantHost)
			}
			if gotPath != tt.wantPath {
				t.Errorf("parseGitLabRemote() gotPath = %v, want %v", gotPath, tt.wantPath)
			}
		})
	}
}
