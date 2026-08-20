package immich

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/dustin/go-humanize"

	"github.com/eng618/eng/internal/containers"
	"github.com/eng618/eng/internal/ui/theme"
)

var execCommandContext = exec.CommandContext

// Manager handles Immich stack operations and health inspection.
type Manager struct {
	BasePath      string
	BackupDir     string
	PostgresData  string
	ComposeFile   string
	ServiceUnit   string
	BackupTimer   string
	BackupScript  string
	RestoreScript string
	HTTPClient    *http.Client
}

// SystemdUnitStatus represents systemd unit health.
type SystemdUnitStatus struct {
	Name          string `json:"name"`
	ActiveState   string `json:"activeState"`
	SubState      string `json:"subState"`
	UnitFileState string `json:"unitFileState"`
	Description   string `json:"description"`
}

// TimerStatus represents systemd timer scheduling.
type TimerStatus struct {
	Name        string `json:"name"`
	NextTrigger string `json:"nextTrigger"`
	LastTrigger string `json:"lastTrigger"`
	Passed      string `json:"passed"`
	Left        string `json:"left"`
}

// APIStatus represents the REST API ping health.
type APIStatus struct {
	Reachable bool    `json:"reachable"`
	Response  string  `json:"response"`
	LatencyMs float64 `json:"latencyMs"`
	URL       string  `json:"url"`
}

// DatabaseStats represents record statistics from PostgreSQL.
type DatabaseStats struct {
	Users      int    `json:"users"`
	Assets     int    `json:"assets"`
	Albums     int    `json:"albums"`
	Tables     int    `json:"tables"`
	StorageDir string `json:"storageDir"`
}

// BackupSummary represents information about local and NAS backups.
type BackupSummary struct {
	Destination  string `json:"destination"`
	LatestDB     string `json:"latestDb"`
	LatestSize   string `json:"latestSize"`
	LatestTime   string `json:"latestTime"`
	TotalBackups int    `json:"totalBackups"`
	LatestConfig string `json:"latestConfig"`
}

// StatusResult aggregates full system state.
type StatusResult struct {
	IsHostHostable bool                         `json:"isHostHostable"`
	HostOS         string                       `json:"hostOs"`
	Service        SystemdUnitStatus            `json:"service"`
	Timer          TimerStatus                  `json:"timer"`
	Containers     []containers.ContainerDetail `json:"containers"`
	API            APIStatus                    `json:"api"`
	Database       DatabaseStats                `json:"database"`
	Backup         BackupSummary                `json:"backup"`
}

// BackupResult holds results from executing backup-immich.sh.
type BackupResult struct {
	BackupFile    string        `json:"backupFile"`
	Size          string        `json:"size"`
	ChecksumFile  string        `json:"checksumFile"`
	ConfigArchive string        `json:"configArchive"`
	Duration      time.Duration `json:"duration"`
	Output        string        `json:"output"`
}

// NewManager creates an Immich manager targeting the user container setup.
func NewManager(basePath string) *Manager {
	home, _ := os.UserHomeDir()
	if basePath == "" {
		basePath = filepath.Join(home, "bin", "containers", "immich-app")
	}

	return &Manager{
		BasePath:      basePath,
		BackupDir:     filepath.Join(home, "media", "Recovery", "immich_backups"),
		PostgresData:  filepath.Join(home, ".immich", "postgres"),
		ComposeFile:   filepath.Join(basePath, "docker-compose.yml"),
		ServiceUnit:   "immich.service",
		BackupTimer:   "immich-backup.timer",
		BackupScript:  filepath.Join(basePath, "backup-immich.sh"),
		RestoreScript: filepath.Join(basePath, "restore-immich.sh"),
		HTTPClient:    &http.Client{Timeout: 3 * time.Second},
	}
}

// IsConfigured checks if Immich configuration files exist on this machine.
func (m *Manager) IsConfigured() bool {
	if _, err := os.Stat(m.ComposeFile); err == nil {
		return true
	}
	return false
}

// HasSystemd checks if systemctl user session is available.
func (m *Manager) HasSystemd() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	_, err := exec.LookPath("systemctl")
	return err == nil
}

// HasDocker checks if Docker CLI is available.
func (m *Manager) HasDocker() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

