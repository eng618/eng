package system

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/config"
)

func TestListProxyConfigurations(t *testing.T) {
	// Setup dummy config
	viper.Reset()
	viper.SetConfigType("json")
	proxies := []config.ProxyConfig{
		{Title: "Test Proxy", Value: "http://test:8080", Enabled: true},
	}
	viper.Set("proxies", proxies)

	// Create a test command with the required flags
	testCmd := &cobra.Command{}
	testCmd.Flags().Bool("compact", false, "")
	testCmd.Flags().Bool("env", false, "")
	testCmd.Flags().Bool("lowercase-env", false, "")

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	listProxyConfigurations(testCmd)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "Test Proxy") {
		t.Error("Expected output to contain 'Test Proxy'")
	}
	// The format is now: "1. ★ Test Proxy (http://test:8080) [ACTIVE]"
	if !strings.Contains(output, "1.") || !strings.Contains(output, "ACTIVE") {
		t.Error("Expected output to show proxy 1 as active with [ACTIVE]")
	}
}

func TestExportCmd_Enabled(t *testing.T) {
	viper.Reset()
	viper.SetConfigType("json")
	proxies := []config.ProxyConfig{
		{Title: "Test Proxy", Value: "http://test:8080", Enabled: true},
	}
	viper.Set("proxies", proxies)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run export subcommand logic
	exportCmd.Run(exportCmd, []string{})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "export HTTP_PROXY='http://test:8080'") {
		t.Error("Expected export command for HTTP_PROXY")
	}
}

func TestExportCmd_Disabled(t *testing.T) {
	viper.Reset()
	viper.SetConfigType("json")
	proxies := []config.ProxyConfig{
		{Title: "Test Proxy", Value: "http://test:8080", Enabled: false},
	}
	viper.Set("proxies", proxies)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	exportCmd.Run(exportCmd, []string{})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "unset HTTP_PROXY") {
		t.Error("Expected unset command for HTTP_PROXY when no proxy is enabled")
	}
}

func TestResolveProxyIndex(t *testing.T) {
	proxies := []config.ProxyConfig{
		{Title: "Corp VPN", Value: "http://corp:8080", Enabled: true},
		{Title: "Home Relay", Value: "http://home:1080", Enabled: false},
	}

	// Test 1-based index string
	if idx := resolveProxyIndex("1", -1, "", proxies); idx != 0 {
		t.Errorf("Expected 0 for 1-based index '1', got %d", idx)
	}

	if idx := resolveProxyIndex("2", -1, "", proxies); idx != 1 {
		t.Errorf("Expected 1 for 1-based index '2', got %d", idx)
	}

	// Test title resolution
	if idx := resolveProxyIndex("home relay", -1, "", proxies); idx != 1 {
		t.Errorf("Expected 1 for title match 'home relay', got %d", idx)
	}

	// Test flag fallback
	if idx := resolveProxyIndex("", 0, "", proxies); idx != 0 {
		t.Errorf("Expected 0 for flag fallback index 0, got %d", idx)
	}

	// Test non-match fallback
	if idx := resolveProxyIndex("unknown", -1, "", proxies); idx != -1 {
		t.Errorf("Expected -1 for unknown proxy, got %d", idx)
	}
}

