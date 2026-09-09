package immich

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/eng618/eng/internal/execx"
	"github.com/eng618/eng/internal/paths"
	"github.com/eng618/eng/internal/ui/theme"
)

var execCommandContext = execx.CommandContext

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

// NewManager creates an Immich manager targeting the user container setup.
func NewManager(basePath string) *Manager {
	home := paths.MustHome()
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
	_, err := execx.LookPath("systemctl")
	return err == nil
}

// HasDocker checks if Docker CLI is available.
func (m *Manager) HasDocker() bool {
	_, err := execx.LookPath("docker")
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
