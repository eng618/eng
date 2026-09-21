package sysinfo

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/sensors"
	"golang.org/x/sync/errgroup"

	"github.com/eng618/eng/internal/execx"
)

// NA is the display sentinel for values that could not be determined.
const NA = "n/a"

var (
	// execCommandContext runs OS helper commands with cancellation. Mockable for testing.
	execCommandContext = execx.CommandContext
	// osHostname reports the machine hostname. Mockable for testing.
	osHostname = os.Hostname
	// hostInfoWithContext reports host identity and kernel details. Mockable for testing.
	hostInfoWithContext = host.InfoWithContext
	// hostUptimeWithContext reports seconds since boot. Mockable for testing.
	hostUptimeWithContext = host.UptimeWithContext
	// cpuInfoWithContext reports CPU model details. Mockable for testing.
	cpuInfoWithContext = cpu.InfoWithContext
	// cpuCountsWithContext reports logical or physical core counts. Mockable for testing.
	cpuCountsWithContext = cpu.CountsWithContext
	// virtualMemoryWithContext reports RAM statistics. Mockable for testing.
	virtualMemoryWithContext = mem.VirtualMemoryWithContext
	// diskUsageWithContext reports filesystem usage for a path. Mockable for testing.
	diskUsageWithContext = disk.UsageWithContext
	// sensorsTemperaturesWithContext reports hardware temperature sensors. Mockable for testing.
	sensorsTemperaturesWithContext = sensors.TemperaturesWithContext
	// netInterfaces lists network interfaces. Mockable for testing.
	netInterfaces = net.Interfaces
	// publicIPURL is the endpoint used for public IP lookup.
	publicIPURL = "https://api.ipify.org"
	// httpClient performs the public IP request. Mockable for testing.
	httpClient = &http.Client{}
	// perCallTimeout bounds each individual collector call.
	perCallTimeout = 4 * time.Second
)

// SystemInfo holds best-effort diagnostics for one machine. Empty strings,
// zeroes, and empty slices mean "could not be determined" and render as NA.
type SystemInfo struct {
	Hostname         string   `json:"hostname"           yaml:"hostname"`
	OS               string   `json:"os"                 yaml:"os"`
	Distro           string   `json:"distro"             yaml:"distro"`
	Kernel           string   `json:"kernel"             yaml:"kernel"`
	Arch             string   `json:"arch"               yaml:"arch"`
	CPUModel         string   `json:"cpu_model"          yaml:"cpu_model"`
	CPUCoresLogical  int      `json:"cpu_cores_logical"  yaml:"cpu_cores_logical"`
	CPUCoresPhysical int      `json:"cpu_cores_physical" yaml:"cpu_cores_physical"`
	MemTotalBytes    uint64   `json:"mem_total_bytes"    yaml:"mem_total_bytes"`
	MemUsedBytes     uint64   `json:"mem_used_bytes"     yaml:"mem_used_bytes"`
	MemUsedPercent   float64  `json:"mem_used_percent"   yaml:"mem_used_percent"`
	DiskPath         string   `json:"disk_path"          yaml:"disk_path"`
	DiskTotalBytes   uint64   `json:"disk_total_bytes"   yaml:"disk_total_bytes"`
	DiskUsedBytes    uint64   `json:"disk_used_bytes"    yaml:"disk_used_bytes"`
	DiskFreeBytes    uint64   `json:"disk_free_bytes"    yaml:"disk_free_bytes"`
	DiskUsedPercent  float64  `json:"disk_used_percent"  yaml:"disk_used_percent"`
	UptimeSeconds    uint64   `json:"uptime_seconds"     yaml:"uptime_seconds"`
	LocalIPs         []string `json:"local_ips"          yaml:"local_ips"`
	Gateway          string   `json:"gateway"            yaml:"gateway"`
	DNS              []string `json:"dns"                yaml:"dns"`
	PublicIP         string   `json:"public_ip"          yaml:"public_ip"`
	TailscaleIP      string   `json:"tailscale_ip"       yaml:"tailscale_ip"`
	GPU              string   `json:"gpu"                yaml:"gpu"`
	Battery          string   `json:"battery"            yaml:"battery"`
	Temperature      string   `json:"temperature"        yaml:"temperature"`
	DeviceModel      string   `json:"device_model"       yaml:"device_model"`
}