// EnsureHostEnvironment performs pre-flight checks before running host-level actions.
func (m *Manager) EnsureHostEnvironment(action string) error {
	if !m.IsConfigured() {
		return theme.NewActionableError(
			fmt.Errorf("immich stack is not configured on this machine (%s was not found)", m.ComposeFile),
			"The '"+action+"' command must be run directly on the host server running Immich, or verify that your containers directory is at ~/bin/containers/immich-app.",
		)
	}

	if !m.HasDocker() {
		return theme.NewActionableError(
			errors.New("docker command-line tool is not installed or not in PATH"),
			"Install Docker to manage Immich containers on this host.",
		)
	}

	return nil
}

// GetStatus probes systemd, docker, API, database, and backups gracefully across all machines.
func (m *Manager) GetStatus(ctx context.Context) (*StatusResult, error) {
	res := &StatusResult{
		HostOS:         runtime.GOOS,
		IsHostHostable: m.IsConfigured(),
		Database:       DatabaseStats{StorageDir: m.PostgresData},
		Backup:         BackupSummary{Destination: m.BackupDir},
	}

	// 1. Systemd Service Status (if on Linux with systemd)
	if m.HasSystemd() {
		res.Service = m.checkSystemdService(ctx)
		res.Timer = m.checkSystemdTimer(ctx)
	} else {
		res.Service = SystemdUnitStatus{
			Name:        m.ServiceUnit,
			ActiveState: "n/a",
			SubState:    runtime.GOOS + " (no systemd)",
			Description: "Systemd is not available on " + runtime.GOOS,
		}
		res.Timer = TimerStatus{
			Name: m.BackupTimer,
			Left: "n/a (" + runtime.GOOS + ")",
		}
	}

	// 2. Docker Containers Inspection (if configured and Docker is present)
	if m.IsConfigured() && m.HasDocker() {
		res.Containers = m.checkContainers(ctx)
		res.Database = m.checkDatabase(ctx)
		res.Backup = m.checkBackups(ctx)
	}

	// 3. API Ping (attempts to probe local or network API endpoint)
	res.API = m.checkAPI(ctx)

	return res, nil
}

func (m *Manager) checkSystemdService(ctx context.Context) SystemdUnitStatus {
	cmd := execCommandContext(ctx, "systemctl", "--user", "show", m.ServiceUnit,
		"--property=Id,ActiveState,SubState,UnitFileState,Description")
	out, err := cmd.Output()
	if err != nil {
		return SystemdUnitStatus{Name: m.ServiceUnit, ActiveState: "inactive", SubState: "unknown"}
	}

	status := SystemdUnitStatus{Name: m.ServiceUnit}
	lines := strings.Split(string(out), "\n")
	for _, l := range lines {
		parts := strings.SplitN(l, "=", 2)
		if len(parts) == 2 {
			switch parts[0] {
			case "ActiveState":
				status.ActiveState = parts[1]
			case "SubState":
				status.SubState = parts[1]
			case "UnitFileState":
				status.UnitFileState = parts[1]
			case "Description":
				status.Description = parts[1]
			}
		}
	}
	return status
}

func (m *Manager) checkSystemdTimer(ctx context.Context) TimerStatus {
	timer := TimerStatus{Name: m.BackupTimer}
	cmd := execCommandContext(ctx, "systemctl", "--user", "list-timers", "--all")
	out, err := cmd.Output()
	if err != nil {
		return timer
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, m.BackupTimer) {
			fields := strings.Fields(line)
			if len(fields) >= 6 {
				timer.NextTrigger = fields[0] + " " + fields[1] + " " + fields[2]
				timer.Left = fields[3]
			}
			break
		}
	}
	return timer
}

func (m *Manager) checkContainers(ctx context.Context) []containers.ContainerDetail {
	cmd := execCommandContext(ctx, "docker", "compose", "-f", m.ComposeFile, "ps", "--format", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var result []containers.ContainerDetail
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var c containers.ContainerDetail
		if err := json.Unmarshal([]byte(line), &c); err == nil {
			result = append(result, c)
		}
	}
	return result
}

