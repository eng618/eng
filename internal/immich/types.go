package immich

import (
	"time"

	"github.com/eng618/eng/internal/containers"
)

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
