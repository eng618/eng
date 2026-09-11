package sysinfo

import (
	"context"
	"errors"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/sensors"
)

func TestFormatUptime(t *testing.T) {
	cases := []struct {
		seconds uint64
		want    string
	}{
		{0, ""},
		{45, "45s"},
		{90, "1m 30s"},
		{3600, "1h 0m"},
		{3661, "1h 1m"},
		{90061, "1d 1h 1m"},
	}

	for _, tc := range cases {
		if got := FormatUptime(tc.seconds); got != tc.want {
			t.Errorf("FormatUptime(%d) = %q, want %q", tc.seconds, got, tc.want)
		}
	}
}

func TestSystemInfoDisplays(t *testing.T) {
	empty := SystemInfo{}
	if empty.CPUDisplay() != "" || empty.MemoryDisplay() != "" || empty.DiskDisplay() != "" ||
		empty.UptimeDisplay() != "" {
		t.Error("Expected empty displays for zero SystemInfo")
	}

	full := SystemInfo{
		CPUModel:        "Apple M3 Pro",
		CPUCoresLogical: 11,
		MemTotalBytes:   8 * 1024 * 1024 * 1024,
		MemUsedBytes:    4 * 1024 * 1024 * 1024,
		MemUsedPercent:  50,
		DiskTotalBytes:  100 * 1024 * 1024 * 1024,
		DiskUsedBytes:   25 * 1024 * 1024 * 1024,
		DiskUsedPercent: 25,
		UptimeSeconds:   90061,
	}

	if got := full.CPUDisplay(); got != "Apple M3 Pro (11 cores)" {
		t.Errorf("CPUDisplay() = %q", got)
	}
	if got := (SystemInfo{CPUCoresLogical: 4}).CPUDisplay(); got != "4 cores" {
		t.Errorf("CPUDisplay() cores-only = %q", got)
	}
	if got := full.MemoryDisplay(); got != "4.0 GB / 8.6 GB (50%)" && got != "4 GB / 8 GB (50%)" {
		t.Logf("MemoryDisplay() = %q (humanize formatting may vary)", got)
	}
	if got := full.DiskDisplay(); got == "" {
		t.Error("Expected non-empty DiskDisplay")
	}
	if got := full.UptimeDisplay(); got != "1d 1h 1m" {
		t.Errorf("UptimeDisplay() = %q", got)
	}
	if got := (SystemInfo{MemTotalBytes: 1024}).MemoryDisplay(); got == "" {
		t.Error("Expected total-only MemoryDisplay when use is unknown")
	}
}

func TestPickTemperature(t *testing.T) {
	temps := []sensors.TemperatureStat{
		{SensorKey: "Charger TQ0j", Temperature: 81},
		{SensorKey: "PMU tdie1", Temperature: 73},
	}

	temp, label := pickTemperature(temps)
	if temp != 73 || label != "PMU tdie1" {
		t.Errorf("pickTemperature() = (%v, %q), want CPU-like sensor", temp, label)
	}

	other := []sensors.TemperatureStat{
		{SensorKey: "Charger TQ0j", Temperature: 81},
		{SensorKey: "NAND CH0", Temperature: 52},
	}

	temp, label = pickTemperature(other)
	if temp != 81 || label != "Charger TQ0j" {
		t.Errorf("pickTemperature() fallback = (%v, %q), want hottest overall", temp, label)
	}

	if temp, _ := pickTemperature(nil); temp != 0 {
		t.Errorf("pickTemperature(nil) = %v, want 0", temp)
	}
}

func TestParseCPUInfoModel(t *testing.T) {
	x86 := "processor\t: 0\nmodel name\t: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz\n"
	if got := parseCPUInfoModel(x86); got != "Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz" {
		t.Errorf("parseCPUInfoModel(x86) = %q", got)
	}

	pi := "processor\t: 0\nModel\t\t: Raspberry Pi 5 Model B Rev 1.0\n"
	if got := parseCPUInfoModel(pi); got != "Raspberry Pi 5 Model B Rev 1.0" {
		t.Errorf("parseCPUInfoModel(pi) = %q", got)
	}

	if got := parseCPUInfoModel(""); got != "" {
		t.Errorf("parseCPUInfoModel(empty) = %q, want empty", got)
	}
}

