package setup

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/execx"
)

func TestSetupDotfiles(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	called := false
	origExec := execCommand
	defer func() { execCommand = origExec }()
	execCommand = func(name string, args ...string) *exec.Cmd {
		// setupDotfiles calls the executable with "dotfiles", "install"
		for _, arg := range args {
			if arg == "dotfiles" {
				called = true
			}
		}
		return exec.Command("echo", "mock")
	}

	// We expect this to fail in test because os.Executable() might not be what we expect or we don't handle it
	// But we want to see if execCommand is called
	_ = setupDotfiles(false)

	if !called {
		t.Error("dotfiles install command was not called")
	}
}

func TestSetupDotfiles_RunsSecretsRestoreWhenConfigured(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	tempDir := t.TempDir()
	manifestPath := filepath.Join(tempDir, "bin", "secrets", "server.manifest")
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
		t.Fatalf("failed to create manifest directory: %v", err)
	}
	if err := os.WriteFile(manifestPath, []byte("# test\n"), 0o644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}

	viper.Set("dotfiles.worktree_path", tempDir)
	t.Setenv("BWS_ACCESS_TOKEN", "test-token")

	var calls []string
	execCommand = func(name string, args ...string) *exec.Cmd {
		calls = append(calls, strings.Join(args, " "))
		return exec.Command("echo", "mock")
	}
	defer func() { execCommand = execx.Command }()

	if err := setupDotfiles(false); err != nil {
		t.Fatalf("setupDotfiles returned error: %v", err)
	}

	joined := strings.Join(calls, "\n")
	if !strings.Contains(joined, "dotfiles install") {
		t.Fatalf("expected dotfiles install call, got: %s", joined)
	}
	if !strings.Contains(joined, "dotfiles secrets restore --manifest "+manifestPath) {
		t.Fatalf("expected dotfiles secrets restore call, got: %s", joined)
	}
}

func TestSetupDotfiles_SkipsSecretsRestoreWithoutToken(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	tempDir := t.TempDir()
	manifestPath := filepath.Join(tempDir, "bin", "secrets", "server.manifest")
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
		t.Fatalf("failed to create manifest directory: %v", err)
	}
	if err := os.WriteFile(manifestPath, []byte("# test\n"), 0o644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}

	viper.Set("dotfiles.worktree_path", tempDir)
	t.Setenv("BWS_ACCESS_TOKEN", "")

	var calls []string
	execCommand = func(name string, args ...string) *exec.Cmd {
		calls = append(calls, strings.Join(args, " "))
		return exec.Command("echo", "mock")
	}
	defer func() { execCommand = execx.Command }()

	if err := setupDotfiles(false); err != nil {
		t.Fatalf("setupDotfiles returned error: %v", err)
	}

	joined := strings.Join(calls, "\n")
	if !strings.Contains(joined, "dotfiles install") {
		t.Fatalf("expected dotfiles install call, got: %s", joined)
	}
	if strings.Contains(joined, "dotfiles secrets restore") {
		t.Fatalf("did not expect dotfiles secrets restore call, got: %s", joined)
	}
}
