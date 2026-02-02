package system

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/utils/config"
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