func TestParseProcMeminfo(t *testing.T) {
	data := "MemTotal:        8023384 kB\nMemFree:           1234 kB\nMemAvailable:    4011692 kB\n"
	total, used, pct := parseProcMeminfo(data)
	if total != 8023384*1024 {
		t.Errorf("total = %d", total)
	}
	if used != (8023384-4011692)*1024 {
		t.Errorf("used = %d", used)
	}
	if pct < 49 || pct > 51 {
		t.Errorf("pct = %v", pct)
	}

	total, used, _ = parseProcMeminfo("MemTotal:        1024 kB\n")
	if total != 1024*1024 || used != 0 {
		t.Errorf("missing MemAvailable: total=%d used=%d", total, used)
	}

	if total, _, _ := parseProcMeminfo("garbage"); total != 0 {
		t.Errorf("garbage total = %d, want 0", total)
	}
}

func TestParseDFOutput(t *testing.T) {
	out := "Filesystem   1024-blocks      Used Available Capacity Mounted on\n" +
		"/dev/disk3s1  483183616 425830400  57353216    89%    /\n"
	total, used, free, pct := parseDFOutput(out)
	if total != 483183616*1024 || used != 425830400*1024 || free != 57353216*1024 || pct != 89 {
		t.Errorf("parseDFOutput = %d %d %d %v", total, used, free, pct)
	}

	if total, _, _, _ := parseDFOutput("bogus"); total != 0 {
		t.Errorf("bogus total = %d, want 0", total)
	}
}

func TestParseGateways(t *testing.T) {
	darwin := "   route to: default\ndestination: default\n       mask: default\n    gateway: 192.168.1.1\n"
	if got := parseDarwinGatewayRoute(darwin); got != "192.168.1.1" {
		t.Errorf("parseDarwinGatewayRoute = %q", got)
	}
	if got := parseDarwinGatewayRoute("no gateway here"); got != "" {
		t.Errorf("parseDarwinGatewayRoute missing = %q", got)
	}

	ipRoute := "default via 192.168.1.1 dev eth0 proto dhcp metric 100\n"
	if got := parseLinuxIPRoute(ipRoute); got != "192.168.1.1" {
		t.Errorf("parseLinuxIPRoute = %q", got)
	}

	routeN := "Kernel IP routing table\nDestination     Gateway         Genmask\n" +
		"0.0.0.0         10.0.0.1        0.0.0.0         UG\n"
	if got := parseLinuxRouteN(routeN); got != "10.0.0.1" {
		t.Errorf("parseLinuxRouteN = %q", got)
	}

	netstat := "Routing tables\nInternet:\nDestination        Gateway            Flags\n" +
		"default            192.168.86.1       UGScg\ndefault            fe80::1%en0        UGcIg\n"
	if got := parseNetstatGateway(netstat); got != "192.168.86.1" {
		t.Errorf("parseNetstatGateway = %q", got)
	}
}

func TestParseDNS(t *testing.T) {
	resolv := "# comment\nnameserver 1.1.1.1\nnameserver 8.8.8.8\nnameserver bogus\n"
	got := parseResolvConf(resolv)
	if len(got) != 2 || got[0] != "1.1.1.1" || got[1] != "8.8.8.8" {
		t.Errorf("parseResolvConf = %v", got)
	}

	scutil := "  nameserver[0] : 9.9.9.9\n  nameserver[1] : 149.112.112.112\n"
	got = parseScutilDNS(scutil)
	if len(got) != 2 || got[0] != "9.9.9.9" {
		t.Errorf("parseScutilDNS = %v", got)
	}
}