// Collect gathers best-effort system diagnostics. It never returns an error:
// fields that cannot be determined keep their zero values for NA rendering.
func Collect(ctx context.Context) SystemInfo {
	info := SystemInfo{Arch: runtime.GOARCH, DiskPath: "/"}

	if hostname, err := osHostname(); err == nil && hostname != "" {
		info.Hostname = hostname
	}

	distro := Detect()
	info.OS = distro.RawOS
	info.Distro = distro.PrettyName

	group := errgroup.Group{}

	var cpuModel string

	var cpuLogical, cpuPhysical int

	group.Go(func() error {
		cctx, cancel := scoped(ctx)
		defer cancel()
		cpuModel, cpuLogical, cpuPhysical = collectCPU(cctx)

		return nil
	})

	var memTotal, memUsed uint64

	var memPct float64

	group.Go(func() error {
		cctx, cancel := scoped(ctx)
		defer cancel()
		memTotal, memUsed, memPct = collectMemory(cctx)

		return nil
	})

	var diskTotal, diskUsed, diskFree uint64

	var diskPct float64

	group.Go(func() error {
		cctx, cancel := scoped(ctx)
		defer cancel()
		diskTotal, diskUsed, diskFree, diskPct = collectDisk(cctx)

		return nil
	})

	var uptime uint64

	var kernel string

	group.Go(func() error {
		cctx, cancel := scoped(ctx)
		defer cancel()
		uptime = collectUptime(cctx)
		kernel = collectKernel(cctx)

		return nil
	})

	var localIPs []string

	var gateway string

	var dns []string

	group.Go(func() error {
		localIPs = collectLocalIPs()
		gateway = collectGateway(ctx)
		dns = collectDNS(ctx)

		return nil
	})

	var publicIP, tailscaleIP string

	group.Go(func() error {
		cctx, cancel := scoped(ctx)
		defer cancel()
		publicIP = collectPublicIP(cctx)
		tailscaleIP = collectTailscaleIP(ctx)

		return nil
	})

	var gpu, battery, temperature, deviceModel string

	group.Go(func() error {
		cctx, cancel := scoped(ctx)
		defer cancel()
		gpu = collectGPU(ctx)
		battery = collectBattery(ctx)
		temperature = collectTemperature(cctx)
		deviceModel = collectDeviceModel(ctx)

		return nil
	})

	_ = group.Wait()

	info.CPUModel = cpuModel
	info.CPUCoresLogical = cpuLogical
	info.CPUCoresPhysical = cpuPhysical
	info.MemTotalBytes = memTotal
	info.MemUsedBytes = memUsed
	info.MemUsedPercent = memPct
	info.DiskTotalBytes = diskTotal
	info.DiskUsedBytes = diskUsed
	info.DiskFreeBytes = diskFree
	info.DiskUsedPercent = diskPct
	info.UptimeSeconds = uptime
	info.Kernel = kernel
	info.LocalIPs = localIPs
	info.Gateway = gateway
	info.DNS = dns
	info.PublicIP = publicIP
	info.TailscaleIP = tailscaleIP
	info.GPU = gpu
	info.Battery = battery
	info.Temperature = temperature
	info.DeviceModel = deviceModel

	return info
}

// scoped derives a per-collector timeout from the parent context.
func scoped(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, perCallTimeout)
}

