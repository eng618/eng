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

// These variables are set at build time using -ldflags
// They are exported so they can be used by root.go for the --version flag
var (
	Version = "dev"     // Default value if not built with ldflags
	Commit  = "none"    // Default value
	Date    = "unknown" // Default value
)

const (
	githubRepoOwner = "eng618"
	githubRepoName  = "eng"
	githubAPIURL    = "https://api.github.com/repos/%s/%s/releases/latest"
	requestTimeout  = 5 * time.Second // Add a timeout for the HTTP request
)

// Struct to decode the relevant part of the GitHub API response
type githubReleaseInfo struct {
	TagName string `json:"tag_name"` // We only need the tag name
	HTMLURL string `json:"html_url"` // URL to the release page
}

// VersionCmd represents the version command
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of eng and check for updates",
	Long: `All software has versions. This is eng's.
It shows the Git tag, commit hash, build date, Go version, OS/Arch,
and checks GitHub for the latest available release.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("eng version: %s", Version)
		log.Message("  Git Commit: %s", Commit)
		log.Message("  Build Date: %s", Date)
		log.Message("  Go Version: %s", runtime.Version())
		log.Message("  OS/Arch:    %s/%s", runtime.GOOS, runtime.GOARCH)
		log.Message("") // Add a newline for separation

		// --- Start Spinner ---
		// Create and start the spinner before the potentially long operation
		sp := utils.NewSpinner("Checking for latest version...")
		sp.Start()
		defer sp.Stop()
		// --- Spinner Started ---

		// Check for updates (this is the potentially slow part)
		latestRelease, err := getLatestRelease(githubRepoOwner, githubRepoName)

		// --- Process Results ---
		if err != nil {
			// Make sure spinner is stopped before printing error (defer handles this)
			log.Warn("Could not check for updates: %v", err)
			return // Exit after printing the warning
		}

		if latestRelease == nil || latestRelease.TagName == "" {
			// Make sure spinner is stopped (defer handles this)
			log.Warn("Could not determine latest release version.")
			return
		}

		// Compare versions if the current version is not "dev"
		if Version == "dev" {
			log.Info("Currently running development version.")
			log.Info("Latest release is %s: %s", latestRelease.TagName, latestRelease.HTMLURL)
			return
		}

		currentSemVer, err := semver.NewVersion(Version)
		if err != nil {
			log.Warn("Could not parse current version (%s) for comparison: %v", Version, err)
			log.Info("Latest release is %s: %s", latestRelease.TagName, latestRelease.HTMLURL) // Still show latest release info
			return
		}

		latestSemVer, err := semver.NewVersion(latestRelease.TagName)
		if err != nil {
			log.Warn("Could not parse latest release tag (%s) for comparison: %v", latestRelease.TagName, err)
			return
		}

		// Perform the comparison
		if latestSemVer.GreaterThan(currentSemVer) {
			log.Success("A newer version is available: %s", latestRelease.TagName)
			log.Info("  Get it here: %s", latestRelease.HTMLURL) // Info for the link
		} else if latestSemVer.Equal(currentSemVer) {
			log.Success("You are running the latest version.")
		} else {
			// This case might happen if running a pre-release or dev build newer than the latest stable
			log.Info("You are running a version newer than the latest release (%s).", latestRelease.TagName)
		}
	},
}

// getLatestRelease fetches the latest release information from GitHub API
func getLatestRelease(owner, repo string) (release *githubReleaseInfo, err error) {
	url := fmt.Sprintf(githubAPIURL, owner, repo)

	client := &http.Client{
		Timeout: requestTimeout,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	// Set Accept header for GitHub API v3
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		// Network errors, timeouts etc.
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}

	defer func() {
		closeErr := resp.Body.Close()
		// Only assign the close error if the function wasn't already returning an error
		if err == nil && closeErr != nil {
			err = fmt.Errorf("failed to close response body: %w", closeErr)
		}
		if closeErr != nil {
			log.Warn("Error closing response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		// Handle common cases like 404 Not Found (no releases) or 403 Forbidden (rate limit)
		if resp.StatusCode == http.StatusNotFound {
			// Changed this to return nil error but nil info, as "no releases" isn't strictly an error for the check
			return nil, nil // Indicate no release found, but not a technical error
		}
		if resp.StatusCode == http.StatusForbidden {
			return nil, fmt.Errorf("github API rate limit exceeded or access forbidden (status: %d)", resp.StatusCode)
		}
		return nil, fmt.Errorf("unexpected status code fetching latest release: %d", resp.StatusCode)
	}

	var releaseInfo githubReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&releaseInfo); err != nil {
		return nil, fmt.Errorf("failed to decode release response: %w", err)
	}

	// Handle case where API returns success but tag_name is empty (shouldn't happen for /latest)
	if releaseInfo.TagName == "" {
		return nil, fmt.Errorf("latest release found but tag name is empty")
	}

	return &releaseInfo, nil
}
