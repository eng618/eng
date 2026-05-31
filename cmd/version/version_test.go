package version

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Masterminds/semver/v3"

	"github.com/eng618/eng/internal/utils/log"
)

func TestPrintVersionInfo(t *testing.T) {
	var outBuf bytes.Buffer
	log.SetWriters(&outBuf, &outBuf)
	defer log.ResetWriters()

	// Backup original values and restore them after the test
	origVersion := Version
	origCommit := Commit
	origDate := Date
	defer func() {
		Version = origVersion
		Commit = origCommit
		Date = origDate
	}()

	Version = "1.2.3"
	Commit = "abcdef"
	Date = "2023-10-27"

	printVersionInfo()

	output := outBuf.String()

	if !strings.Contains(output, "eng version: 1.2.3") {
		t.Errorf("Expected output to contain 'eng version: 1.2.3', got: %s", output)
	}
	if !strings.Contains(output, "Git Commit: abcdef") {
		t.Errorf("Expected output to contain 'Git Commit: abcdef', got: %s", output)
	}
	if !strings.Contains(output, "Build Date: 2023-10-27") {
		t.Errorf("Expected output to contain 'Build Date: 2023-10-27', got: %s", output)
	}
}

func TestParseVersions(t *testing.T) {
	tests := []struct {
		name          string
		currentVer    string
		latestTag     string
		expectErr     bool
		expectGreater bool // latest > current
	}{
		{
			name:          "Valid versions, latest is newer",
			currentVer:    "1.0.0",
			latestTag:     "v1.1.0",
			expectErr:     false,
			expectGreater: true,
		},
		{
			name:          "Valid versions, latest is older",
			currentVer:    "1.1.0",
			latestTag:     "v1.0.0",
			expectErr:     false,
			expectGreater: false,
		},
		{
			name:          "Valid versions, same version",
			currentVer:    "1.0.0",
			latestTag:     "v1.0.0",
			expectErr:     false,
			expectGreater: false, // not greater, they are equal
		},
		{
			name:          "Invalid current version",
			currentVer:    "invalid",
			latestTag:     "v1.0.0",
			expectErr:     true,
			expectGreater: false,
		},
		{
			name:          "Invalid latest version",
			currentVer:    "1.0.0",
			latestTag:     "invalid",
			expectErr:     true,
			expectGreater: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current, latest, err := parseVersions(tt.currentVer, tt.latestTag)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if current == nil || latest == nil {
				t.Errorf("Expected versions to not be nil")
				return
			}

			isGreater := latest.GreaterThan(current)
			if isGreater != tt.expectGreater {
				t.Errorf("Expected latest > current to be %v, got %v", tt.expectGreater, isGreater)
			}
		})
	}
}

func TestGetLatestRelease(t *testing.T) {
	// Setup mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/eng618/eng/releases/latest" {
			t.Errorf("Expected path '/repos/eng618/eng/releases/latest', got '%s'", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		resp := githubReleaseInfo{
			TagName: "v1.2.3",
			HTMLURL: "https://github.com/eng618/eng/releases/tag/v1.2.3",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	// Backup original URL and restore after test
	origGithubAPIURL := githubAPIURL
	defer func() { githubAPIURL = origGithubAPIURL }()

	// Override URL to use mock server
	githubAPIURL = ts.URL + "/repos/%s/%s/releases/latest"

	release, err := getLatestRelease("eng618", "eng", false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if release == nil {
		t.Fatalf("Expected release info, got nil")
	}

	if release.TagName != "v1.2.3" {
		t.Errorf("Expected TagName 'v1.2.3', got '%s'", release.TagName)
	}
}

func TestCompareAndHandleUpdate(t *testing.T) {
	var outBuf bytes.Buffer
	log.SetWriters(&outBuf, &outBuf)
	defer log.ResetWriters()

	currentSemVer, _ := semver.NewVersion("1.0.0")
	latestSemVer, _ := semver.NewVersion("1.1.0")

	release := &githubReleaseInfo{
		TagName: "v1.1.0",
		HTMLURL: "https://github.com/eng618/eng/releases/tag/v1.1.0",
	}

	updateFlag = false // Ensure updateFlag is false initially

	compareAndHandleUpdate(currentSemVer, latestSemVer, release, false)

	output := outBuf.String()
	if !strings.Contains(output, "A newer version is available: v1.1.0") {
		t.Errorf("Expected output to mention newer version, got: %s", output)
	}
}
