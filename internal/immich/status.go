package immich

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"

	"github.com/eng618/eng/internal/containers"
)

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