// runCmd executes a helper binary and returns trimmed stdout.
func runCmd(ctx context.Context, name string, args ...string) (string, error) {
	cctx, cancel := context.WithTimeout(ctx, perCallTimeout)
	defer cancel()

	out, err := execCommandContext(cctx, name, args...).Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

// CPUDisplay renders "model (N cores)" with graceful degradation.
func (s SystemInfo) CPUDisplay() string {
	switch {
	case s.CPUModel != "" && s.CPUCoresLogical > 0:
		return fmt.Sprintf("%s (%d cores)", s.CPUModel, s.CPUCoresLogical)
	case s.CPUModel != "":
		return s.CPUModel
	case s.CPUCoresLogical > 0:
		return fmt.Sprintf("%d cores", s.CPUCoresLogical)
	default:
		return ""
	}
}

// MemoryDisplay renders "used / total (pct%)" or "total" when use is unknown.
func (s SystemInfo) MemoryDisplay() string {
	if s.MemTotalBytes == 0 {
		return ""
	}
	if s.MemUsedBytes == 0 {
		return humanize.Bytes(s.MemTotalBytes)
	}

	return fmt.Sprintf(
		"%s / %s (%.0f%%)",
		humanize.Bytes(s.MemUsedBytes),
		humanize.Bytes(s.MemTotalBytes),
		s.MemUsedPercent,
	)
}

// DiskDisplay renders "used / total (pct%)" or "total" when use is unknown.
func (s SystemInfo) DiskDisplay() string {
	if s.DiskTotalBytes == 0 {
		return ""
	}
	if s.DiskUsedBytes == 0 {
		return humanize.Bytes(s.DiskTotalBytes)
	}

	return fmt.Sprintf(
		"%s / %s (%.0f%%)",
		humanize.Bytes(s.DiskUsedBytes),
		humanize.Bytes(s.DiskTotalBytes),
		s.DiskUsedPercent,
	)
}

// UptimeDisplay renders seconds as "3d 4h 12m".
func (s SystemInfo) UptimeDisplay() string {
	return FormatUptime(s.UptimeSeconds)
}

// FormatUptime renders seconds in compact human form.
func FormatUptime(seconds uint64) string {
	if seconds == 0 {
		return ""
	}

	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	switch {
	case days > 0:
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	case hours > 0:
		return fmt.Sprintf("%dh %dm", hours, minutes)
	case minutes > 0:
		return fmt.Sprintf("%dm %ds", minutes, secs)
	default:
		return fmt.Sprintf("%ds", secs)
	}
}

func collectCPU(ctx context.Context) (string, int, int) {
	var model string

	var logical, physical int

	if infos, err := cpuInfoWithContext(ctx); err == nil {
		for _, ci := range infos {
			if model == "" && strings.TrimSpace(ci.ModelName) != "" {
				model = strings.TrimSpace(ci.ModelName)
			}
		}
	}
	if n, err := cpuCountsWithContext(ctx, true); err == nil && n > 0 {
		logical = n
	}
	if n, err := cpuCountsWithContext(ctx, false); err == nil && n > 0 {
		physical = n
	}
	if logical == 0 {
		logical = runtime.NumCPU()
	}
	if model == "" {
		model = fallbackCPUModel(ctx)
	}

	return model, logical, physical
}

func fallbackCPUModel(ctx context.Context) string {
	if RuntimeGOOS == "darwin" {
		if out, err := runCmd(ctx, "sysctl", "-n", "machdep.cpu.brand_string"); err == nil && out != "" {
			return out
		}

		return ""
	}

	data, err := ReadFile("/proc/cpuinfo")
	if err != nil {
		return ""
	}

	return parseCPUInfoModel(string(data))
}

// parseCPUInfoModel extracts the CPU model from /proc/cpuinfo content,
// preferring "model name" with Raspberry Pi "Model" as fallback.
func parseCPUInfoModel(data string) string {
	var piModel string

	for line := range strings.Lines(data) {
		if model, ok := strings.CutPrefix(line, "model name"); ok {
			if _, val, found := strings.Cut(model, ":"); found && strings.TrimSpace(val) != "" {
				return strings.TrimSpace(val)
			}
		}
		if m, ok := strings.CutPrefix(line, "Model"); ok {
			if _, val, found := strings.Cut(m, ":"); found && strings.TrimSpace(val) != "" {
				piModel = strings.TrimSpace(val)
			}
		}
	}

	return piModel
}

func collectMemory(ctx context.Context) (uint64, uint64, float64) {
	if vm, err := virtualMemoryWithContext(ctx); err == nil && vm != nil && vm.Total > 0 {
		return vm.Total, vm.Used, vm.UsedPercent
	}

	return fallbackMemory(ctx)
}

func fallbackMemory(ctx context.Context) (uint64, uint64, float64) {
	if RuntimeGOOS == "darwin" {
		out, err := runCmd(ctx, "sysctl", "-n", "hw.memsize")
		if err != nil || out == "" {
			return 0, 0, 0
		}

		total, err := strconv.ParseUint(strings.TrimSpace(out), 10, 64)
		if err != nil {
			return 0, 0, 0
		}

		return total, 0, 0
	}

	data, err := ReadFile("/proc/meminfo")
	if err != nil {
		return 0, 0, 0
	}

	return parseProcMeminfo(string(data))
}

// parseProcMeminfo derives total/used bytes and use percent from
// /proc/meminfo content. Used is unknown when MemAvailable is absent.
func parseProcMeminfo(data string) (uint64, uint64, float64) {
	var total, available uint64

	for line := range strings.Lines(data) {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		kb, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}

		switch fields[0] {
		case "MemTotal:":
			total = kb * 1024
		case "MemAvailable:":
			available = kb * 1024
		}
	}

	if total == 0 {
		return 0, 0, 0
	}

	var used uint64

	var pct float64

	if available > 0 && available <= total {
		used = total - available
		pct = float64(used) / float64(total) * 100
	}

	return total, used, pct
}

