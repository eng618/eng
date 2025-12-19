package system

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/eng618/eng/utils/config"
)

func TestListProxyConfigurations(t *testing.T) {
	// Setup dummy config
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

	listProxyConfigurations()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "Test Proxy") {
		t.Error("Expected output to contain 'Test Proxy'")
	}
	if !strings.Contains(output, "[*] 1. Test Proxy") {
		t.Error("Expected output to show proxy 1 as enabled with [*]")
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
