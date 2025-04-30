// Package version implements the 'eng version' command, which displays
// the application's version information and checks for available updates.
package version

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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
	brewCmd         = "brew"          // Command for Homebrew
	brewPkgName     = "eng"           // Package name in Homebrew
)

// Flag variable for the --update flag
var updateFlag bool

// githubReleaseInfo defines the structure for decoding the relevant fields
// from the GitHub API's latest release endpoint response.
type githubReleaseInfo struct {
	TagName string `json:"tag_name"` // The Git tag name of the release (e.g., "v0.1.0").
	HTMLURL string `json:"html_url"` // The URL to the release page on GitHub.
}

// VersionCmd represents the Cobra command for 'eng version'.
// It displays the current version details and checks GitHub for the latest release.
// Includes an optional --update flag to attempt an update via Homebrew.
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of eng and check for updates",
	Long: `Displays the application's version, build commit, build date, Go version,
and target OS/Architecture.

It also checks the GitHub repository (eng618/eng) for the latest official release
and compares it with the currently running version.

If a newer version is available and eng was installed via Homebrew,
you can use the --update flag to attempt an automatic upgrade.`,
	Run: func(cmd *cobra.Command, args []string) {
		isVerbose, _ := cmd.Flags().GetBool("verbose")

		printVersionInfo()

		sp := utils.NewSpinner("Checking for latest version...")
		sp.Start()
		latestRelease, err := getLatestRelease(githubRepoOwner, githubRepoName, isVerbose)
		sp.Stop() // Stop spinner before printing results or attempting update

		if err != nil {
			log.Warn("Could not check for updates: %v", err)
			return
		}

		if latestRelease == nil || latestRelease.TagName == "" {
			log.Warn("Could not determine the latest release version from GitHub.")
			return
		}

		if Version == "dev" {
			handleDevVersion(latestRelease)
			return
		}

		currentSemVer, latestSemVer, err := parseVersions(Version, latestRelease.TagName)
		if err != nil {
			// Errors are logged within parseVersions, just provide latest release info
			log.Info("Latest release is %s: %s", latestRelease.TagName, latestRelease.HTMLURL)
			return
		}

		compareAndHandleUpdate(currentSemVer, latestSemVer, latestRelease, isVerbose)
	},
}

// init registers the command and its flags.
func init() {
	// Add the --update flag
	VersionCmd.Flags().BoolVarP(&updateFlag, "update", "u", false, "Attempt to update eng to the latest version (requires Homebrew)")
	// Note: You would typically add VersionCmd to your root command in cmd/root.go
	// Example: rootCmd.AddCommand(version.VersionCmd)
}

// printVersionInfo displays the static build and runtime information.
func printVersionInfo() {
	log.Info("eng version: %s", Version)
	log.Message("  Git Commit: %s", Commit)
	log.Message("  Build Date: %s", Date)
	log.Message("  Go Version: %s", runtime.Version())
	log.Message("  OS/Arch:    %s/%s", runtime.GOOS, runtime.GOARCH)
	log.Message("") // Separator line
}

// handleDevVersion logs information when running a development version.
func handleDevVersion(latestRelease *githubReleaseInfo) {
	log.Info("Currently running development version.")
	log.Info("Latest official release is %s: %s", latestRelease.TagName, latestRelease.HTMLURL)
	if updateFlag {
		log.Info("--update flag ignored when running a dev version.")
	}
}

// parseVersions attempts to parse both current and latest versions using semver.
// It logs warnings if parsing fails.
func parseVersions(currentVerStr, latestTagStr string) (current, latest *semver.Version, err error) {
	current, err = semver.NewVersion(currentVerStr)
	if err != nil {
		log.Warn("Could not parse current version (%s) for comparison: %v", currentVerStr, err)
		return nil, nil, err // Return error to signal failure
	}

	latest, err = semver.NewVersion(latestTagStr)
	if err != nil {
		log.Warn("Could not parse latest release tag (%s) as semver: %v", latestTagStr, err)
		// Log raw tag info if parsing fails, but don't return an error here,
		// as we might still want to show the current version status.
		// The calling function will handle the nil latest version.
		return current, nil, err // Return error to signal failure
	}
	return current, latest, nil
}

// compareAndHandleUpdate compares versions and handles the update logic if requested.
func compareAndHandleUpdate(currentSemVer, latestSemVer *semver.Version, latestRelease *githubReleaseInfo, isVerbose bool) {
	brewDetected := isBrewInstallation(isVerbose)

	if latestSemVer.GreaterThan(currentSemVer) {
		log.Success("A newer version is available: %s", latestRelease.TagName)

		if updateFlag {
			if brewDetected {
				log.Info("Attempting update via Homebrew...")
				err := runBrewUpgrade(isVerbose)
				if err != nil {
					log.Error("Brew update failed: %v", err)
					log.Info("  Please try manually: %s upgrade %s", brewCmd, brewPkgName)
					log.Info("  Or get it from GitHub: %s", latestRelease.HTMLURL)
				} else {
					log.Success("Update via %s successful!", brewCmd)
					// Optionally: You could re-verify the version here, but exiting is simpler.
				}
				// Exit after attempting update, regardless of success/failure
				// to avoid printing redundant "Get it here" messages.
				return
			} else {
				log.Warn("--update flag specified, but cannot automatically update.")
				log.Info("  Installation method not recognized as Homebrew.")
				log.Info("  Try updating manually with: go install %s/%s@latest", githubRepoOwner, githubRepoName)
				log.Info("  Or get it from GitHub: %s", latestRelease.HTMLURL)
			}
		} else {
			// Just inform the user how to update
			if brewDetected {
				log.Info("  Run `eng version --update` or `eng version -u` to attempt an automatic update.")
			} else { // If not brew detected, suggest go install
				log.Info("  Try updating with: go install %s/%s@latest", githubRepoOwner, githubRepoName)
			}
			log.Info("  Or get it manually here: %s", latestRelease.HTMLURL)
		}
	} else if latestSemVer.Equal(currentSemVer) {
		log.Success("You are running the latest version.")
		if updateFlag {
			log.Info("--update flag specified, but no newer version is available.")
		}
	} else {
		// Current version is newer than the latest official release
		log.Info("You are running a version newer than the latest official release (%s).", latestRelease.TagName)
		if updateFlag {
			log.Info("--update flag specified, but you are already running a newer version.")
		}
	}
}

