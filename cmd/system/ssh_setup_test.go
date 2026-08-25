package system

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eng618/eng/internal/ui"
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
		t.Errorf("expected Host host2 once, got count %d in:\n%s", strings.Count(string(data), "Host host2"), string(data))
	}
}

func TestFindGitHubSSHKey(t *testing.T) {
	tempDir := t.TempDir()
	sshDir := filepath.Join(tempDir, ".ssh")
	_ = os.MkdirAll(sshDir, 0o700)

	// When no keys exist, defaults to github path
	defaultKey := FindGitHubSSHKey(sshDir)
	if defaultKey != filepath.Join(sshDir, "github") {
		t.Errorf("expected default key %s, got %s", filepath.Join(sshDir, "github"), defaultKey)
	}

	// When github key exists, returns github key
	_ = os.WriteFile(filepath.Join(sshDir, "github"), []byte("mock-key"), 0o600)
	if key := FindGitHubSSHKey(sshDir); key != filepath.Join(sshDir, "github") {
		t.Errorf("expected %s, got %s", filepath.Join(sshDir, "github"), key)
	}
	_ = os.Remove(filepath.Join(sshDir, "github"))

	// When config points to a custom identity file
	customKeyPath := filepath.Join(sshDir, "custom_gh_key")
	_ = os.WriteFile(customKeyPath, []byte("mock-custom"), 0o600)
	cfgContent := "Host github.com\n    IdentityFile " + customKeyPath + "\n"
	_ = os.WriteFile(filepath.Join(sshDir, "config"), []byte(cfgContent), 0o600)
	if key := FindGitHubSSHKey(sshDir); key != customKeyPath {
		t.Errorf("expected custom key from config %s, got %s", customKeyPath, key)
	}
	_ = os.Remove(filepath.Join(sshDir, "config"))

	// When id_ed25519 exists
	_ = os.WriteFile(filepath.Join(sshDir, "id_ed25519"), []byte("mock-ed25519"), 0o600)
	if key := FindGitHubSSHKey(sshDir); key != filepath.Join(sshDir, "id_ed25519") {
		t.Errorf("expected %s, got %s", filepath.Join(sshDir, "id_ed25519"), key)
	}
	_ = os.Remove(filepath.Join(sshDir, "id_ed25519"))

	// When id_rsa exists
	_ = os.WriteFile(filepath.Join(sshDir, "id_rsa"), []byte("mock-rsa"), 0o600)
	if key := FindGitHubSSHKey(sshDir); key != filepath.Join(sshDir, "id_rsa") {
		t.Errorf("expected %s, got %s", filepath.Join(sshDir, "id_rsa"), key)
	}
}

func TestEnsureSSHConfig(t *testing.T) {
	tempDir := t.TempDir()
	keyPath := filepath.Join(tempDir, ".ssh", "github")
	_ = os.MkdirAll(filepath.Dir(keyPath), 0o700)

	if err := EnsureSSHConfig(keyPath); err != nil {
		t.Fatalf("EnsureSSHConfig failed: %v", err)
	}

	cfgPath := filepath.Join(tempDir, ".ssh", "config")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("could not read generated config: %v", err)
	}

	if !strings.Contains(string(data), "Host github.com") || !strings.Contains(string(data), keyPath) {
		t.Errorf("config missing expected entries, got:\n%s", string(data))
	}

	// Calling again should not duplicate
	if err := EnsureSSHConfig(keyPath); err != nil {
		t.Fatalf("second call to EnsureSSHConfig failed: %v", err)
	}
	data, _ = os.ReadFile(cfgPath)
	if strings.Count(string(data), "Host github.com") != 1 {
		t.Errorf("expected Host github.com once, got:\n%s", string(data))
	}
}

func TestValidateGitHubSSHAuth(t *testing.T) {
	origExec := execCommand
	defer func() { execCommand = origExec }()

	// 1. Success case
	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("echo", "Hi user! You've successfully authenticated, but GitHub does not provide shell access.")
	}

	if err := ValidateGitHubSSHAuth("/fake/key", false); err != nil {
		t.Errorf("expected validation success, got error: %v", err)
	}

	// 2. Failure case
	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("sh", "-c", "echo 'Permission denied (publickey).' && exit 1")
	}

	if err := ValidateGitHubSSHAuth("/fake/key", false); err == nil {
		t.Errorf("expected validation failure, got nil error")
	}
}

func TestGenerateSSHKey(t *testing.T) {
	tempDir := t.TempDir()
	sshDir := filepath.Join(tempDir, ".ssh")
	keyPath := filepath.Join(sshDir, "github")

	origExec := execCommand
	origLookPath := lookPath
	defer func() {
		execCommand = origExec
		lookPath = origLookPath
	}()

	lookPath = func(path string) (string, error) {
		if path == "gh" {
			return "/usr/bin/gh", nil
		}
		return "", errors.New("not found")
	}

	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "ssh-keygen" {
			// Create dummy key files
			_ = os.WriteFile(keyPath, []byte("dummy-private-key"), 0o600)
			_ = os.WriteFile(keyPath+".pub", []byte("dummy-public-key"), 0o644)
			return exec.Command("echo", "mock keygen")
		}
		if name == "gh" {
			return exec.Command("echo", "mock gh")
		}
		return exec.Command("echo", "mock")
	}

	autoRegistered, err := GenerateSSHKey(keyPath, false)
	if err != nil {
		t.Fatalf("GenerateSSHKey returned error: %v", err)
	}
	if !autoRegistered {
		t.Errorf("expected key to be automatically registered with gh mock")
	}

	if _, err := os.Stat(keyPath); err != nil {
		t.Errorf("expected private key file at %s", keyPath)
	}
}

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
