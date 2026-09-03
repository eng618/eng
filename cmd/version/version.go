// Package version implements the 'eng version' command, which displays
// the application's version information and checks for available updates.
package version

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
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
	requestTimeout  = 5 * time.Second // Timeout for the GitHub API request.
	brewCmd         = "brew"          // Command for Homebrew
	brewPkgName     = "eng"           // Package name in Homebrew
	scriptTimeout   = 60 * time.Second
)

var (
	githubAPIURL = "https://api.github.com/repos/%s/%s/releases/latest"
	// installScriptURL is the canonical install.sh used for curl installs and
	// script-based self-updates (overridable in tests).
	installScriptURL = "https://raw.githubusercontent.com/eng618/eng/main/install.sh"
	execCommand      = exec.Command
	osExecutable     = os.Executable
	evalSymlinks     = filepath.EvalSymlinks
	lookPath         = exec.LookPath
)

// Flag variable for the --update flag.
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

If a newer version is available, you can use the --update flag to attempt
an automatic upgrade via Homebrew (if installed that way) or via the
install script (if installed via curl).`,
	Run: func(cmd *cobra.Command, _args []string) {
		isVerbose, _ := cmd.Flags().GetBool("verbose")

		printVersionInfo(isVerbose)

		var sp *ui.Spinner
		if !ui.DisableProgress {
			sp = ui.NewSpinner("Checking for latest version...")
			sp.Start()
		}
		latestRelease, err := getLatestRelease(githubRepoOwner, githubRepoName, isVerbose)
		if sp != nil {
			sp.Stop() // Stop spinner before printing results or attempting update
		}

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
	VersionCmd.Flags().
		BoolVarP(&updateFlag, "update", "u", false, "Attempt to update eng to the latest version (Homebrew or install script)")

	// Note: You would typically add VersionCmd to your root command in cmd/root.go
	// Example: rootCmd.AddCommand(version.VersionCmd)
}

// getInstallSource returns a human-readable install source based on the
// executable path heuristic.
func getInstallSource(isVerbose bool) string {
	if isBrewInstallation(isVerbose) {
		return "Homebrew"
	}
	if dir, err := currentInstallDir(); err == nil {
		home, _ := os.UserHomeDir()
		switch dir {
		case "/usr/local/bin", "/opt/homebrew/bin", "/usr/bin", "/opt/local/bin",
			filepath.Join(home, ".local", "bin"):
			return "Install Script (curl)"
		}
	}
	return "Binary / Go Install"
}

// currentInstallDir returns the directory containing the running executable,
// resolving symlinks when possible.
func currentInstallDir() (string, error) {
	executablePath, err := osExecutable()
	if err != nil {
		return "", err
	}
	if resolved, err := evalSymlinks(executablePath); err == nil {
		executablePath = resolved
	}
	return filepath.Dir(executablePath), nil
}

// printVersionInfo displays the static build and runtime information.
func printVersionInfo(isVerbose bool) {
	installSource := getInstallSource(isVerbose)

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary).
		MarginBottom(1)

	if !ui.DisableProgress {
		fmt.Fprintln(log.Out, headerStyle.Render("📌 eng CLI Version Details"))
	} else {
		log.Info("eng version: %s", Version)
		log.Message("  Git Commit: %s", Commit)
		log.Message("  Build Date: %s", Date)
		log.Message("  Go Version: %s", runtime.Version())
		log.Message("  OS/Arch:    %s/%s", runtime.GOOS, runtime.GOARCH)
		log.Message("  Install Source: %s", installSource)
		log.Message("") // Separator line
		return
	}

	var cardLines []string
	cardLines = append(
		cardLines,
		fmt.Sprintf("  %-16s %s", theme.BoldText.Render("Version:"), theme.PrimaryText.Bold(true).Render(Version)),
	)
	cardLines = append(
		cardLines,
		fmt.Sprintf("  %-16s %s", theme.BoldText.Render("Git Commit:"), theme.MutedText.Render(Commit)),
	)
	cardLines = append(
		cardLines,
		fmt.Sprintf("  %-16s %s", theme.BoldText.Render("Build Date:"), theme.MutedText.Render(Date)),
	)
	cardLines = append(
		cardLines,
		fmt.Sprintf("  %-16s %s", theme.BoldText.Render("Go Version:"), theme.BaseText.Render(runtime.Version())),
	)
	cardLines = append(
		cardLines,
		fmt.Sprintf(
			"  %-16s %s",
			theme.BoldText.Render("OS/Arch:"),
			theme.BaseText.Render(fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)),
		),
	)
	cardLines = append(
		cardLines,
		fmt.Sprintf("  %-16s %s", theme.BoldText.Render("Install Source:"), theme.BaseText.Render(installSource)),
	)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Primary).
		Padding(0, 1).
		MarginBottom(1)

	fmt.Fprintln(log.Out, boxStyle.Render(strings.Join(cardLines, "\n")))
}

// handleDevVersion logs information when running a development version.
func handleDevVersion(latestRelease *githubReleaseInfo) {
	log.Info("Currently running development version.")
	log.Message(
		"  Latest official release is %s: %s",
		theme.PrimaryText.Render(latestRelease.TagName),
		theme.MutedText.Render(latestRelease.HTMLURL),
	)
	if updateFlag {
		log.Warn("--update flag ignored when running a dev version.")
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
func compareAndHandleUpdate(
	currentSemVer, latestSemVer *semver.Version,
	latestRelease *githubReleaseInfo,
	isVerbose bool,
) {
	brewDetected := isBrewInstallation(isVerbose)

	if latestSemVer.GreaterThan(currentSemVer) {
		log.Success("A newer version is available: %s", latestRelease.TagName)
		log.Message("  Release notes: %s", theme.MutedText.Render(latestRelease.HTMLURL))
		log.Message("")

		if updateFlag {
			if brewDetected {
				log.Info("Attempting update via Homebrew...")
				err := runBrewUpgrade(isVerbose)
				if err != nil {
					log.Error("Brew update failed: %v", err)
					log.Info("  Please try manually: %s upgrade %s", brewCmd, brewPkgName)
					log.Info("  Or get it from GitHub: %s", latestRelease.HTMLURL)
				}
				// Exit after attempting update, regardless of success/failure
				// to avoid printing redundant "Get it here" messages.
				return
			}
			log.Info("Attempting update via install script...")
			if err := runScriptUpgrade(isVerbose); err != nil {
				log.Error("Install-script update failed: %v", err)
				log.Info("  Try manually: curl -sSfL %s | sh", installScriptURL)
				log.Info("  Or get it from GitHub: %s", latestRelease.HTMLURL)
			}
			return
		}
		// Just inform the user how to update
		if brewDetected {
			log.Info("Run `eng version --update` or `eng version -u` to attempt an automatic update via Homebrew.")
		} else {
			log.Info(
				"Run `eng version --update` or `eng version -u` to attempt an automatic update via the install script.",
			)
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
	executablePath, err := osExecutable()
	if err != nil {
		log.Verbose(isVerbose, "Could not get executable path: %v", err)
		return false
	}

	// Resolve symlinks, as brew often uses them (e.g., /usr/local/bin/eng -> ../Caskroom/eng/0.1.0/eng or ../Cellar/eng/0.1.0/bin/eng)
	resolvedPath, err := evalSymlinks(executablePath)
	if err != nil {
		// If symlink resolution fails, use the original path
		resolvedPath = executablePath
		log.Verbose(isVerbose, "Could not resolve symlink for executable path: %v", err)
	}

	// Common Homebrew installation prefixes:
	// - macOS Intel: /usr/local/Cellar (Formula), /usr/local/Caskroom (Cask)
	// - macOS Apple Silicon: /opt/homebrew/Cellar (Formula), /opt/homebrew/Caskroom (Cask)
	// - Linux (Linuxbrew): /home/linuxbrew/.linuxbrew/Cellar (Formula), /home/linuxbrew/.linuxbrew/Caskroom (Cask)
	brewPrefixes := []string{
		"/usr/local/Cellar",
		"/usr/local/Caskroom",
		"/opt/homebrew/Cellar",
		"/opt/homebrew/Caskroom",
		"/home/linuxbrew/.linuxbrew/Cellar",
		"/home/linuxbrew/.linuxbrew/Caskroom",
	}

	for _, prefix := range brewPrefixes {
		if strings.HasPrefix(resolvedPath, prefix) {
			log.Verbose(isVerbose, "Detected Homebrew installation path: %s", resolvedPath)
			// Check if 'brew' command actually exists for higher confidence
			_, err := lookPath(brewCmd)
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
// It first runs 'brew update' to refresh package information (including taps),
// then runs 'brew upgrade eng'.
// When verbose is false, outputs are hidden behind progress spinners.
// When verbose is true, outputs are streamed directly to the terminal.
func runBrewUpgrade(isVerbose bool) error {
	if isVerbose {
		// Verbose mode: stream raw output directly
		log.Info("Running '%s update'...", brewCmd)
		updateCmd := execCommand(brewCmd, "update")
		updateCmd.Stdout = log.Writer()
		updateCmd.Stderr = log.ErrorWriter()
		log.Verbose(isVerbose, "Executing command: %s", updateCmd.String())
		err := updateCmd.Run()
		if err != nil {
			log.Warn("'%s update' command finished with error (proceeding with upgrade attempt): %v", brewCmd, err)
		}

		log.Info("Running '%s upgrade %s'...", brewCmd, brewPkgName)
		upgradeCmd := execCommand(brewCmd, "upgrade", brewPkgName)
		upgradeCmd.Stdout = log.Writer()
		upgradeCmd.Stderr = log.ErrorWriter()
		log.Verbose(isVerbose, "Executing command: %s", upgradeCmd.String())
		err = upgradeCmd.Run()
		if err != nil {
			return fmt.Errorf("'%s upgrade %s' command failed: %w", brewCmd, brewPkgName, err)
		}
		log.Success("Homebrew upgrade successful! %s upgraded to latest version.", brewPkgName)
		return nil
	}

	// Non-verbose mode: hide raw brew outputs behind spinners
	var updateSpinner *ui.Spinner
	if !ui.DisableProgress {
		updateSpinner = ui.NewSpinner("Updating Homebrew package index...")
		updateSpinner.Start()
	}
	updateCmd := execCommand(brewCmd, "update")
	var updateStderr bytes.Buffer
	updateCmd.Stderr = &updateStderr
	err := updateCmd.Run()
	if updateSpinner != nil {
		updateSpinner.Stop()
	}
	if err != nil {
		log.Verbose(isVerbose, "brew update warning: %v (%s)", err, strings.TrimSpace(updateStderr.String()))
	}

	var upgradeSpinner *ui.Spinner
	if !ui.DisableProgress {
		upgradeSpinner = ui.NewSpinner(fmt.Sprintf("Upgrading %s package via Homebrew...", brewPkgName))
		upgradeSpinner.Start()
	}
	upgradeCmd := execCommand(brewCmd, "upgrade", brewPkgName)
	var upgradeStderr bytes.Buffer
	upgradeCmd.Stderr = &upgradeStderr
	err = upgradeCmd.Run()
	if upgradeSpinner != nil {
		upgradeSpinner.Stop()
	}
	if err != nil {
		stderrOutput := strings.TrimSpace(upgradeStderr.String())
		if stderrOutput != "" {
			log.Error("Brew upgrade error output:\n%s", stderrOutput)
		}
		return fmt.Errorf("'%s upgrade %s' command failed: %w", brewCmd, brewPkgName, err)
	}

	log.Success("Homebrew upgrade successful! %s upgraded to latest version.", brewPkgName)
	return nil
}

// runScriptUpgrade self-updates a curl/binary installation by downloading
// install.sh and re-executing it into the current install directory.
// The install script handles checksum verification, sudo, and PATH checks.
func runScriptUpgrade(isVerbose bool) error {
	destDir, err := currentInstallDir()
	if err != nil {
		return fmt.Errorf("could not determine install directory: %w", err)
	}
	if destDir == "" {
		return fmt.Errorf("empty install directory")
	}

	if isVerbose {
		log.Info("Downloading install script from %s...", installScriptURL)
	}

	var sp *ui.Spinner
	if !isVerbose && !ui.DisableProgress {
		sp = ui.NewSpinner("Downloading install script...")
		sp.Start()
	}

	tmpFile, err := downloadInstallScript(installScriptURL)
	if sp != nil {
		sp.Stop()
	}
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(tmpFile)
	}()

	args := []string{tmpFile, "--to", destDir}
	log.Verbose(isVerbose, "Executing: sh %s --to %s", tmpFile, destDir)

	var upgradeSpinner *ui.Spinner
	if !isVerbose && !ui.DisableProgress {
		upgradeSpinner = ui.NewSpinner(fmt.Sprintf("Upgrading eng in %s...", destDir))
		upgradeSpinner.Start()
	}
	upgradeCmd := execCommand("sh", args...)
	if isVerbose {
		upgradeCmd.Stdout = log.Writer()
		upgradeCmd.Stderr = log.ErrorWriter()
	} else {
		var stderrBuf bytes.Buffer
		upgradeCmd.Stderr = &stderrBuf
		err = upgradeCmd.Run()
		if upgradeSpinner != nil {
			upgradeSpinner.Stop()
		}
		if err != nil {
			if out := strings.TrimSpace(stderrBuf.String()); out != "" {
				log.Error("Install script error output:\n%s", out)
			}
			return fmt.Errorf("install script failed: %w", err)
		}
		log.Success("eng upgraded successfully in %s.", destDir)
		return nil
	}

	err = upgradeCmd.Run()
	if err != nil {
		return fmt.Errorf("install script failed: %w", err)
	}
	log.Success("eng upgraded successfully in %s.", destDir)
	return nil
}

// downloadInstallScript fetches the install script URL to a temp file.
func downloadInstallScript(url string) (string, error) {
	client := &http.Client{Timeout: scriptTimeout}
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download install script: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download install script: unexpected status %d", resp.StatusCode)
	}
	tmp, err := os.CreateTemp("", "eng-install-*.sh")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := io.Copy(tmp, resp.Body); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return "", fmt.Errorf("failed to write install script: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return "", fmt.Errorf("failed to close install script: %w", err)
	}
	if err := os.Chmod(tmpName, 0o755); err != nil {
		_ = os.Remove(tmpName)
		return "", fmt.Errorf("failed to chmod install script: %w", err)
	}
	return tmpName, nil
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
			return nil, fmt.Errorf(
				"github API request forbidden (status %d). Check rate limits or token permissions",
				resp.StatusCode,
			)
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