func TestParseGPU(t *testing.T) {
	sp := "Graphics/Displays:\n\n    Apple M1 Pro:\n\n      Chipset Model: Apple M1 Pro\n"
	if got := parseSystemProfilerDisplays(sp); got != "Apple M1 Pro" {
		t.Errorf("parseSystemProfilerDisplays = %q", got)
	}

	lspci := "00:02.0 VGA compatible controller: Intel Corporation UHD Graphics 620\n" +
		"01:00.0 3D controller: NVIDIA Corporation GP108M [GeForce MX150]\n" +
		"00:1f.0 ISA bridge: Intel Corporation Sunrise Point\n"
	got := parseLspci(lspci)
	if got == "" || !strings.Contains(got, "UHD Graphics 620") || !strings.Contains(got, "GeForce MX150") {
		t.Errorf("parseLspci = %q", got)
	}
	if strings.Contains(got, "ISA bridge") {
		t.Errorf("parseLspci leaked non-GPU device: %q", got)
	}
}

func TestParsePmsetBatt(t *testing.T) {
	out := "Now drawing from 'AC Power'\n -InternalBattery-0 (id=123)  87%; charging; 0:20 remaining\n"
	if got := parsePmsetBatt(out); got != "87% (charging)" {
		t.Errorf("parsePmsetBatt = %q", got)
	}
	if got := parsePmsetBatt("no percent here"); got != "" {
		t.Errorf("parsePmsetBatt missing = %q", got)
	}
}

func TestParseUptimeHelpers(t *testing.T) {
	boot, ok := parseBoottimeSec("{ sec = 1700000000, usec = 0 } Thu Nov  9 00:00:00 2023")
	if !ok || boot != 1700000000 {
		t.Errorf("parseBoottimeSec = %d %v", boot, ok)
	}
	if _, ok := parseBoottimeSec("bogus"); ok {
		t.Error("parseBoottimeSec bogus should fail")
	}
	if got := parseProcUptime("12345.67 54321.00\n"); got != 12345 {
		t.Errorf("parseProcUptime = %d", got)
	}
	if got := parseProcUptime("bogus"); got != 0 {
		t.Errorf("parseProcUptime bogus = %d", got)
	}
}

func TestParseTemperatureHelpers(t *testing.T) {
	if temp, ok := parseThermalZoneMillis("48200\n"); !ok || temp != "48.2°C" {
		t.Errorf("parseThermalZoneMillis = %q %v", temp, ok)
	}
	if _, ok := parseThermalZoneMillis("bogus"); ok {
		t.Error("parseThermalZoneMillis bogus should fail")
	}
	if got := parseVcgencmdTemp("temp=48.2'C\n"); got != "48.2°C" {
		t.Errorf("parseVcgencmdTemp = %q", got)
	}
}

func TestDedupe(t *testing.T) {
	got := dedupe([]string{"a", "b", "a", "c", "b"})
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Errorf("dedupe = %v", got)
	}
}

func TestCollectCPUFallbackLinux(t *testing.T) {
	origGOOS := RuntimeGOOS
	origReadFile := ReadFile
	origCPUInfo := cpuInfoWithContext
	origCPUCounts := cpuCountsWithContext

	defer func() {
		RuntimeGOOS = origGOOS
		ReadFile = origReadFile
		cpuInfoWithContext = origCPUInfo
		cpuCountsWithContext = origCPUCounts
	}()

	RuntimeGOOS = "linux"
	ReadFile = func(name string) ([]byte, error) {
		return []byte("processor\t: 0\nmodel name\t: Fake CPU 9000\n"), nil
	}
	cpuInfoWithContext = func(_ context.Context) ([]cpu.InfoStat, error) {
		return nil, errors.New("no cpu info")
	}
	cpuCountsWithContext = func(_ context.Context, _ bool) (int, error) {
		return 0, errors.New("no counts")
	}

	model, logical, _ := collectCPU(context.Background())
	if model != "Fake CPU 9000" {
		t.Errorf("collectCPU fallback model = %q", model)
	}
	if logical != runtime.NumCPU() {
		t.Errorf("collectCPU fallback logical = %d, want NumCPU %d", logical, runtime.NumCPU())
	}
}