func collectDisk(ctx context.Context) (uint64, uint64, uint64, float64) {
	if usage, err := diskUsageWithContext(ctx, "/"); err == nil && usage != nil && usage.Total > 0 {
		return usage.Total, usage.Used, usage.Free, usage.UsedPercent
	}

	return fallbackDisk(ctx)
}

func fallbackDisk(ctx context.Context) (uint64, uint64, uint64, float64) {
	out, err := runCmd(ctx, "df", "-k", "/")
	if err != nil || out == "" {
		return 0, 0, 0, 0
	}

	return parseDFOutput(out)
}

// parseDFOutput extracts total/used/free bytes and use percent from
// `df -k /` output. All zeroes when the output is unparseable.
func parseDFOutput(out string) (uint64, uint64, uint64, float64) {
	lines := strings.Split(out, "\n")
	if len(lines) < 2 {
		return 0, 0, 0, 0
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 5 {
		return 0, 0, 0, 0
	}

	totalKB, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return 0, 0, 0, 0
	}

	usedKB, err := strconv.ParseUint(fields[2], 10, 64)
	if err != nil {
		return 0, 0, 0, 0
	}

	availKB, _ := strconv.ParseUint(fields[3], 10, 64)
	pct, _ := strconv.ParseFloat(strings.TrimSuffix(fields[4], "%"), 64)

	return totalKB * 1024, usedKB * 1024, availKB * 1024, pct
}

func collectUptime(ctx context.Context) uint64 {
	if uptime, err := hostUptimeWithContext(ctx); err == nil && uptime > 0 {
		return uptime
	}

	return fallbackUptime(ctx)
}

func fallbackUptime(ctx context.Context) uint64 {
	if RuntimeGOOS == "darwin" {
		out, err := runCmd(ctx, "sysctl", "-n", "kern.boottime")
		if err != nil || out == "" {
			return 0
		}

		boot, ok := parseBoottimeSec(out)
		if !ok {
			return 0
		}

		now := uint64(time.Now().Unix())
		if now > boot {
			return now - boot
		}

		return 0
	}

	data, err := ReadFile("/proc/uptime")
	if err != nil {
		return 0
	}

	return parseProcUptime(string(data))
}

// parseBoottimeSec extracts boot epoch seconds from
// `sysctl -n kern.boottime` output like "{ sec = 1234, usec = 0 } ...".
func parseBoottimeSec(out string) (uint64, bool) {
	matches := regexp.MustCompile(`sec = (\d+)`).FindStringSubmatch(out)
	if len(matches) != 2 {
		return 0, false
	}

	boot, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return 0, false
	}

	return boot, true
}

// parseProcUptime extracts uptime seconds from /proc/uptime content.
func parseProcUptime(data string) uint64 {
	fields := strings.Fields(data)
	if len(fields) == 0 {
		return 0
	}

	seconds, err := strconv.ParseFloat(fields[0], 64)
	if err != nil || seconds <= 0 {
		return 0
	}

	return uint64(seconds)
}

