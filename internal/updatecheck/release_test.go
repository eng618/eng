package updatecheck

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func clearGitHubTokens(t *testing.T) {
	t.Helper()
	t.Setenv("GH_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "")
}

func TestFetchLatestFromAPI(t *testing.T) {
	clearGitHubTokens(t)

	// tag_name sits past the old 512-byte error-body cap. A real GitHub
	// release payload is larger than that because it includes the changelog.
	padding := strings.Repeat("x", 800)
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/vnd.github+json", r.Header.Get("Accept"))
		assert.Contains(t, r.Header.Get("User-Agent"), "eng/")
		assert.Empty(t, r.Header.Get("Authorization"))
		_, _ = w.Write([]byte(`{"filler":"` + padding +
			`","tag_name":"v1.2.3","html_url":"https://github.com/eng618/eng/releases/tag/v1.2.3"}`))
	}))
	t.Cleanup(api.Close)

	rel, err := fetchLatestRelease(context.Background(), api.URL, "http://127.0.0.1:1/releases/latest")
	require.NoError(t, err)
	assert.Equal(t, "v1.2.3", rel.TagName)
	assert.Equal(t, "https://github.com/eng618/eng/releases/tag/v1.2.3", rel.HTMLURL)
}

func TestFetchLatestSendsGHTokenOnlyToAPI(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "should-not-win")
	t.Setenv("GH_TOKEN", "test-token")

	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"API rate limit exceeded"}`))
	}))
	t.Cleanup(api.Close)

	page := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.Header.Get("Authorization"))
		w.Header().Set("Location", "https://github.com/eng618/eng/releases/tag/v9.9.9")
		w.WriteHeader(http.StatusFound)
	}))
	t.Cleanup(page.Close)

	rel, err := fetchLatestRelease(context.Background(), api.URL, page.URL+"/releases/latest")
	require.NoError(t, err)
	assert.Equal(t, "v9.9.9", rel.TagName)
	assert.Equal(t, "https://github.com/eng618/eng/releases/tag/v9.9.9", rel.HTMLURL)
}

func TestFetchLatestFallsBackWhenAPIForbidden(t *testing.T) {
	clearGitHubTokens(t)

	apiCalls := 0
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		apiCalls++
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"API rate limit exceeded for ` + strings.Repeat("x", 400) + `"}`))
	}))
	t.Cleanup(api.Close)

	pageCalls := 0
	page := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		pageCalls++
		w.Header().Set("Location", "/eng618/eng/releases/tag/v1.66.1")
		w.WriteHeader(http.StatusFound)
	}))
	t.Cleanup(page.Close)

	rel, err := fetchLatestRelease(context.Background(), api.URL, page.URL+"/releases/latest")
	require.NoError(t, err)
	assert.Equal(t, "v1.66.1", rel.TagName)
	assert.Equal(t, page.URL+"/eng618/eng/releases/tag/v1.66.1", rel.HTMLURL)
	assert.Equal(t, 1, apiCalls)
	assert.Equal(t, 1, pageCalls)
}

func TestFetchLatestReportsBothFailures(t *testing.T) {
	clearGitHubTokens(t)

	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"API rate limit exceeded"}`))
	}))
	t.Cleanup(api.Close)

	page := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(page.Close)

	_, err := fetchLatestRelease(context.Background(), api.URL, page.URL+"/releases/latest")
	require.Error(t, err)
	assert.ErrorContains(t, err, "github API rate limit exceeded (status 403)")
	assert.ErrorContains(t, err, "releases page:")
	assert.ErrorContains(t, err, "status 500")
	assert.NotContains(t, err.Error(), "token")
}

func TestFetchLatestNoReleaseWhenBothMissing(t *testing.T) {
	clearGitHubTokens(t)

	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(api.Close)
	page := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(page.Close)

	_, err := fetchLatestRelease(context.Background(), api.URL, page.URL)
	require.ErrorIs(t, err, ErrNoRelease)
}

func TestFetchLatestNotFoundAPIStillUsesRedirect(t *testing.T) {
	clearGitHubTokens(t)

	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(api.Close)
	page := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/releases/tag/v3.0.0", http.StatusFound)
	}))
	t.Cleanup(page.Close)

	rel, err := fetchLatestRelease(context.Background(), api.URL, page.URL+"/releases/latest")
	require.NoError(t, err)
	assert.Equal(t, "v3.0.0", rel.TagName)
}

func TestFetchLatestAPINotFoundAndPageFailureIsAnError(t *testing.T) {
	clearGitHubTokens(t)

	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(api.Close)
	page := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	t.Cleanup(page.Close)

	_, err := fetchLatestRelease(context.Background(), api.URL, page.URL)
	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrNoRelease)
	assert.ErrorContains(t, err, "releases page:")
}

func TestTagFromReleaseURL(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"tag", "https://github.com/eng618/eng/releases/tag/v1.2.3", "v1.2.3"},
		{"trailing slash", "https://github.com/eng618/eng/releases/tag/v1.2.3/", "v1.2.3"},
		{"query ignored", "https://github.com/eng618/eng/releases/tag/v1.2.3?utm=1", "v1.2.3"},
		{"missing marker", "https://github.com/eng618/eng/releases/latest", ""},
		{"extra segment", "https://github.com/eng618/eng/releases/tag/v1.2.3/extra", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.raw)
			require.NoError(t, err)
			assert.Equal(t, tt.want, tagFromReleaseURL(u))
		})
	}
	assert.Empty(t, tagFromReleaseURL(nil))
}

func TestAPIStatusErrorOmitsTokenAdvice(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusForbidden,
		Header:     http.Header{"X-RateLimit-Remaining": []string{"0"}},
	}
	err := apiStatusError(resp, []byte(`{"message":"API rate limit exceeded"}`))
	require.Error(t, err)
	assert.ErrorContains(t, err, "rate limit exceeded")
	assert.NotContains(t, err.Error(), "token")

	forbidden := &http.Response{StatusCode: http.StatusForbidden, Header: http.Header{}}
	err = apiStatusError(forbidden, []byte(`{"message":"Resource not accessible"}`))
	assert.ErrorContains(t, err, "forbidden (status 403): Resource not accessible")
	assert.False(t, errors.Is(err, ErrNoRelease))
}