func (m *Manager) checkAPI(ctx context.Context) APIStatus {
	url := "http://localhost:2283/api/server/ping"
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return APIStatus{URL: url, Reachable: false}
	}

	resp, err := m.HTTPClient.Do(req)
	latency := float64(time.Since(start).Microseconds()) / 1000.0
	if err != nil {
		return APIStatus{URL: url, Reachable: false, LatencyMs: latency}
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(resp.Body)
	return APIStatus{
		Reachable: resp.StatusCode == http.StatusOK,
		Response:  strings.TrimSpace(buf.String()),
		LatencyMs: latency,
		URL:       url,
	}
}

func (m *Manager) checkDatabase(ctx context.Context) DatabaseStats {
	stats := DatabaseStats{StorageDir: m.PostgresData}

	// Query database counts via container
	query := `SELECT (SELECT count(*) FROM "user"), (SELECT count(*) FROM "asset"), (SELECT count(*) FROM "album");`
	cmd := execCommandContext(
		ctx,
		"docker",
		"exec",
		"immich_postgres",
		"psql",
		"-U",
		"postgres",
		"-d",
		"immich",
		"-t",
		"-c",
		query,
	)
	out, err := cmd.Output()
	if err == nil {
		parts := strings.Split(strings.TrimSpace(string(out)), "|")
		if len(parts) == 3 {
			stats.Users, _ = strconv.Atoi(strings.TrimSpace(parts[0]))
			stats.Assets, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
			stats.Albums, _ = strconv.Atoi(strings.TrimSpace(parts[2]))
		}
	}

	tableQuery := `SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public';`
	tableCmd := execCommandContext(
		ctx,
		"docker",
		"exec",
		"immich_postgres",
		"psql",
		"-U",
		"postgres",
		"-d",
		"immich",
		"-t",
		"-c",
		tableQuery,
	)
	tOut, tErr := tableCmd.Output()
	if tErr == nil {
		stats.Tables, _ = strconv.Atoi(strings.TrimSpace(string(tOut)))
	}

	return stats
}

func (m *Manager) checkBackups(_ context.Context) BackupSummary {
	summary := BackupSummary{Destination: m.BackupDir}
	dbDir := filepath.Join(m.BackupDir, "db")

	entries, err := os.ReadDir(dbDir)
	if err != nil {
		return summary
	}

	var dumpFiles []os.FileInfo
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "immich_db_") && strings.HasSuffix(e.Name(), ".sql.gz") {
			if info, err := e.Info(); err == nil {
				dumpFiles = append(dumpFiles, info)
			}
		}
	}

	summary.TotalBackups = len(dumpFiles)
	if len(dumpFiles) > 0 {
		latest := dumpFiles[0]
		for _, f := range dumpFiles[1:] {
			if f.ModTime().After(latest.ModTime()) {
				latest = f
			}
		}
		summary.LatestDB = latest.Name()
		summary.LatestSize = humanize.Bytes(uint64(latest.Size()))
		summary.LatestTime = latest.ModTime().Format("2006-01-02 15:04:05")
	}

	metaDir := filepath.Join(m.BackupDir, "meta")
	if metaEntries, err := os.ReadDir(metaDir); err == nil {
		for _, e := range metaEntries {
			if strings.HasPrefix(e.Name(), "immich_config_") && strings.HasSuffix(e.Name(), ".tar.gz") {
				summary.LatestConfig = e.Name()
			}
		}
	}

	return summary
}

// RunBackup executes backup-immich.sh with environment validation.
func (m *Manager) RunBackup(ctx context.Context, retention int) (*BackupResult, error) {
	if err := m.EnsureHostEnvironment("backup"); err != nil {
		return nil, err
	}

	if _, err := os.Stat(m.BackupScript); os.IsNotExist(err) {
		return nil, theme.NewActionableError(
			fmt.Errorf("backup script not found at %s", m.BackupScript),
			"Ensure the Immich backup script is deployed on this host.",
		)
	}

	start := time.Now()
	args := []string{}
	cmd := execCommandContext(ctx, m.BackupScript, args...)
	if retention > 0 {
		cmd.Env = append(os.Environ(), fmt.Sprintf("RETENTION_DAYS=%d", retention))
	}

	out, err := cmd.CombinedOutput()
	duration := time.Since(start)
	if err != nil {
		return nil, fmt.Errorf("backup failed: %w\nOutput: %s", err, string(out))
	}

	sum := m.checkBackups(ctx)

	return &BackupResult{
		BackupFile:    filepath.Join(m.BackupDir, "db", sum.LatestDB),
		Size:          sum.LatestSize,
		ChecksumFile:  filepath.Join(m.BackupDir, "db", sum.LatestDB+".sha256"),
		ConfigArchive: filepath.Join(m.BackupDir, "meta", sum.LatestConfig),
		Duration:      duration,
		Output:        string(out),
	}, nil
}