func collectKernel(ctx context.Context) string {
	if hi, err := hostInfoWithContext(ctx); err == nil && hi != nil {
		if kernel := strings.TrimSpace(hi.KernelVersion); kernel != "" {
			return kernel
		}
	}

	if out, err := runCmd(ctx, "uname", "-r"); err == nil && out != "" {
		return out
	}

	return ""
}

func collectLocalIPs() []string {
	ifaces, err := netInterfaces()
	if err != nil || len(ifaces) == 0 {
		return fallbackLocalIPs()
	}

	var ips []string

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP

			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
				continue
			}

			ips = append(ips, ip.String())
		}
	}

	if len(ips) == 0 {
		return fallbackLocalIPs()
	}

	return dedupe(ips)
}

func fallbackLocalIPs() []string {
	ctx := context.Background()

	if RuntimeGOOS == "darwin" {
		var ips []string

		for _, iface := range []string{"en0", "en1"} {
			if out, err := runCmd(ctx, "ipconfig", "getifaddr", iface); err == nil && out != "" {
				if ip := net.ParseIP(out); ip != nil {
					ips = append(ips, ip.String())
				}
			}
		}

		return ips
	}

	out, err := runCmd(ctx, "hostname", "-I")
	if err != nil || out == "" {
		return nil
	}

	var ips []string

	for _, field := range strings.Fields(out) {
		if ip := net.ParseIP(field); ip != nil && !ip.IsLoopback() {
			ips = append(ips, ip.String())
		}
	}

	return ips
}

func collectGateway(ctx context.Context) string {
	if RuntimeGOOS == "darwin" {
		if out, err := runCmd(ctx, "route", "-n", "get", "default"); err == nil && out != "" {
			if gw := parseDarwinGatewayRoute(out); gw != "" {
				return gw
			}
		}

		return gatewayFromNetstat(ctx)
	}

	if out, err := runCmd(ctx, "ip", "route", "show", "default"); err == nil && out != "" {
		if gw := parseLinuxIPRoute(out); gw != "" {
			return gw
		}
	}

	out, err := runCmd(ctx, "route", "-n")
	if err != nil || out == "" {
		return ""
	}

	return parseLinuxRouteN(out)
}

// parseDarwinGatewayRoute extracts the gateway from
// `route -n get default` output.
func parseDarwinGatewayRoute(out string) string {
	for line := range strings.Lines(out) {
		line = strings.TrimSpace(line)
		if gw, ok := strings.CutPrefix(line, "gateway:"); ok {
			if ip := net.ParseIP(strings.TrimSpace(gw)); ip != nil {
				return ip.String()
			}
		}
	}

	return ""
}

// parseLinuxIPRoute extracts the gateway from `ip route show default` output.
func parseLinuxIPRoute(out string) string {
	for line := range strings.Lines(out) {
		fields := strings.Fields(line)
		for i, field := range fields {
			if field == "via" && i+1 < len(fields) {
				if ip := net.ParseIP(fields[i+1]); ip != nil {
					return ip.String()
				}
			}
		}
	}

	return ""
}

// parseLinuxRouteN extracts the gateway from `route -n` output.
func parseLinuxRouteN(out string) string {
	for line := range strings.Lines(out) {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "0.0.0.0" {
			if ip := net.ParseIP(fields[1]); ip != nil {
				return ip.String()
			}
		}
	}

	return ""
}

// gatewayFromNetstat parses the IPv4 default gateway from netstat routing
// output. Used on macOS when the default route has no gateway of its own,
// for example while a VPN owns the default route.
func gatewayFromNetstat(ctx context.Context) string {
	out, err := runCmd(ctx, "netstat", "-rn", "-f", "inet")
	if err != nil || out == "" {
		return ""
	}

	return parseNetstatGateway(out)
}

// parseNetstatGateway extracts the first non-link-local IPv4 default gateway
// from `netstat -rn -f inet` output.
func parseNetstatGateway(out string) string {
	for line := range strings.Lines(out) {
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[0] != "default" {
			continue
		}

		if ip := net.ParseIP(fields[1]); ip != nil && !ip.IsLinkLocalUnicast() {
			return ip.String()
		}
	}

	return ""
}

