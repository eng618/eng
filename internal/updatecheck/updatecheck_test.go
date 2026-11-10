package updatecheck

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	appversion "github.com/eng618/eng/internal/version"
)

// withTestCache points the cache at a temp dir and enables the check paths
// that ui.DisableProgress otherwise skips (mirrors runlog's test convention).
// It also pins a release build version, since tests run as "dev" (silent).
func withTestCache(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv(cacheFileEnvVar, filepath.Join(dir, cacheFileName))
	t.Setenv(disableEnvVar, "")
	old := ui.DisableProgress
	ui.DisableProgress = false
	t.Cleanup(func() { ui.DisableProgress = old })
	oldVersion := appversion.Version
	appversion.Version = "v0.1.0"
	t.Cleanup(func() { appversion.Version = oldVersion })
	// A developer token must not leak into httptest assertions or requests.
	t.Setenv("GH_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "")
	return dir
}

func writeEntry(t *testing.T, path, tag string, checkedAt time.Time) {
	t.Helper()
	data, err := json.Marshal(cacheEntry{LatestTag: tag, CheckedAt: checkedAt.Unix()})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0o644))
}

func TestNewerAvailable(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
		want    bool
	}{
		{"newer patch", "v0.1.0", "v0.1.1", true},
		{"newer minor", "v0.1.0", "v0.2.0", true},
		{"newer major", "v1.0.0", "v2.0.0", true},
		{"equal", "v0.1.0", "v0.1.0", false},
		{"older cached", "v0.2.0", "v0.1.0", false},
		{"unparseable current", "dev", "v0.1.0", false},
		{"unparseable latest", "v0.1.0", "not-a-version", false},
		{"empty latest", "v0.1.0", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, newerAvailable(tt.current, tt.latest))
		})
	}
}

func TestShouldSkip(t *testing.T) {
	assert.True(t, ShouldSkip("version"))
	assert.True(t, ShouldSkip("completion"))
	assert.True(t, ShouldSkip("__complete"))
	assert.False(t, ShouldSkip("git"))
	assert.False(t, ShouldSkip(""))
}

func TestDisabled(t *testing.T) {
	t.Setenv(disableEnvVar, "")
	assert.False(t, Disabled())
	t.Setenv(disableEnvVar, "1")
	assert.True(t, Disabled())
}

func TestNotifyIfAvailable(t *testing.T) {
	withTestCache(t)
	path, err := CacheFile()
	require.NoError(t, err)

	t.Run("newer cached tag notifies on stderr", func(t *testing.T) {
		writeEntry(t, path, "v9999.0.0", time.Now())
		var errBuf bytes.Buffer
		log.SetWriters(&bytes.Buffer{}, &errBuf)
		defer log.ResetWriters()
		NotifyIfAvailable("git")
		out := errBuf.String()
		assert.Contains(t, out, "Update available")
		assert.Contains(t, out, "v0.1.0 → v9999.0.0")
		assert.Contains(t, out, "eng version -u")
	})

	t.Run("older cached tag stays silent", func(t *testing.T) {
		writeEntry(t, path, "v0.0.1", time.Now())
		var errBuf bytes.Buffer
		log.SetWriters(&bytes.Buffer{}, &errBuf)
		defer log.ResetWriters()
		NotifyIfAvailable("git")
		assert.Empty(t, errBuf.String())
	})

	t.Run("dev build stays silent", func(t *testing.T) {
		appversion.Version = "dev"
		defer func() { appversion.Version = "v0.1.0" }()
		writeEntry(t, path, "v9999.0.0", time.Now())
		var errBuf bytes.Buffer
		log.SetWriters(&bytes.Buffer{}, &errBuf)
		defer log.ResetWriters()
		NotifyIfAvailable("git")
		assert.Empty(t, errBuf.String())
	})

	t.Run("skipped command stays silent", func(t *testing.T) {
		writeEntry(t, path, "v9999.0.0", time.Now())
		var errBuf bytes.Buffer
		log.SetWriters(&bytes.Buffer{}, &errBuf)
		defer log.ResetWriters()
		NotifyIfAvailable("version")
		assert.Empty(t, errBuf.String())
	})

	t.Run("opt-out stays silent", func(t *testing.T) {
		writeEntry(t, path, "v9999.0.0", time.Now())
		t.Setenv(disableEnvVar, "1")
		var errBuf bytes.Buffer
		log.SetWriters(&bytes.Buffer{}, &errBuf)
		defer log.ResetWriters()
		NotifyIfAvailable("git")
		assert.Empty(t, errBuf.String())
	})

	t.Run("missing cache stays silent", func(t *testing.T) {
		require.NoError(t, os.Remove(path))
		var errBuf bytes.Buffer
		log.SetWriters(&bytes.Buffer{}, &errBuf)
		defer log.ResetWriters()
		NotifyIfAvailable("git")
		assert.Empty(t, errBuf.String())
	})
}