// RunRestore executes restore-immich.sh with environment validation.
func (m *Manager) RunRestore(ctx context.Context, backupFile string, autoConfirm bool) error {
	if err := m.EnsureHostEnvironment("restore"); err != nil {
		return err
	}

	if _, err := os.Stat(m.RestoreScript); os.IsNotExist(err) {
		return theme.NewActionableError(
			fmt.Errorf("restore script not found at %s", m.RestoreScript),
			"Ensure the Immich restore script is deployed on this host.",
		)
	}

	var args []string
	if autoConfirm {
		args = append(args, "--yes")
	}
	if backupFile != "" {
		args = append(args, "--file", backupFile)
	}

	cmd := execCommandContext(ctx, m.RestoreScript, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Start brings up the service via systemd or docker compose fallback.
func (m *Manager) Start(ctx context.Context) error {
	if err := m.EnsureHostEnvironment("start"); err != nil {
		return err
	}

	if m.HasSystemd() {
		cmd := execCommandContext(ctx, "systemctl", "--user", "start", m.ServiceUnit)
		if out, err := cmd.CombinedOutput(); err == nil {
			return nil
		} else {
			_ = out
		}
	}

	// Fallback to docker compose up
	composeCmd := execCommandContext(ctx, "docker", "compose", "-f", m.ComposeFile, "up", "-d")
	if cOut, cErr := composeCmd.CombinedOutput(); cErr != nil {
		return fmt.Errorf("failed to start immich via docker compose: %s", string(cOut))
	}
	return nil
}

// Stop stops the service gracefully via systemd or docker compose fallback.
func (m *Manager) Stop(ctx context.Context) error {
	if err := m.EnsureHostEnvironment("stop"); err != nil {
		return err
	}

	if m.HasSystemd() {
		cmd := execCommandContext(ctx, "systemctl", "--user", "stop", m.ServiceUnit)
		if out, err := cmd.CombinedOutput(); err == nil {
			return nil
		} else {
			_ = out
		}
	}

	// Fallback to docker compose stop
	composeCmd := execCommandContext(ctx, "docker", "compose", "-f", m.ComposeFile, "stop", "-t", "60")
	if cOut, cErr := composeCmd.CombinedOutput(); cErr != nil {
		return fmt.Errorf("failed to stop immich via docker compose: %s", string(cOut))
	}
	return nil
}

// Restart restarts the service via systemd or docker compose fallback.
func (m *Manager) Restart(ctx context.Context) error {
	if err := m.EnsureHostEnvironment("restart"); err != nil {
		return err
	}

	if m.HasSystemd() {
		cmd := execCommandContext(ctx, "systemctl", "--user", "restart", m.ServiceUnit)
		if out, err := cmd.CombinedOutput(); err == nil {
			return nil
		} else {
			_ = out
		}
	}

	if err := m.Stop(ctx); err != nil {
		return err
	}
	return m.Start(ctx)
}

// Logs streams logs from journalctl or docker compose.
func (m *Manager) Logs(ctx context.Context, service string, follow bool, tail int) error {
	if err := m.EnsureHostEnvironment("logs"); err != nil {
		return err
	}

	if service == "" && m.HasSystemd() {
		args := []string{"--user", "-u", m.ServiceUnit}
		if follow {
			args = append(args, "-f")
		}
		if tail > 0 {
			args = append(args, "-n", strconv.Itoa(tail))
		}
		cmd := execCommandContext(ctx, "journalctl", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// Service-specific logs via Docker Compose
	args := []string{"compose", "-f", m.ComposeFile, "logs"}
	if follow {
		args = append(args, "-f")
	}
	if tail > 0 {
		args = append(args, "--tail", strconv.Itoa(tail))
	}
	if service != "" {
		args = append(args, service)
	}

	cmd := execCommandContext(ctx, "docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RenderStatus renders a comprehensive formatted status card.
func RenderStatus(s *StatusResult, termWidth int) string {
	if termWidth <= 0 {
		termWidth = 100
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Background).
		Background(theme.Primary).
		Padding(0, 1)

	cellStyle := lipgloss.NewStyle().Padding(0, 1)
	borderStyle := lipgloss.NewStyle().Foreground(theme.Primary)

	var sb strings.Builder

	// Title
	title := theme.PrimaryText.Bold(true).Render("📸 Immich Photo Stack Health & Metrics")
	sb.WriteString(title)
	sb.WriteString("\n\n")

	// Host check note if not configured locally
	if !s.IsHostHostable {
		notice := theme.WarningBanner.Render("NOTE") + " " +
			theme.MutedText.Render(
				fmt.Sprintf(
					"Immich stack is not configured on this host (%s). Probing remote/network API only.",
					s.HostOS,
				),
			)
		sb.WriteString(notice)
		sb.WriteString("\n\n")
	}

	// Overview Table
	tOverview := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(borderStyle).
		Headers("COMPONENT", "STATE / VALUE", "DETAILS").
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return headerStyle
			}
			return cellStyle
		})

	// Service state badge
	svcBadge := formatBadge(s.Service.ActiveState, s.Service.ActiveState == "active")
	timerBadge := theme.MutedText.Render(s.Timer.Left)
	if s.Timer.NextTrigger != "" {
		timerBadge = fmt.Sprintf("Next: %s (%s)", s.Timer.NextTrigger, s.Timer.Left)
	}

	apiBadge := formatBadge("ONLINE", s.API.Reachable)
	if !s.API.Reachable {
		apiBadge = formatBadge("OFFLINE", false)
	}
	apiDetails := fmt.Sprintf("%.2fms latency (%s)", s.API.LatencyMs, s.API.URL)

	dbDetails := fmt.Sprintf("Users: %d | Assets: %s | Albums: %d (Storage: %s)",
		s.Database.Users, humanize.Comma(int64(s.Database.Assets)), s.Database.Albums, s.Database.StorageDir)
	if !s.IsHostHostable {
		dbDetails = "Local database storage not present on this host"
	}

	backupDetails := fmt.Sprintf("Latest: %s (%s) | Total Snapshots: %d",
		s.Backup.LatestDB, s.Backup.LatestSize, s.Backup.TotalBackups)
	if s.Backup.LatestDB == "" {
		backupDetails = "No database backups found on local storage"
	}

	tOverview.Row("Systemd Service", svcBadge, s.Service.Name+" ("+s.Service.UnitFileState+")")
	tOverview.Row("Backup Timer", s.Timer.Name, timerBadge)
	tOverview.Row("Immich API", apiBadge, apiDetails)
	tOverview.Row("Postgres Database", fmt.Sprintf("%d Tables", s.Database.Tables), dbDetails)
	tOverview.Row("Backup Storage", s.Backup.Destination, backupDetails)

	sb.WriteString(tOverview.Render())
	sb.WriteString("\n\n")

	// Containers Table
	if len(s.Containers) > 0 {
		cTable := table.New().
			Border(lipgloss.RoundedBorder()).
			BorderStyle(borderStyle).
			Headers("CONTAINER", "SERVICE", "STATUS", "HEALTH").
			StyleFunc(func(row, col int) lipgloss.Style {
				if row == 0 {
					return headerStyle
				}
				return cellStyle
			})

		for _, c := range s.Containers {
			cTable.Row(
				theme.BoldText.Render(c.Name),
				c.Service,
				formatBadge(c.State, c.State == "running"),
				formatBadge(c.Health, c.Health == "healthy" || c.Health == ""),
			)
		}
		sb.WriteString(theme.PrimaryText.Bold(true).Render("Containers:"))
		sb.WriteString("\n")
		sb.WriteString(cTable.Render())
		sb.WriteString("\n")
	}

	return sb.String()
}

func formatBadge(text string, isOk bool) string {
	badge := lipgloss.NewStyle().Bold(true).Padding(0, 1)
	if isOk {
		return badge.Background(lipgloss.Color("#10B981")).
			Foreground(lipgloss.Color("#000000")).
			Render(" " + strings.ToUpper(text) + " ")
	}
	return badge.Background(lipgloss.Color("#EF4444")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Render(" " + strings.ToUpper(text) + " ")
}
