package config

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui/theme"
)

// ProxyConfig represents a single proxy configuration.
func AddOrUpdateProxy() ([]ProxyConfig, int) {
	proxies, _ := GetProxyConfigs()

	title, value, noProxy, err := PromptProxyValues("", "", "")
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			log.Info("Operation canceled.")
			return proxies, -1
		}
		cobra.CheckErr(err)
	}

	// Normalize proxy value and no_proxy list
	value = NormalizeProxyURLString(value)
	noProxy = normalizeNoProxyList(noProxy)

	// Check if we're updating an existing proxy
	index := -1
	for i, proxy := range proxies {
		if strings.EqualFold(proxy.Title, title) {
			index = i
			break
		}
	}

	if index >= 0 {
		// Update existing proxy
		proxies[index].Value = value
		proxies[index].NoProxy = noProxy
		proxies[index].Title = title // preserve casing
	} else {
		// Add new proxy
		newProxy := ProxyConfig{
			Title:   title,
			Value:   value,
			NoProxy: noProxy,
			Enabled: false,
		}
		proxies = append(proxies, newProxy)
		index = len(proxies) - 1
	}

	// Save configurations
	if err := SaveProxyConfigs(proxies); err != nil {
		cobra.CheckErr(err)
	}

	log.Success("Proxy '%s' added/updated successfully", title)
	return proxies, index
}

// SelectProxy prompts the user to select a proxy from the list and returns the index.
func SelectProxy(proxies []ProxyConfig) (int, error) {
	if len(proxies) == 0 {
		return -1, errors.New("no proxy configurations found")
	}

	var options []string
	for _, proxy := range proxies {
		options = append(options, FormatProxyOption(proxy))
	}

	selectedStr, err := SelectPrompt("Select a proxy configuration:", options, "")
	if err != nil {
		return -1, err
	}

	var selectedIndex int
	for i, opt := range options {
		if opt == selectedStr {
			selectedIndex = i
			break
		}
	}

	return selectedIndex, nil
}

// FormatProxyOption renders a single proxy as a stylized radio option string.
// Example: "● Corp Proxy (http://proxy:8080)" with colored markers and dimmed value.
func FormatProxyOption(proxy ProxyConfig) string {
	// Marker and label with stronger contrast: ★ ACTIVE vs • inactive
	marker := lipgloss.NewStyle().Foreground(theme.MutedForeground).Render("•")
	label := lipgloss.NewStyle().Foreground(theme.MutedForeground).Render("[inactive]")
	title := proxy.Title

	if proxy.Enabled {
		marker = lipgloss.NewStyle().Foreground(theme.Secondary).Bold(true).Render("★")
		label = lipgloss.NewStyle().Foreground(theme.Secondary).Render("[ACTIVE]")
		title = theme.BoldText.Render(proxy.Title)
	}

	// Value in dim gray
	value := lipgloss.NewStyle().Foreground(theme.MutedForeground).Render(fmt.Sprintf("(%s)", proxy.Value))

	return fmt.Sprintf("%s %s %s %s", marker, title, value, label)
}

// Validation helpers.

// allowedSchemes lists the proxy URL schemes eng accepts.
var allowedSchemes = map[string]bool{
	"http":    true,
	"https":   true,
	"socks5":  true,
	"socks5h": true,
}

// --- Programmatic helpers ---

// FindProxyIndexByTitle returns the index of the proxy matching the title (case-insensitive), or -1.
func FindProxyIndexByTitle(proxies []ProxyConfig, title string) int {
	t := strings.TrimSpace(title)
	if t == "" {
		return -1
	}
	tLower := strings.ToLower(t)
	for i, p := range proxies {
		if strings.ToLower(p.Title) == tLower {
			return i
		}
	}
	return -1
}

