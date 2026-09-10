package updatecheck

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	appversion "github.com/eng618/eng/internal/version"
)

const (
	githubRepoOwner = "eng618"
	githubRepoName  = "eng"

	// CheckInterval is how often the cached latest-release tag is refreshed.
	CheckInterval = 24 * time.Hour
	// requestTimeout bounds the background GitHub API call.
	requestTimeout = 3 * time.Second

	// disableEnvVar opts out of the automatic check when set to any non-empty value.
	disableEnvVar = "ENG_NO_UPDATE_CHECK"
	// cacheFileEnvVar overrides the cache file path (used by tests).
	cacheFileEnvVar = "ENG_UPDATE_CHECK_FILE"

	cacheFileName = "update-check.json"
)

// githubAPIURL is the latest-release endpoint (overridable in tests).
var githubAPIURL = "https://api.github.com/repos/%s/%s/releases/latest"

// Commands that never trigger the notice; version performs its own check and
// completion output must stay machine-clean.
var skippedCommands = map[string]bool{
	"version":    true,
	"completion": true,
}

// cacheEntry is the on-disk state: last known upstream tag + when fetched.
type cacheEntry struct {
	LatestTag string `json:"latest_tag"`
	CheckedAt int64  `json:"checked_at_unix"`
}

// CacheFile returns the cache path, creating the parent dir when possible.
// ENG_UPDATE_CHECK_FILE overrides the default (OS cache dir), keeping tests hermetic.
func CacheFile() (string, error) {
	if override := strings.TrimSpace(os.Getenv(cacheFileEnvVar)); override != "" {
		if dir := filepath.Dir(override); dir != "" {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return "", err
			}
		}
		return override, nil
	}
	base, err := os.UserCacheDir()
	if err != nil || base == "" {
		base = os.TempDir()
	}
	dir := filepath.Join(base, "eng")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, cacheFileName), nil
}

// Disabled reports whether the automatic check is turned off.
func Disabled() bool {
	return strings.TrimSpace(os.Getenv(disableEnvVar)) != ""
}

// ShouldSkip reports whether the notice/refresh should be skipped for the
// invoked command name (e.g. "version", "completion", "__complete").
func ShouldSkip(cmdName string) bool {
	if cmdName == "__complete" {
		return true
	}
	return skippedCommands[cmdName]
}

// readCache loads the cache entry; missing/corrupt files yield a zero entry.
func readCache(path string) cacheEntry {
	var entry cacheEntry
	data, err := os.ReadFile(path)
	if err != nil {
		return entry
	}
	if err := json.Unmarshal(data, &entry); err != nil {
		return cacheEntry{}
	}
	return entry
}

// writeCache stores the entry atomically-ish (temp file + rename).
func writeCache(path string, entry cacheEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "update-check-*.tmp")
	if err != nil {
		return
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return
	}
	_ = os.Rename(tmpName, path)
}

// fetchLatestTag queries the GitHub releases API for the latest tag.
func fetchLatestTag(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf(githubAPIURL, githubRepoOwner, githubRepoName), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	if release.TagName == "" {
		return "", fmt.Errorf("empty tag name")
	}
	return release.TagName, nil
}

// newerAvailable reports whether latestTag is a newer semver than current.
// Unparseable versions are treated as "no update" to stay silent.
func newerAvailable(current, latestTag string) bool {
	currentVer, err := semver.NewVersion(current)
	if err != nil {
		return false
	}
	latestVer, err := semver.NewVersion(latestTag)
	if err != nil {
		return false
	}
	return latestVer.GreaterThan(currentVer)
}

// RefreshAsync refreshes the cached latest-release tag in the background when
// the cache is missing or older than CheckInterval. It returns immediately and
// never reports errors; the next invocation serves the refreshed value.
func RefreshAsync() {
	if Disabled() || ui.DisableProgress {
		return
	}
	if appversion.Version == "dev" {
		return
	}
	path, err := CacheFile()
	if err != nil {
		return
	}
	entry := readCache(path)
	if time.Since(time.Unix(entry.CheckedAt, 0)) < CheckInterval && entry.LatestTag != "" {
		return
	}
	go func() {
		bg, cancel := context.WithTimeout(context.Background(), requestTimeout+time.Second)
		defer cancel()
		tag, err := fetchLatestTag(bg)
		if err != nil {
			return
		}
		writeCache(path, cacheEntry{LatestTag: tag, CheckedAt: time.Now().Unix()})
	}()
}

// NotifyIfAvailable prints a short npm-style update notice to stderr when the
// cached latest tag is newer than the running version. It performs no network
// I/O. Pass the invoked (sub)command name so self-checking and
// machine-consumed commands stay quiet.
func NotifyIfAvailable(cmdName string) {
	if Disabled() || ui.DisableProgress {
		return
	}
	if ShouldSkip(cmdName) {
		return
	}
	if appversion.Version == "dev" {
		return
	}
	path, err := CacheFile()
	if err != nil {
		return
	}
	entry := readCache(path)
	if entry.LatestTag == "" || !newerAvailable(appversion.Version, entry.LatestTag) {
		return
	}
	log.Warn("Update available: %s → %s", appversion.Version, entry.LatestTag)
	log.Warn("Run `eng version -u` to update.")
}
