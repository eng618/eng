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
		assert.Equal(t, "application/vnd.github.v3+json", r.Header.Get("Accept"))
		_ = json.NewEncoder(w).Encode(map[string]string{
			"tag_name": "v9999.0.1",
			"html_url": "https://example.com/releases",
		})
	}))
	defer server.Close()

	oldURL := githubAPIURL
	githubAPIURL = server.URL + "/repos/%s/%s/releases/latest"
	defer func() { githubAPIURL = oldURL }()

	path, err := CacheFile()
	require.NoError(t, err)
	require.NoFileExists(t, path)

	RefreshAsync()

	require.Eventually(t, func() bool {
		data, err := os.ReadFile(path)
		if err != nil {
			return false
		}
		var entry cacheEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			return false
		}
		return entry.LatestTag == "v9999.0.1"
	}, 5*time.Second, 50*time.Millisecond)
}

func TestRefreshAsyncSkipsFreshCache(t *testing.T) {
	withTestCache(t)

	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		_ = json.NewEncoder(w).Encode(map[string]string{"tag_name": "v9999.0.2"})
	}))
	defer server.Close()

	oldURL := githubAPIURL
	githubAPIURL = server.URL + "/repos/%s/%s/releases/latest"
	defer func() { githubAPIURL = oldURL }()

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

	oldURL := githubAPIURL
	githubAPIURL = "http://127.0.0.1:1/repos/%s/%s/releases/latest"
	defer func() { githubAPIURL = oldURL }()

	assert.NotPanics(t, func() {
		RefreshAsync()
		time.Sleep(300 * time.Millisecond)
	})
}
