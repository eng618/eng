package ssh

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrateSSHFromBackup(t *testing.T) {
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "backup_ssh")
	destDir := filepath.Join(tempDir, ".ssh")

	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}

	// Create test files in backup directory
	privKeyContent := "-----BEGIN OPENSSH PRIVATE KEY-----\ntest-key-data\n-----END OPENSSH PRIVATE KEY-----\n"
	pubKeyContent := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGithubKey test@example.com\n"
	cfgContent := "Host myserver\n    HostName 1.2.3.4\n    User admin\n"
	knownHostsContent := "github.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl\n"

	if err := os.WriteFile(filepath.Join(sourceDir, "github"), []byte(privKeyContent), 0o600); err != nil {
		t.Fatalf("failed to write priv key: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "github.pub"), []byte(pubKeyContent), 0o644); err != nil {
		t.Fatalf("failed to write pub key: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "config"), []byte(cfgContent), 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "known_hosts"), []byte(knownHostsContent), 0o600); err != nil {
		t.Fatalf("failed to write known_hosts: %v", err)
	}
	// Write a non-SSH file that should be skipped
	if err := os.WriteFile(filepath.Join(sourceDir, "notes.txt"), []byte("not an ssh key"), 0o644); err != nil {
		t.Fatalf("failed to write notes.txt: %v", err)
	}

	if err := MigrateSSHFromBackup(sourceDir, destDir, true); err != nil {
		t.Fatalf("MigrateSSHFromBackup returned error: %v", err)
	}

	// Verify dest directory mode
	destInfo, err := os.Stat(destDir)
	if err != nil {
		t.Fatalf("dest dir does not exist: %v", err)
	}
	if destInfo.Mode().Perm() != 0o700 {
		t.Errorf("expected dest dir mode 0700, got %04o", destInfo.Mode().Perm())
	}

	// Verify private key
	privDestPath := filepath.Join(destDir, "github")
	privInfo, err := os.Stat(privDestPath)
	if err != nil {
		t.Fatalf("migrated private key not found: %v", err)
	}
	if privInfo.Mode().Perm() != 0o600 {
		t.Errorf("expected private key mode 0600, got %04o", privInfo.Mode().Perm())
	}

	// Verify public key
	pubDestPath := filepath.Join(destDir, "github.pub")
	pubInfo, err := os.Stat(pubDestPath)
	if err != nil {
		t.Fatalf("migrated public key not found: %v", err)
	}
	if pubInfo.Mode().Perm() != 0o644 {
		t.Errorf("expected public key mode 0644, got %04o", pubInfo.Mode().Perm())
	}

	// Verify config
	cfgDestPath := filepath.Join(destDir, "config")
	cfgInfo, err := os.Stat(cfgDestPath)
	if err != nil {
		t.Fatalf("migrated config not found: %v", err)
	}
	if cfgInfo.Mode().Perm() != 0o600 {
		t.Errorf("expected config mode 0600, got %04o", cfgInfo.Mode().Perm())
	}

	// Verify notes.txt was skipped
	if _, err := os.Stat(filepath.Join(destDir, "notes.txt")); err == nil {
		t.Errorf("notes.txt should not have been copied to ~/.ssh")
	}
}

func TestMergeOrWriteSSHConfig(t *testing.T) {
	tempDir := t.TempDir()
	cfgPath := filepath.Join(tempDir, "config")

	// 1. Initial write
	block1 := "Host host1\n    HostName 10.0.0.1\n"
	if err := mergeOrWriteSSHConfig(cfgPath, block1); err != nil {
		t.Fatalf("initial write failed: %v", err)
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if !strings.Contains(string(data), "Host host1") {
		t.Errorf("expected Host host1 in config")
	}

	// 2. Merging a new block
	block2 := "Host host2\n    HostName 10.0.0.2\n"
	if err := mergeOrWriteSSHConfig(cfgPath, block2); err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	data, err = os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if !strings.Contains(string(data), "Host host1") || !strings.Contains(string(data), "Host host2") {
		t.Errorf("expected both hosts in config, got:\n%s", string(data))
	}

	// 3. Duplicate write should not duplicate
	if err := mergeOrWriteSSHConfig(cfgPath, block2); err != nil {
		t.Fatalf("duplicate write failed: %v", err)
	}
	data, _ = os.ReadFile(cfgPath)
	if strings.Count(string(data), "Host host2") != 1 {
		t.Errorf(
			"expected Host host2 once, got count %d in:\n%s",
			strings.Count(string(data), "Host host2"),
			string(data),
		)
	}
}
