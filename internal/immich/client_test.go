package immich

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eng618/eng/internal/containers"
)

func TestNewManager(t *testing.T) {
	mgr := NewManager("")
	assert.NotNil(t, mgr)
	assert.Contains(t, mgr.BasePath, "immich-app")
	assert.Contains(t, mgr.BackupDir, "immich_backups")
	assert.Contains(t, mgr.PostgresData, "postgres")
}

func TestManager_SafetyGuards(t *testing.T) {
	mgr := NewManager("/non/existent/path")
	assert.False(t, mgr.IsConfigured())

	err := mgr.EnsureHostEnvironment("backup")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")

	restoreErr := mgr.RunRestore(context.Background(), "", true)
	assert.Error(t, restoreErr)
	assert.Contains(t, restoreErr.Error(), "not configured")

	startErr := mgr.Start(context.Background())
	assert.Error(t, startErr)

	stopErr := mgr.Stop(context.Background())
	assert.Error(t, stopErr)

	restartErr := mgr.Restart(context.Background())
	assert.Error(t, restartErr)

	logErr := mgr.Logs(context.Background(), "", false, 10)
	assert.Error(t, logErr)
}

func TestRenderStatus(t *testing.T) {
	status := &StatusResult{
		IsHostHostable: true,
		HostOS:         "linux",
		Service: SystemdUnitStatus{
			Name:          "immich.service",
			ActiveState:   "active",
			SubState:      "running",
			UnitFileState: "enabled",
		},
		Timer: TimerStatus{
			Name:        "immich-backup.timer",
			NextTrigger: "2026-08-20 03:00:00",
			Left:        "2h 30m",
		},
		Containers: []containers.ContainerDetail{
			{
				Name:    "immich_server",
				Service: "immich-server",
				State:   "running",
				Health:  "healthy",
			},
			{
				Name:    "immich_postgres",
				Service: "database",
				State:   "running",
				Health:  "healthy",
			},
		},
		API: APIStatus{
			Reachable: true,
			Response:  "{\"res\":\"pong\"}",
			LatencyMs: 2.34,
			URL:       "http://localhost:2283/api/server/ping",
		},
		Database: DatabaseStats{
			Users:      2,
			Assets:     117902,
			Albums:     250,
			Tables:     67,
			StorageDir: "/home/eng618/.immich/postgres",
		},
		Backup: BackupSummary{
			Destination:  "/home/eng618/media/Recovery/immich_backups",
			LatestDB:     "immich_db_20260819_235655.sql.gz",
			LatestSize:   "663 MB",
			LatestTime:   "2026-08-20 00:05:43",
			TotalBackups: 3,
		},
	}

	rendered := RenderStatus(status, 100)
	assert.NotEmpty(t, rendered)
	assert.Contains(t, rendered, "Immich Photo Stack Health")
	assert.Contains(t, rendered, "immich.service")
	assert.Contains(t, rendered, "117,902")
	assert.Contains(t, rendered, "663 MB")
	assert.Contains(t, rendered, "immich_server")
}

func TestRenderStatus_NonHostable(t *testing.T) {
	status := &StatusResult{
		IsHostHostable: false,
		HostOS:         "darwin",
		Service: SystemdUnitStatus{
			Name:        "immich.service",
			ActiveState: "n/a",
		},
		Timer: TimerStatus{
			Name: "immich-backup.timer",
			Left: "n/a (darwin)",
		},
		API: APIStatus{
			Reachable: false,
			URL:       "http://localhost:2283/api/server/ping",
		},
	}

	rendered := RenderStatus(status, 100)
	assert.NotEmpty(t, rendered)
	assert.Contains(t, rendered, "not configured on this host")
}

func TestFormatBadge(t *testing.T) {
	okBadge := formatBadge("running", true)
	assert.Contains(t, okBadge, "RUNNING")

	errBadge := formatBadge("failed", false)
	assert.Contains(t, errBadge, "FAILED")
}

func TestManager_GetStatus_Mock(t *testing.T) {
	mgr := NewManager("/tmp/mock-immich")
	res, err := mgr.GetStatus(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.False(t, res.IsHostHostable)
}