func collectDNS(ctx context.Context) []string {
	var servers []string

	if data, err := ReadFile("/etc/resolv.conf"); err == nil {
		servers = append(servers, parseResolvConf(string(data))...)
	}

	if RuntimeGOOS == "darwin" {
		if out, err := runCmd(ctx, "scutil", "--dns"); err == nil && out != "" {
			servers = append(servers, parseScutilDNS(out)...)
		}
	}

	return dedupe(servers)
}

// parseResolvConf extracts nameserver IPs from resolv.conf content.
func parseResolvConf(data string) []string {
	var servers []string

	for line := range strings.Lines(data) {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "nameserver" {
			if ip := net.ParseIP(fields[1]); ip != nil {
				servers = append(servers, ip.String())
			}
		}
	}

	return servers
}

// parseScutilDNS extracts nameserver IPs from `scutil --dns` output.
func parseScutilDNS(out string) []string {
	var servers []string

	for line := range strings.Lines(out) {
		line = strings.TrimSpace(line)
		if _, val, found := strings.Cut(line, ":"); found && strings.Contains(line, "nameserver") {
			if ip := net.ParseIP(strings.TrimSpace(val)); ip != nil {
				servers = append(servers, ip.String())
			}
		}
	}

	return servers
}

func collectPublicIP(ctx context.Context) string {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, publicIPURL, nil)
	if err != nil {
		return ""
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return ""
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64))
	if err != nil {
		return ""
	}

	if ip := net.ParseIP(strings.TrimSpace(string(body))); ip != nil {
		return ip.String()
	}

	return ""
}

func collectTailscaleIP(ctx context.Context) string {
	if _, err := LookPath("tailscale"); err != nil {
		return ""
	}

	out, err := runCmd(ctx, "tailscale", "ip", "-4")
	if err != nil || out == "" {
		return ""
	}

	first, _, _ := strings.Cut(out, "\n")
	if ip := net.ParseIP(strings.TrimSpace(first)); ip != nil {
		return ip.String()
	}

	return ""
}

func collectGPU(ctx context.Context) string {
	if RuntimeGOOS == "darwin" {
		out, err := runCmd(ctx, "system_profiler", "SPDisplaysDataType")
		if err != nil || out == "" {
			return ""
		}

		return parseSystemProfilerDisplays(out)
	}

	if _, err := LookPath("lspci"); err != nil {
		return ""
	}

	out, err := runCmd(ctx, "lspci")
	if err != nil || out == "" {
		return ""
	}

	return parseLspci(out)
}

// parseSystemProfilerDisplays extracts GPU models from
// `system_profiler SPDisplaysDataType` output.
func parseSystemProfilerDisplays(out string) string {
	var models []string

	for line := range strings.Lines(out) {
		line = strings.TrimSpace(line)
		if model, ok := strings.CutPrefix(line, "Chipset Model:"); ok {
			if model = strings.TrimSpace(model); model != "" {
				models = append(models, model)
			}
		}
	}

	return strings.Join(dedupe(models), " / ")
}

// parseLspci extracts GPU models from `lspci` output.
func parseLspci(out string) string {
	var models []string

	for line := range strings.Lines(out) {
		lower := strings.ToLower(line)
		if !strings.Contains(lower, "vga compatible controller") &&
			!strings.Contains(lower, "3d controller") &&
			!strings.Contains(lower, "display controller") {
			continue
		}

		if _, model, found := strings.Cut(line, ": "); found && strings.TrimSpace(model) != "" {
			models = append(models, strings.TrimSpace(model))
		}
	}

	return strings.Join(dedupe(models), " / ")
}

func collectBattery(ctx context.Context) string {
	if RuntimeGOOS == "darwin" {
		out, err := runCmd(ctx, "pmset", "-g", "batt")
		if err != nil || out == "" {
			return ""
		}

		return parsePmsetBatt(out)
	}

	for _, bat := range []string{"BAT0", "BAT1"} {
		capacity, err := ReadFile("/sys/class/power_supply/" + bat + "/capacity")
		if err != nil {
			continue
		}

		pct := strings.TrimSpace(string(capacity))
		if pct == "" {
			continue
		}

		state := ""
		if data, err := ReadFile("/sys/class/power_supply/" + bat + "/status"); err == nil {
			state = strings.ToLower(strings.TrimSpace(string(data)))
		}

		if state != "" {
			return pct + "% (" + state + ")"
		}

		return pct + "%"
	}

	return ""
}

