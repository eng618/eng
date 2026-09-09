package immich

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui/theme"
)

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
	cmd.Stdout = log.Out
	cmd.Stderr = log.Err
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
		cmd.Stdout = log.Out
		cmd.Stderr = log.Err
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
	cmd.Stdout = log.Out
	cmd.Stderr = log.Err
	return cmd.Run()
}