func TestRefreshAsyncFetchesAndCaches(t *testing.T) {
	withTestCache(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/vnd.github+json", r.Header.Get("Accept"))
		assert.Contains(t, r.Header.Get("User-Agent"), "eng/")
		assert.Empty(t, r.Header.Get("Authorization"))
		_ = json.NewEncoder(w).Encode(map[string]string{
			"tag_name": "v9999.0.1",
			"html_url": "https://example.com/releases",
		})
	}))
	defer server.Close()

	oldURL, oldPage := githubAPIURL, githubLatestPageURL
	githubAPIURL = server.URL + "/repos/eng618/eng/releases/latest"
	githubLatestPageURL = "http://127.0.0.1:1/releases/latest"
	t.Cleanup(func() {
		githubAPIURL = oldURL
		githubLatestPageURL = oldPage
	})

	done := make(chan struct{})
	fetchDone = func() { close(done) }
	t.Cleanup(func() { fetchDone = nil })

	path, err := CacheFile()
	require.NoError(t, err)
	require.NoFileExists(t, path)

	RefreshAsync()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("background update check did not finish")
	}

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var entry cacheEntry
	require.NoError(t, json.Unmarshal(data, &entry))
	assert.Equal(t, "v9999.0.1", entry.LatestTag)
}

func TestRefreshAsyncSkipsFreshCache(t *testing.T) {
	withTestCache(t)

	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		_ = json.NewEncoder(w).Encode(map[string]string{"tag_name": "v9999.0.2"})
	}))
	defer server.Close()

	oldURL, oldPage := githubAPIURL, githubLatestPageURL
	githubAPIURL = server.URL + "/repos/eng618/eng/releases/latest"
	githubLatestPageURL = "http://127.0.0.1:1/releases/latest"
	defer func() {
		githubAPIURL = oldURL
		githubLatestPageURL = oldPage
	}()

	path, err := CacheFile()
	require.NoError(t, err)
	writeEntry(t, path, "v1.0.0", time.Now())

	RefreshAsync()
	time.Sleep(300 * time.Millisecond)
	assert.Equal(t, 0, calls, "fresh cache must not trigger a network fetch")

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var entry cacheEntry
	require.NoError(t, json.Unmarshal(data, &entry))
	assert.Equal(t, "v1.0.0", entry.LatestTag)
}

func TestRefreshAsyncSwallowsErrors(t *testing.T) {
	withTestCache(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	oldURL, oldPage := githubAPIURL, githubLatestPageURL
	githubAPIURL = server.URL
	githubLatestPageURL = server.URL + "/latest"
	t.Cleanup(func() {
		githubAPIURL = oldURL
		githubLatestPageURL = oldPage
	})

	done := make(chan struct{})
	fetchDone = func() { close(done) }
	t.Cleanup(func() { fetchDone = nil })

	assert.NotPanics(t, func() { RefreshAsync() })
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("background update check did not finish")
	}

	path, err := CacheFile()
	require.NoError(t, err)
	_, statErr := os.Stat(path)
	assert.True(t, os.IsNotExist(statErr), "a failed check must not write the cache")
}