// AddOrUpdateProxyWithValues adds or updates a proxy using provided values (non-interactive).
// Returns updated proxies, the affected index, or an error.
func AddOrUpdateProxyWithValues(title, value, noProxy string) ([]ProxyConfig, int, error) {
	proxies, _ := GetProxyConfigs()

	if err := validateTitle(title); err != nil {
		return proxies, -1, err
	}
	if err := ValidateProxyURLString(value); err != nil {
		return proxies, -1, err
	}
	noProxy = normalizeNoProxyList(noProxy)

	index := FindProxyIndexByTitle(proxies, title)
	if index >= 0 {
		// Update existing
		proxies[index].Value = value
		proxies[index].NoProxy = noProxy
	} else {
		// Add new
		newProxy := ProxyConfig{
			Title:   strings.TrimSpace(title),
			Value:   strings.TrimSpace(value),
			NoProxy: noProxy,
			Enabled: false,
		}
		proxies = append(proxies, newProxy)
		index = len(proxies) - 1
	}

	// Save configurations
	if err := SaveProxyConfigs(proxies); err != nil {
		return proxies, -1, err
	}
	return proxies, index, nil
}

// validateTitle ensures the title is non-empty after trimming.
func ValidateProxyURLString(value string) error {
	s := strings.TrimSpace(value)
	if s == "" {
		return errors.New("proxy address is required")
	}

	// Normalize scheme-less values to default http for validation consistency
	s = NormalizeProxyURLString(s)

	u, err := url.Parse(s)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}
	if !allowedSchemes[u.Scheme] {
		return fmt.Errorf("unsupported scheme '%s' (allowed: http, https, socks5, socks5h)", u.Scheme)
	}
	if u.Host == "" {
		return errors.New("missing host:port in proxy address")
	}
	// Require port
	_, _, err = net.SplitHostPort(u.Host)
	if err != nil {
		return errors.New("proxy address must include host:port")
	}
	return nil
}

// normalizeNoProxyList trims whitespace, removes empty entries, and de-duplicates.
func normalizeNoProxyList(list string) string {
	if strings.TrimSpace(list) == "" {
		return ""
	}
	parts := strings.Split(list, ",")
	seen := make(map[string]struct{})
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return strings.Join(out, ",")
}

// NormalizeProxyURLString adds a default http scheme when missing.
func NormalizeProxyURLString(value string) string {
	s := strings.TrimSpace(value)
	if s == "" {
		return s
	}
	if strings.Contains(s, "://") {
		return s
	}
	// If looks like host:port, prepend http://
	if strings.Contains(s, ":") {
		// Best effort: assume http if no scheme provided
		return "http://" + s
	}
	// No port provided; leave as-is (validator will catch missing port)
	return s
}

// RemoveProxy removes the proxy configuration at the specified index.
func RemoveProxy(index int) ([]ProxyConfig, error) {
	proxies, _ := GetProxyConfigs()
	if index < 0 || index >= len(proxies) {
		return proxies, errors.New("proxy index out of range")
	}

	// If deleting active proxy, unset env vars
	if proxies[index].Enabled {
		UnsetProxyEnvVars()
	}

	title := proxies[index].Title
	proxies = append(proxies[:index], proxies[index+1:]...)

	if err := SaveProxyConfigs(proxies); err != nil {
		return proxies, err
	}

	log.Success("Proxy '%s' removed successfully", title)
	return proxies, nil
}

// TestProxyConnection tests HTTP connectivity using the specified proxy URL.
func TestProxyConnection(proxyURLStr, targetURLStr string) (time.Duration, error) {
	if targetURLStr == "" {
		targetURLStr = "https://1.1.1.1"
	}

	normalizedProxy := NormalizeProxyURLString(proxyURLStr)
	parsedProxy, err := url.Parse(normalizedProxy)
	if err != nil {
		return 0, fmt.Errorf("invalid proxy URL '%s': %w", proxyURLStr, err)
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(parsedProxy),
		DialContext: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   7 * time.Second,
	}

	start := time.Now()
	resp, err := client.Get(targetURLStr)
	duration := time.Since(start)
	if err != nil {
		return duration, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return duration, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	return duration, nil
}
