package sysinfo

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eng618/eng/internal/log"
	infosys "github.com/eng618/eng/internal/sysinfo"
	"github.com/eng618/eng/internal/ui"
)

func TestSystemInfoRowsFallback(t *testing.T) {
	rows := systemInfoRows(infosys.SystemInfo{})
	require.Len(t, rows, 18)

	for _, row := range rows {
		assert.Contains(t, row.Value, infosys.NA, "row %q should render n/a", row.Label)
	}
}

func TestSystemInfoRowsValues(t *testing.T) {
	info := infosys.SystemInfo{
		Hostname:        "testhost",
		OS:              "linux",
		Distro:          "Ubuntu 24.04 LTS",
		Kernel:          "6.8.0",
		Arch:            "amd64",
		CPUModel:        "Fake CPU",
		CPUCoresLogical: 4,
		MemTotalBytes:   8 * 1024 * 1024 * 1024,
		MemUsedBytes:    2 * 1024 * 1024 * 1024,
		MemUsedPercent:  25,
		DiskTotalBytes:  100 * 1024 * 1024 * 1024,
		DiskUsedBytes:   10 * 1024 * 1024 * 1024,
		DiskUsedPercent: 10,
		UptimeSeconds:   3661,
		LocalIPs:        []string{"192.168.1.2"},
		Gateway:         "192.168.1.1",
		DNS:             []string{"1.1.1.1"},
		PublicIP:        "203.0.113.5",
		TailscaleIP:     "100.1.2.3",
		GPU:             "Fake GPU",
		Battery:         "87% (discharging)",
		Temperature:     "48.2°C",
		DeviceModel:     "Test Model",
	}

	byLabel := map[string]string{}
	for _, row := range systemInfoRows(info) {
		byLabel[row.Label] = row.Value
	}

	assert.Equal(t, "testhost", byLabel["Hostname"])
	assert.Equal(t, "Fake CPU (4 cores)", byLabel["CPU"])
	assert.Contains(t, byLabel["Memory"], "25%")
	assert.Contains(t, byLabel["Uptime"], "1h 1m")
	assert.Equal(t, "192.168.1.2", byLabel["Local IPs"])
	assert.Equal(t, "203.0.113.5", byLabel["Public IP"])
}

func TestRunSysinfoInvalidOutput(t *testing.T) {
	origOutput := sysinfoOutput
	origTimeout := sysinfoTimeout

	defer func() {
		sysinfoOutput = origOutput
		sysinfoTimeout = origTimeout
	}()

	sysinfoOutput = "xml"
	err := runSysinfo(&cobra.Command{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --output")
}

func TestRunSysinfoFormats(t *testing.T) {
	origOutput := sysinfoOutput
	origTimeout := sysinfoTimeout
	origOut := log.Out
	origDisable := ui.DisableProgress

	defer func() {
		sysinfoOutput = origOutput
		sysinfoTimeout = origTimeout
		log.Out = origOut
		ui.DisableProgress = origDisable
	}()

	ui.DisableProgress = true
	sysinfoTimeout = 15 * time.Second

	for _, format := range []string{"table", "json", "yaml"} {
		var buf bytes.Buffer
		log.Out = &buf
		sysinfoOutput = format

		err := runSysinfo(&cobra.Command{})
		require.NoError(t, err, "format %s", format)

		output := buf.String()
		switch format {
		case "table":
			assert.Contains(t, output, "FIELD", "format %s", format)
			assert.Contains(t, output, "Hostname", "format %s", format)
		case "json":
			var decoded map[string]any
			require.NoError(t, json.Unmarshal([]byte(output), &decoded), "format %s", format)
			assert.Contains(t, decoded, "hostname")
			assert.Contains(t, decoded, "cpu_model")
		case "yaml":
			assert.Contains(t, output, "hostname:")
			assert.Contains(t, output, "cpu_model:")
		}

		if !strings.Contains(output, "Hostname") && format == "table" {
			t.Errorf("table output missing Hostname row")
		}
	}
}
