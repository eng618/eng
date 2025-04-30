// Package version implements the 'eng version' command, which displays
// the application's version information and checks for available updates.
package version

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/log"
	"github.com/spf13/cobra"
)

// Build-time variables
// These variables are populated during the build process using ldflags.
var (
	// Version holds the application's version string (e.g., "0.1.0" or "dev").
	Version = "dev"
	// Commit holds the Git commit hash from which the application was built.
	Commit = "none"
	// Date holds the build date of the application.
	Date = "unknown"
)

const (
	githubRepoOwner = "eng618"
	githubRepoName  = "eng"
	githubAPIURL    = "https://api.github.com/repos/%s/%s/releases/latest"
	requestTimeout  = 5 * time.Second // Timeout for the GitHub API request.
)

// githubReleaseInfo defines the structure for decoding the relevant fields
// from the GitHub API's latest release endpoint response.
type githubReleaseInfo struct {
	TagName string `json:"tag_name"` // The Git tag name of the release (e.g., "v0.1.0").
	HTMLURL string `json:"html_url"` // The URL to the release page on GitHub.
}

// VersionCmd represents the Cobra command for 'eng version'.
// It displays the current version details and checks GitHub for the latest release.
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of eng and check for updates",
	Long: `Displays the application's version, build commit, build date, Go version,
and target OS/Architecture.

It also checks the GitHub repository for the latest official release
and compares it with the currently running version.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("eng version: %s", Version)
		log.Message("  Git Commit: %s", Commit)
		log.Message("  Build Date: %s", Date)
		log.Message("  Go Version: %s", runtime.Version())
		log.Message("  OS/Arch:    %s/%s", runtime.GOOS, runtime.GOARCH)
		log.Message("") // Separator line

		sp := utils.NewSpinner("Checking for latest version...")
		sp.Start()

		latestRelease, err := getLatestRelease(githubRepoOwner, githubRepoName)

		// Stop the spinner *before* printing the results so the output is clean.
		sp.Stop()

		if err != nil {
			log.Warn("Could not check for updates: %v", err)
			return
		}

		// Handle cases where no release information was found (e.g., 404, empty response)
		if latestRelease == nil || latestRelease.TagName == "" {
			log.Warn("Could not determine the latest release version from GitHub.")
			return
		}

		// If running a development build, just show the latest release info.
		if Version == "dev" {
			log.Info("Currently running development version.")
			log.Info("Latest release is %s: %s", latestRelease.TagName, latestRelease.HTMLURL)
			return
		}

		// Attempt to parse the current version string as semantic version.
		currentSemVer, err := semver.NewVersion(Version)
		if err != nil {
			log.Warn("Could not parse current version (%s) for comparison: %v", Version, err)
			// Still provide info about the latest release even if comparison fails.
			log.Info("Latest release is %s: %s", latestRelease.TagName, latestRelease.HTMLURL)
			return
		}

		// Attempt to parse the latest release tag name as semantic version.
		// GitHub tags might not always be perfect semver (e.g., missing 'v'),
		// but NewVersion is generally robust enough.
		latestSemVer, err := semver.NewVersion(latestRelease.TagName)
		if err != nil {
			log.Warn("Could not parse latest release tag (%s) as semver: %v", latestRelease.TagName, err)
			// Log the raw tag name if parsing fails.
			log.Info("Latest release tag is: %s", latestRelease.TagName)
			log.Info("  Release page: %s", latestRelease.HTMLURL)
			return
		}

		// Compare the versions and report the result.
		if latestSemVer.GreaterThan(currentSemVer) {
			log.Success("A newer version is available: %s", latestRelease.TagName)
			log.Info("  Get it here: %s", latestRelease.HTMLURL)
		} else if latestSemVer.Equal(currentSemVer) {
			log.Success("You are running the latest version.")
		} else {
			// This case implies the current version is newer than the latest *stable* release
			// (e.g., running a pre-release or a local build from a newer commit).
			log.Info("You are running a version newer than the latest official release (%s).", latestRelease.TagName)
		}
	},
}

// getLatestRelease fetches the latest release information for a given GitHub repository.
// It sends a GET request to the GitHub API's 'latest release' endpoint.
// If the repository has no releases (404 Not Found), it returns (nil, nil).
// For other HTTP errors or decoding issues, it returns an error.
func getLatestRelease(owner, repo string) (release *githubReleaseInfo, err error) {
	url := fmt.Sprintf(githubAPIURL, owner, repo)

	client := &http.Client{
		Timeout: requestTimeout,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	// Request the specific GitHub API v3 format.
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}

	// Ensure the response body is closed, handling potential errors during close.
	defer func() {
		closeErr := resp.Body.Close()
		if err == nil && closeErr != nil {
			// Only assign closeErr if no other error occurred during the function.
			err = fmt.Errorf("failed to close response body: %w", closeErr)
		} else if closeErr != nil {
			// Log if an error occurs during close but another error already happened.
			log.Error("Error closing response body: %v", closeErr) // Use Debug for less critical errors
		}
	}()

	// Handle non-successful status codes.
	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusNotFound:
			// Treat "no releases found" as a valid state, not an error.
			return nil, nil
		case http.StatusForbidden:
			// Likely rate limiting or authentication issue.
			return nil, fmt.Errorf("github API request forbidden (status %d). Check rate limits or token permissions", resp.StatusCode)
		default:
			// Catch-all for other unexpected statuses.
			return nil, fmt.Errorf("unexpected status code %d from GitHub API", resp.StatusCode)
		}
	}

	// Decode the JSON response body into the githubReleaseInfo struct.
	var releaseInfo githubReleaseInfo
	if decodeErr := json.NewDecoder(resp.Body).Decode(&releaseInfo); decodeErr != nil {
		return nil, fmt.Errorf("failed to decode GitHub API response: %w", decodeErr)
	}

	// Basic validation: Ensure the tag name is present.
	if releaseInfo.TagName == "" {
		// This shouldn't typically happen for the /latest endpoint if status is 200 OK.
		return nil, fmt.Errorf("received success status but latest release tag name is empty")
	}

	return &releaseInfo, nil
}