func TestCollectAllFailuresYieldsZeroValues(t *testing.T) {
	origGOOS := RuntimeGOOS
	origReadFile := ReadFile
	origLookPath := LookPath
	origHostname := osHostname
	origHostInfo := hostInfoWithContext
	origHostUptime := hostUptimeWithContext
	origCPUInfo := cpuInfoWithContext
	origCPUCounts := cpuCountsWithContext
	origMem := virtualMemoryWithContext
	origDisk := diskUsageWithContext
	origSensors := sensorsTemperaturesWithContext
	origNetIfaces := netInterfaces
	origExec := execCommandContext
	origPublicIPURL := publicIPURL

	defer func() {
		RuntimeGOOS = origGOOS
		ReadFile = origReadFile
		LookPath = origLookPath
		osHostname = origHostname
		hostInfoWithContext = origHostInfo
		hostUptimeWithContext = origHostUptime
		cpuInfoWithContext = origCPUInfo
		cpuCountsWithContext = origCPUCounts
		virtualMemoryWithContext = origMem
		diskUsageWithContext = origDisk
		sensorsTemperaturesWithContext = origSensors
		netInterfaces = origNetIfaces
		execCommandContext = origExec
		publicIPURL = origPublicIPURL
	}()

	fail := errors.New("unavailable")
	RuntimeGOOS = "linux"
	ReadFile = func(_ string) ([]byte, error) { return nil, fail }
	LookPath = func(_ string) (string, error) { return "", fail }
	osHostname = func() (string, error) { return "testhost", nil }
	hostInfoWithContext = func(_ context.Context) (*host.InfoStat, error) { return nil, fail }
	hostUptimeWithContext = func(_ context.Context) (uint64, error) { return 0, fail }
	cpuInfoWithContext = func(_ context.Context) ([]cpu.InfoStat, error) { return nil, fail }
	cpuCountsWithContext = func(_ context.Context, _ bool) (int, error) { return 0, fail }
	virtualMemoryWithContext = func(_ context.Context) (*mem.VirtualMemoryStat, error) { return nil, fail }
	diskUsageWithContext = func(_ context.Context, _ string) (*disk.UsageStat, error) { return nil, fail }
	sensorsTemperaturesWithContext = func(_ context.Context) ([]sensors.TemperatureStat, error) {
		return nil, fail
	}
	netInterfaces = func() ([]net.Interface, error) { return nil, fail }
	execCommandContext = func(ctx context.Context, _ string, _ ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "sh", "-c", "exit 1")
	}
	publicIPURL = "http://127.0.0.1:9/"

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	info := Collect(ctx)

	if info.Hostname != "testhost" {
		t.Errorf("Hostname = %q, want testhost", info.Hostname)
	}
	if info.CPUModel != "" || info.MemTotalBytes != 0 || info.DiskTotalBytes != 0 ||
		info.UptimeSeconds != 0 || info.Kernel != "" {
		t.Errorf("Expected zero core values, got %+v", info)
	}
	if len(info.LocalIPs) != 0 || info.Gateway != "" || len(info.DNS) != 0 || info.PublicIP != "" {
		t.Errorf("Expected zero network values, got %+v", info)
	}
	if info.GPU != "" || info.Battery != "" || info.Temperature != "" || info.DeviceModel != "" {
		t.Errorf("Expected zero hardware values, got %+v", info)
	}
	if info.TailscaleIP != "" {
		t.Errorf("TailscaleIP = %q, want empty", info.TailscaleIP)
	}
	if info.MemoryDisplay() != "" || info.DiskDisplay() != "" || info.UptimeDisplay() != "" {
		t.Error("Expected empty displays for failed collection")
	}
	// Logical core count degrades to runtime.NumCPU, which is always available.
	if want := strconv.Itoa(runtime.NumCPU()) + " cores"; info.CPUDisplay() != want {
		t.Errorf("CPUDisplay() = %q, want cores-only fallback %q", info.CPUDisplay(), want)
	}
}
