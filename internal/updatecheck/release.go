package updatecheck

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	appversion "github.com/eng618/eng/internal/version"
)

const (
	releaseTagMarker = "/releases/tag/"
	// maxReleaseBody caps a successful release payload. The body includes the
	// changelog, so it is much larger than an error response.
	maxReleaseBody = 1 << 20
	// maxAPIErrorBody caps the snippet kept from a non-success API response.
	maxAPIErrorBody = 512
)

// ErrNoRelease means the repository has no published GitHub release.
var ErrNoRelease = errors.New("no github release published")

// Release is the newest published eng release.
type Release struct {
	TagName string
	HTMLURL string
}

// githubLatestPageURL is the public releases page. GitHub answers with a
// redirect to /releases/tag/<version>, which does not use the REST API quota.
var githubLatestPageURL = "https://github.com/eng618/eng/releases/latest"

// FetchLatest returns the newest eng release. It tries the GitHub REST API,
// then the public releases page when the API is forbidden, rate-limited, or
// unreachable.
func FetchLatest(ctx context.Context) (Release, error) {
	return fetchLatestRelease(ctx, githubAPIURL, githubLatestPageURL)
}

func fetchLatestRelease(ctx context.Context, apiURL, latestPageURL string) (Release, error) {
	rel, apiErr := fetchLatestFromAPI(ctx, apiURL)
	if apiErr == nil && rel.TagName != "" {
		return rel, nil
	}
	if apiErr == nil {
		apiErr = errors.New("empty tag name")
	}

	fallback, fbErr := fetchLatestFromRedirect(ctx, latestPageURL)
	if fbErr == nil && fallback.TagName != "" {
		return fallback, nil
	}
	// A 404 means "no published release" only when the public page agrees.
	// A later transport or status failure is still a failed check. That error
	// must not wrap ErrNoRelease: callers treat ErrNoRelease as "nothing to compare".
	if errors.Is(apiErr, ErrNoRelease) && (fbErr == nil || errors.Is(fbErr, ErrNoRelease)) {
		return Release{}, ErrNoRelease
	}
	if fbErr == nil {
		fbErr = ErrNoRelease
	}
	if errors.Is(apiErr, ErrNoRelease) {
		return Release{}, fmt.Errorf("github API reported no release (releases page: %v)", fbErr)
	}
	return Release{}, fmt.Errorf("%w (releases page: %v)", apiErr, fbErr)
}

func fetchLatestFromAPI(ctx context.Context, apiURL string) (Release, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return Release{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", requestUserAgent())
	if token, ok := githubToken(); ok {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Release{}, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNotFound {
		return Release{}, ErrNoRelease
	}
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(io.LimitReader(resp.Body, maxAPIErrorBody))
		if err != nil {
			return Release{}, err
		}
		return Release{}, apiStatusError(resp, body)
	}

	var payload struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxReleaseBody)).Decode(&payload); err != nil {
		return Release{}, err
	}
	if payload.TagName == "" {
		return Release{}, errors.New("empty tag name")
	}
	if payload.HTMLURL == "" {
		payload.HTMLURL = releasePageURL(payload.TagName)
	}
	return Release{TagName: payload.TagName, HTMLURL: payload.HTMLURL}, nil
}

func fetchLatestFromRedirect(ctx context.Context, latestPageURL string) (Release, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	// Stop at the redirect. The Location header is the tag; the HTML page is not needed.
	client := &http.Client{
		Timeout: requestTimeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, latestPageURL, nil)
	if err != nil {
		return Release{}, err
	}
	req.Header.Set("User-Agent", requestUserAgent())

	resp, err := client.Do(req)
	if err != nil {
		return Release{}, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode == http.StatusNotFound {
		return Release{}, ErrNoRelease
	}
	if resp.StatusCode < 300 || resp.StatusCode >= 400 {
		return Release{}, fmt.Errorf("releases page returned status %d", resp.StatusCode)
	}

	loc := strings.TrimSpace(resp.Header.Get("Location"))
	if loc == "" {
		return Release{}, errors.New("releases page redirect had no location")
	}
	base, err := url.Parse(latestPageURL)
	if err != nil {
		return Release{}, err
	}
	ref, err := url.Parse(loc)
	if err != nil {
		return Release{}, err
	}
	abs := base.ResolveReference(ref)
	tag := tagFromReleaseURL(abs)
	if tag == "" {
		return Release{}, fmt.Errorf("releases page redirected to %s", abs.String())
	}
	return Release{TagName: tag, HTMLURL: abs.String()}, nil
}

func apiStatusError(resp *http.Response, body []byte) error {
	message := apiErrorMessage(body)
	switch resp.StatusCode {
	case http.StatusForbidden, http.StatusTooManyRequests:
		if resp.Header.Get("X-RateLimit-Remaining") == "0" {
			return fmt.Errorf("github API rate limit exceeded (status %d)", resp.StatusCode)
		}
		if message != "" {
			return fmt.Errorf("github API request forbidden (status %d): %s", resp.StatusCode, message)
		}
		return fmt.Errorf("github API request forbidden (status %d)", resp.StatusCode)
	default:
		if message != "" {
			return fmt.Errorf("unexpected status code %d from GitHub API: %s", resp.StatusCode, message)
		}
		return fmt.Errorf("unexpected status code %d from GitHub API", resp.StatusCode)
	}
}

func apiErrorMessage(body []byte) string {
	var parsed struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(body, &parsed) == nil && strings.TrimSpace(parsed.Message) != "" {
		return oneLine(parsed.Message)
	}
	return oneLine(string(body))
}

func oneLine(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > 160 {
		return s[:160]
	}
	return s
}

func githubToken() (string, bool) {
	for _, key := range []string{"GH_TOKEN", "GITHUB_TOKEN"} {
		if token := strings.TrimSpace(os.Getenv(key)); token != "" {
			return token, true
		}
	}
	return "", false
}

func requestUserAgent() string {
	return fmt.Sprintf("eng/%s (+https://github.com/eng618/eng)", appversion.Version)
}

func releasePageURL(tag string) string {
	return "https://github.com/eng618/eng/releases/tag/" + tag
}

func tagFromReleaseURL(u *url.URL) string {
	if u == nil {
		return ""
	}
	path := strings.TrimSuffix(u.Path, "/")
	i := strings.LastIndex(path, releaseTagMarker)
	if i < 0 {
		return ""
	}
	tag := path[i+len(releaseTagMarker):]
	if tag == "" || strings.Contains(tag, "/") {
		return ""
	}
	return tag
}