// isBrewInstallation checks if the executable path suggests a Homebrew installation.
// This is a heuristic and might not cover all edge cases or future Brew changes.
func isBrewInstallation(isVerbose bool) bool {
	executablePath, err := os.Executable()
	if err != nil {
		log.Verbose(isVerbose, "Could not get executable path: %v", err)
		return false
	}

	// Resolve symlinks, as brew often uses them (e.g., /usr/local/bin/eng -> ../Cellar/eng/0.1.0/bin/eng)
	resolvedPath, err := filepath.EvalSymlinks(executablePath)
	if err != nil {
		// If symlink resolution fails, use the original path
		resolvedPath = executablePath
		log.Verbose(isVerbose, "Could not resolve symlink for executable path: %v", err)
	}

	// Common Homebrew installation prefixes:
	// - macOS Intel: /usr/local/Cellar
	// - macOS Apple Silicon: /opt/homebrew/Cellar
	// - Linux (Linuxbrew): /home/linuxbrew/.linuxbrew/Cellar
	brewPrefixes := []string{"/usr/local/Cellar", "/opt/homebrew/Cellar", "/home/linuxbrew/.linuxbrew/Cellar"}


	for _, prefix := range brewPrefixes {
		if strings.HasPrefix(resolvedPath, prefix) {
			log.Verbose(isVerbose, "Detected Homebrew installation path: %s", resolvedPath)
			// Check if 'brew' command actually exists for higher confidence
			_, err := exec.LookPath(brewCmd)
			if err != nil {
				log.Verbose(isVerbose, "Executable path looks like Brew, but '%s' command not found in PATH.", brewCmd)
				return false // Path looks right, but brew command missing? Be cautious.
			}
			return true
		}
	}

	log.Verbose(isVerbose, "Executable path does not match known Homebrew prefixes: %s", resolvedPath)
	return false
}

// runBrewUpgrade executes the 'brew upgrade eng' command.
// It first runs 'brew update' to refresh formula information (including taps),
// then runs 'brew upgrade eng'. It streams the commands' output directly
// to the user's terminal, respecting verbosity for command logging.
func runBrewUpgrade(isVerbose bool) error {
	// Step 1: Update brew formula information
	log.Info("Running '%s update'...", brewCmd)
	updateCmd := exec.Command(brewCmd, "update")
	updateCmd.Stdout = log.Writer()
	updateCmd.Stderr = log.ErrorWriter()
	log.Verbose(isVerbose, "Executing command: %s", updateCmd.String())
	err := updateCmd.Run()
	if err != nil {
		// Don't necessarily fail the whole process if 'brew update' has minor issues,
		// but log it. The subsequent upgrade might still work.
		log.Warn("'%s update' command finished with error (proceeding with upgrade attempt): %v", brewCmd, err)
	}

	// Step 2: Upgrade the specific package
	log.Info("Running '%s upgrade %s'...", brewCmd, brewPkgName)
	upgradeCmd := exec.Command(brewCmd, "upgrade", brewPkgName)
	upgradeCmd.Stdout = log.Writer()
	upgradeCmd.Stderr = log.ErrorWriter()
	log.Verbose(isVerbose, "Executing command: %s", upgradeCmd.String())
	err = upgradeCmd.Run()
	if err != nil {
		// Return the error from the upgrade command specifically
		return fmt.Errorf("'%s upgrade %s' command failed: %w", brewCmd, brewPkgName, err)
	}
	return nil
}

// getLatestRelease fetches the latest release information for a given GitHub repository.
func getLatestRelease(owner, repo string, isVerbose bool) (release *githubReleaseInfo, err error) {
	url := fmt.Sprintf(githubAPIURL, owner, repo)
	client := &http.Client{Timeout: requestTimeout}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer func() {
		closeErr := resp.Body.Close()
		if err == nil && closeErr != nil {
			err = fmt.Errorf("failed to close response body: %w", closeErr)
		} else if closeErr != nil {
			log.Verbose(isVerbose, "Error closing response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusNotFound:
			return nil, nil // No releases found is not an error here
		case http.StatusForbidden:
			return nil, fmt.Errorf("github API request forbidden (status %d). Check rate limits or token permissions", resp.StatusCode)
		default:
			return nil, fmt.Errorf("unexpected status code %d from GitHub API", resp.StatusCode)
		}
	}

	var releaseInfo githubReleaseInfo
	if decodeErr := json.NewDecoder(resp.Body).Decode(&releaseInfo); decodeErr != nil {
		return nil, fmt.Errorf("failed to decode GitHub API response: %w", decodeErr)
	}
	if releaseInfo.TagName == "" {
		return nil, fmt.Errorf("received success status but latest release tag name is empty")
	}

	return &releaseInfo, nil
}
