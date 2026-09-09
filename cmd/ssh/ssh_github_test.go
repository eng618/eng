package ssh

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

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
		return exec.Command(
			"echo",
			"Hi user! You've successfully authenticated, but GitHub does not provide shell access.",
		)
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
