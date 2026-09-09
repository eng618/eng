package ssh

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eng618/eng/internal/ui"
)

func TestSetupSSH_ExistingValidKey(t *testing.T) {
	tempDir := t.TempDir()
	homeOrig := os.Getenv("HOME")
	_ = os.Setenv("HOME", tempDir)
	origExec := execCommand
	defer func() {
		_ = os.Setenv("HOME", homeOrig)
		execCommand = origExec
	}()

	sshDir := filepath.Join(tempDir, ".ssh")
	_ = os.MkdirAll(sshDir, 0o700)
	keyPath := filepath.Join(sshDir, "github")
	_ = os.WriteFile(keyPath, []byte("mock-key"), 0o600)

	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("echo", "Hi user! You've successfully authenticated")
	}

	if err := SetupSSH(false); err != nil {
		t.Fatalf("SetupSSH failed with existing valid key: %v", err)
	}
}

func TestSetupSSH_BackupChoice(t *testing.T) {
	tempDir := t.TempDir()
	homeOrig := os.Getenv("HOME")
	_ = os.Setenv("HOME", tempDir)
	origExec := execCommand
	origUISelect := ui.Select
	defer func() {
		_ = os.Setenv("HOME", homeOrig)
		execCommand = origExec
		ui.Select = origUISelect
	}()

	backupDir := filepath.Join(tempDir, "Downloads", "ssh")
	_ = os.MkdirAll(backupDir, 0o755)
	_ = os.WriteFile(filepath.Join(backupDir, "github"), []byte("-----BEGIN OPENSSH PRIVATE KEY-----\nkey\n-----END OPENSSH PRIVATE KEY-----"), 0o600)
	_ = os.WriteFile(filepath.Join(backupDir, "github.pub"), []byte("ssh-ed25519 AAAAB3NzaC1lZDI1NTE5AAAAI mock"), 0o644)

	ui.Select = func(prompt string, options []string, def string) (string, error) {
		if strings.Contains(prompt, "How would you like to set up") {
			return options[0], nil // backup
		}
		if strings.Contains(prompt, "Select SSH backup directory") {
			return backupDir, nil
		}
		return options[0], nil
	}

	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("echo", "Hi user! You've successfully authenticated")
	}

	if err := SetupSSH(false); err != nil {
		t.Fatalf("SetupSSH with backup option failed: %v", err)
	}

	// Verify key was migrated
	destKey := filepath.Join(tempDir, ".ssh", "github")
	if _, err := os.Stat(destKey); err != nil {
		t.Errorf("migrated key not found at %s: %v", destKey, err)
	}
}

func TestSetupSSH_SkipChoice(t *testing.T) {
	tempDir := t.TempDir()
	homeOrig := os.Getenv("HOME")
	_ = os.Setenv("HOME", tempDir)
	origUISelect := ui.Select
	defer func() {
		_ = os.Setenv("HOME", homeOrig)
		ui.Select = origUISelect
	}()

	ui.Select = func(prompt string, options []string, def string) (string, error) {
		return "skip - Skip SSH setup", nil
	}

	if err := SetupSSH(false); err != nil {
		t.Fatalf("SetupSSH with skip failed: %v", err)
	}
}