// parsePmsetBatt extracts "NN% (state)" from `pmset -g batt` output.
func parsePmsetBatt(out string) string {
	pct := regexp.MustCompile(`(\d+)%`).FindStringSubmatch(out)
	if len(pct) != 2 {
		return ""
	}

	state := "battery"
	lower := strings.ToLower(out)

	switch {
	case strings.Contains(lower, "discharging"):
		state = "discharging"
	case strings.Contains(lower, "charging"):
		state = "charging"
	case strings.Contains(lower, "charged"):
		state = "charged"
	}

	return pct[1] + "% (" + state + ")"
}

func collectTemperature(ctx context.Context) string {
	if temps, err := sensorsTemperaturesWithContext(ctx); err == nil {
		if temp, label := pickTemperature(temps); temp > 0 {
			if label != "" {
				return fmt.Sprintf("%.1f°C (%s)", temp, label)
			}

			return fmt.Sprintf("%.1f°C", temp)
		}
	}

	if RuntimeGOOS == "linux" {
		if data, err := ReadFile("/sys/class/thermal/thermal_zone0/temp"); err == nil {
			if temp, ok := parseThermalZoneMillis(string(data)); ok {
				return temp
			}
		}

		if out, err := runCmd(ctx, "vcgencmd", "measure_temp"); err == nil && out != "" {
			if temp := parseVcgencmdTemp(out); temp != "" {
				return temp
			}
		}
	}

	return ""
}

// parseThermalZoneMillis formats millidegree-Celsius sysfs content as "48.2°C".
func parseThermalZoneMillis(data string) (string, bool) {
	milli, err := strconv.ParseFloat(strings.TrimSpace(data), 64)
	if err != nil || milli <= 0 {
		return "", false
	}

	return fmt.Sprintf("%.1f°C", milli/1000), true
}

// parseVcgencmdTemp extracts the temperature from
// `vcgencmd measure_temp` output like "temp=48.2'C".
func parseVcgencmdTemp(out string) string {
	if matches := regexp.MustCompile(`([\d.]+)'?C`).FindStringSubmatch(out); len(matches) == 2 {
		return matches[1] + "°C"
	}

	return ""
}

// cpuSensorHints matches sensor keys that most likely report SoC/CPU
// temperature. Charger, battery, and NAND sensors run hot but mislead.
var cpuSensorHints = []string{"cpu", "coretemp", "k10temp", "package", "die", "tdie", "thermal"}

// pickTemperature returns the hottest CPU-like sensor reading, falling back
// to the hottest sensor overall when no CPU-like key matches.
func pickTemperature(temps []sensors.TemperatureStat) (float64, string) {
	var best float64

	var label string

	for _, t := range temps {
		if t.Temperature > best {
			best = t.Temperature
			label = t.SensorKey
		}
	}

	var cpuBest float64

	var cpuLabel string

	for _, t := range temps {
		key := strings.ToLower(t.SensorKey)
		matched := false

		for _, hint := range cpuSensorHints {
			if strings.Contains(key, hint) {
				matched = true

				break
			}
		}

		if matched && t.Temperature > cpuBest {
			cpuBest = t.Temperature
			cpuLabel = t.SensorKey
		}
	}

	if cpuBest > 0 {
		return cpuBest, cpuLabel
	}

	return best, label
}

func collectDeviceModel(ctx context.Context) string {
	if RuntimeGOOS == "darwin" {
		if out, err := runCmd(ctx, "sysctl", "-n", "hw.model"); err == nil && out != "" {
			return out
		}

		return ""
	}

	for _, path := range []string{
		"/proc/device-tree/model",
		"/sys/firmware/devicetree/base/model",
		"/sys/devices/virtual/dmi/id/product_name",
	} {
		if data, err := ReadFile(path); err == nil {
			if model := strings.Trim(strings.TrimSpace(string(data)), "\x00 "); model != "" {
				return model
			}
		}
	}

	return ""
}

func dedupe(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))

	for _, v := range values {
		if _, ok := seen[v]; ok {
			continue
		}

		seen[v] = struct{}{}
		out = append(out, v)
	}

	return out
}
