package sysinfo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/eng618/eng/internal/log"
	infosys "github.com/eng618/eng/internal/sysinfo"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

var (
	sysinfoOutput  string
	sysinfoTimeout time.Duration
)

// SysinfoCmd prints cross-platform system diagnostics in a formatted table.
var SysinfoCmd = &cobra.Command{
	Use:   "sysinfo",
	Short: "Show system diagnostics (CPU, memory, disk, network)",
	Long: `Collects best-effort system diagnostics and prints them as a formatted table.

Works across Raspberry Pi, Fedora, Ubuntu, and macOS. Fields that cannot be
determined render as "n/a", so the command always succeeds.`,
	Example: `  eng sysinfo
  eng sysinfo --output json
  eng sysinfo --output yaml --timeout 5s`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runSysinfo(cmd)
	},
}

func runSysinfo(cmd *cobra.Command) error {
	format := strings.ToLower(strings.TrimSpace(sysinfoOutput))

	switch format {
	case "table", "json", "yaml":
	default:
		return fmt.Errorf("invalid --output %q: must be table, json, or yaml", sysinfoOutput)
	}

	timeout := sysinfoTimeout
	if timeout <= 0 {
		timeout = 8 * time.Second
	}

	parent := cmd.Context()
	if parent == nil {
		parent = context.Background()
	}

	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	info := infosys.Collect(ctx)

	switch format {
	case "json":
		out, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			return err
		}

		fmt.Fprintln(log.Out, string(out))

		return nil
	case "yaml":
		out, err := yaml.Marshal(info)
		if err != nil {
			return err
		}

		fmt.Fprintln(log.Out, string(out))

		return nil
	}

	rendered := ui.RenderKeyValueTable("", systemInfoRows(info), ui.GetTerminalWidth())

	if !ui.DisableProgress {
		theme.InfoMessage("System Information:")
		fmt.Fprintln(log.Out)
		fmt.Fprintln(log.Out, rendered)
	} else {
		log.Message("%s", rendered)
	}

	return nil
}

// systemInfoRows maps diagnostics to presentation rows, rendering unknowns as muted "n/a".
func systemInfoRows(info infosys.SystemInfo) []ui.KeyValueRow {
	orNA := func(value string) string {
		if strings.TrimSpace(value) == "" {
			return theme.MutedText.Render(infosys.NA)
		}

		return value
	}
	joinOrNA := func(values []string) string {
		if len(values) == 0 {
			return theme.MutedText.Render(infosys.NA)
		}

		return strings.Join(values, ", ")
	}

	return []ui.KeyValueRow{
		{Label: "Hostname", Value: orNA(info.Hostname)},
		{Label: "OS", Value: orNA(info.OS)},
		{Label: "Distro", Value: orNA(info.Distro)},
		{Label: "Kernel", Value: orNA(info.Kernel)},
		{Label: "Arch", Value: orNA(info.Arch)},
		{Label: "CPU", Value: orNA(info.CPUDisplay())},
		{Label: "Memory", Value: orNA(info.MemoryDisplay())},
		{Label: "Disk (/)", Value: orNA(info.DiskDisplay())},
		{Label: "Uptime", Value: orNA(info.UptimeDisplay())},
		{Label: "Local IPs", Value: joinOrNA(info.LocalIPs)},
		{Label: "Gateway", Value: orNA(info.Gateway)},
		{Label: "DNS", Value: joinOrNA(info.DNS)},
		{Label: "Public IP", Value: orNA(info.PublicIP)},
		{Label: "Tailscale IP", Value: orNA(info.TailscaleIP)},
		{Label: "GPU", Value: orNA(info.GPU)},
		{Label: "Battery", Value: orNA(info.Battery)},
		{Label: "Temperature", Value: orNA(info.Temperature)},
		{Label: "Device Model", Value: orNA(info.DeviceModel)},
	}
}

func init() {
	SysinfoCmd.Flags().
		StringVarP(&sysinfoOutput, "output", "o", "table", "Output format: table, json, or yaml")
	SysinfoCmd.Flags().
		DurationVar(&sysinfoTimeout, "timeout", 8*time.Second, "Overall timeout for diagnostics collection")
}
