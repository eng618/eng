package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"

	internalconfig "github.com/eng618/eng/internal/config"
)

func setupTestViper(t *testing.T) string {
	t.Helper()
	viper.Reset()
	viper.SetConfigType("yaml")
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".eng.yaml")
	viper.SetConfigFile(configPath)

	// Write an empty config file so WriteConfig doesn't error
	err := os.WriteFile(configPath, []byte("{}"), 0o644)
	if err != nil {
		t.Fatalf("failed to write initial test config file: %v", err)
	}
	return configPath
}

// mockAllPrompts stubs all interactive ui prompt wrappers to return default values immediately.
func mockAllPrompts() func() {
	oldConfirm := internalconfig.ConfirmPrompt
	oldInput := internalconfig.InputPrompt
	oldSelect := internalconfig.SelectPrompt
	oldMultiSelect := internalconfig.MultiSelectPrompt

	internalconfig.ConfirmPrompt = func(_ string, defaultVal bool) (bool, error) {
		return defaultVal, nil
	}
	internalconfig.InputPrompt = func(_, defaultVal string) (string, error) {
		return defaultVal, nil
	}
	internalconfig.SelectPrompt = func(_ string, _ []string, defaultVal string) (string, error) {
		return defaultVal, nil
	}
	internalconfig.MultiSelectPrompt = func(_ string, _, defaultSelected []string) ([]string, error) {
		return defaultSelected, nil
	}

	return func() {
		internalconfig.ConfirmPrompt = oldConfirm
		internalconfig.InputPrompt = oldInput
		internalconfig.SelectPrompt = oldSelect
		internalconfig.MultiSelectPrompt = oldMultiSelect
	}
}

// ============================================================================
// T - Tests (Unit and Table-driven Tests)
// ============================================================================

func TestConfigCmd_Subcommands(t *testing.T) {
	subcommands := ConfigCmd.Commands()
	expected := map[string]bool{
		"list":                             false,
		"edit":                             false,
		"email":                            false,
		"dotfiles-repo":                    false,
		"dotfiles-repo-url":                false,
		"dotfiles-branch":                  false,
		"dotfiles-bare-repo-path":          false,
		"dotfiles-target-repo-path [path]": false,
		"git-dev-path":                     false,
		"ide-url [url]":                    false,
		"telemetry":                        false,
		"verbose":                          false,
	}

	for _, cmd := range subcommands {
		if _, exists := expected[cmd.Use]; exists {
			expected[cmd.Use] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("Expected subcommand %s not found in ConfigCmd", name)
		}
	}
}

func TestConfigCmd_RunAll(t *testing.T) {
	setupTestViper(t)

	// Stub all prompts
	restore := mockAllPrompts()
	defer restore()

	// Override specific values returned by mock for this test so we don't fall into infinite loops or default writes
	internalconfig.ConfirmPrompt = func(_ string, _ bool) (bool, error) {
		return true, nil
	}

	ConfigCmd.SetOut(io.Discard)
	ConfigCmd.SetErr(io.Discard)

	// Test run
	ConfigCmd.Run(ConfigCmd, []string{})
}

func TestVerboseCmd(t *testing.T) {
	setupTestViper(t)
	restore := mockAllPrompts()
	defer restore()

	// Test Case 1: User confirms the current setting
	viper.Set("verbose", true)
	internalconfig.ConfirmPrompt = func(msg string, _ bool) (bool, error) {
		return true, nil // User confirms
	}

	VerboseCmd.Run(VerboseCmd, []string{})
	if !viper.GetBool("verbose") {
		t.Error("Expected verbose to remain true")
	}

	// Test Case 2: User denies, changes setting to false
	internalconfig.ConfirmPrompt = func(msg string, _ bool) (bool, error) {
		return false, nil
	}

	VerboseCmd.Run(VerboseCmd, []string{})
	if viper.GetBool("verbose") {
		t.Error("Expected verbose to be updated to false")
	}
}

func TestEmailCmd(t *testing.T) {
	setupTestViper(t)
	restore := mockAllPrompts()
	defer restore()

	// Test Case 1: Confirm existing
	viper.Set("user-email", "old@example.com")
	internalconfig.ConfirmPrompt = func(_ string, _ bool) (bool, error) {
		return true, nil
	}

	EmailCmd.Run(EmailCmd, []string{})
	if viper.GetString("user-email") != "old@example.com" {
		t.Errorf("Expected email to remain old@example.com, got %s", viper.GetString("user-email"))
	}

	// Test Case 2: Reject existing, update to new
	internalconfig.ConfirmPrompt = func(_ string, _ bool) (bool, error) {
		return false, nil
	}
	internalconfig.InputPrompt = func(_, _ string) (string, error) {
		return "new@example.com", nil
	}

	EmailCmd.Run(EmailCmd, []string{})
	if viper.GetString("user-email") != "new@example.com" {
		t.Errorf("Expected email to update to new@example.com, got %s", viper.GetString("user-email"))
	}
}

func TestDotfilesRepoCmd(t *testing.T) {
	setupTestViper(t)
	restore := mockAllPrompts()
	defer restore()

	// Keep bare_repo_path empty to trigger prompt and update
	viper.Set("dotfiles.bare_repo_path", "")
	internalconfig.InputPrompt = func(_, _ string) (string, error) {
		return "/new/path", nil
	}

	DotfilesRepoCmd.Run(DotfilesRepoCmd, []string{})
	if viper.GetString("dotfiles.bare_repo_path") != "/new/path" {
		t.Errorf("expected dotfiles.bare_repo_path to update, got %s", viper.GetString("dotfiles.bare_repo_path"))
	}
}

func TestDotfilesRepoURLCmd(t *testing.T) {
	setupTestViper(t)
	restore := mockAllPrompts()
	defer restore()

	viper.Set("dotfiles.repo_url", "")
	internalconfig.InputPrompt = func(_, _ string) (string, error) {
		return "git@github.com:new/dotfiles.git", nil
	}

	DotfilesRepoURLCmd.Run(DotfilesRepoURLCmd, []string{})
	if viper.GetString("dotfiles.repo_url") != "git@github.com:new/dotfiles.git" {
		t.Errorf("expected repo_url to update, got %s", viper.GetString("dotfiles.repo_url"))
	}
}

func TestDotfilesBranchCmd(t *testing.T) {
	setupTestViper(t)
	restore := mockAllPrompts()
	defer restore()

	viper.Set("dotfiles.branch", "")
	internalconfig.SelectPrompt = func(_ string, _ []string, _ string) (string, error) {
		return "main", nil
	}

	DotfilesBranchCmd.Run(DotfilesBranchCmd, []string{})
	if viper.GetString("dotfiles.branch") != "main" {
		t.Errorf("expected branch to update, got %s", viper.GetString("dotfiles.branch"))
	}
}

func TestDotfilesBareRepoPathCmd(t *testing.T) {
	setupTestViper(t)
	restore := mockAllPrompts()
	defer restore()

	viper.Set("dotfiles.bare_repo_path", "")
	internalconfig.InputPrompt = func(_, _ string) (string, error) {
		return "~/.new-bare", nil
	}

	DotfilesBareRepoPathCmd.Run(DotfilesBareRepoPathCmd, []string{})
	if viper.GetString("dotfiles.bare_repo_path") != "~/.new-bare" {
		t.Errorf("expected bare repo path to update, got %s", viper.GetString("dotfiles.bare_repo_path"))
	}
}

func TestDotfilesTargetRepoPathCmd(t *testing.T) {
	setupTestViper(t)
	restore := mockAllPrompts()
	defer restore()

	// With arg: validates and persists.
	target := t.TempDir()
	if err := os.MkdirAll(filepath.Join(target, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	DotfilesTargetRepoPathCmd.Run(DotfilesTargetRepoPathCmd, []string{target})
	if viper.GetString("dotfiles.target_repo_path") != target {
		t.Errorf("expected target persisted, got %s", viper.GetString("dotfiles.target_repo_path"))
	}

	// With missing arg path: error, nothing persisted.
	viper.Set("dotfiles.target_repo_path", "")
	DotfilesTargetRepoPathCmd.Run(DotfilesTargetRepoPathCmd, []string{filepath.Join(target, "nope")})
	if viper.GetString("dotfiles.target_repo_path") != "" {
		t.Errorf("expected nothing persisted, got %s", viper.GetString("dotfiles.target_repo_path"))
	}

	// Bare with value set: prints, no prompt.
	viper.Set("dotfiles.target_repo_path", target)
	DotfilesTargetRepoPathCmd.Run(DotfilesTargetRepoPathCmd, []string{})
	if viper.GetString("dotfiles.target_repo_path") != target {
		t.Errorf("expected value untouched, got %s", viper.GetString("dotfiles.target_repo_path"))
	}
}

func TestGitDevPathCmd(t *testing.T) {
	setupTestViper(t)
	restore := mockAllPrompts()
	defer restore()

	viper.Set("git.dev_path", "/old/dev")
	// Confirm prompt returns false, triggering update prompt
	internalconfig.ConfirmPrompt = func(_ string, _ bool) (bool, error) {
		return false, nil
	}
	internalconfig.InputPrompt = func(_, _ string) (string, error) {
		return "/new/dev", nil
	}

	GitDevPathCmd.Run(GitDevPathCmd, []string{})
	if viper.GetString("git.dev_path") != "/new/dev" {
		t.Errorf("expected git.dev_path to update, got %s", viper.GetString("git.dev_path"))
	}
}

func TestEditCmd_ConfigFlags(t *testing.T) {
	if EditCmd.Use != "edit" {
		t.Errorf("expected edit command Use to be 'edit', got %q", EditCmd.Use)
	}

	interactive, err := EditCmd.Flags().GetBool("interactive")
	if err != nil {
		t.Fatalf("failed to retrieve interactive flag: %v", err)
	}

	if !interactive {
		t.Error("expected interactive flag to default to true")
	}
}

func TestIdeURLCmd(t *testing.T) {
	setupTestViper(t)
	restore := mockAllPrompts()
	defer restore()

	// Test Case 1: Pass URL directly as argument
	IdeURLCmd.Run(IdeURLCmd, []string{"https://example.com/antigravity-ide.tar.gz"})
	if viper.GetString("antigravity.ide_download_url") != "https://example.com/antigravity-ide.tar.gz" {
		t.Errorf("expected antigravity.ide_download_url to be https://example.com/antigravity-ide.tar.gz, got %s",
			viper.GetString("antigravity.ide_download_url"))
	}

	// Test Case 2: Interactive prompt update
	internalconfig.ConfirmPrompt = func(_ string, _ bool) (bool, error) {
		return false, nil // reject existing
	}
	internalconfig.InputPrompt = func(_, _ string) (string, error) {
		return "https://new.example.com/ide.tar.gz", nil
	}

	IdeURLCmd.Run(IdeURLCmd, []string{})
	if viper.GetString("antigravity.ide_download_url") != "https://new.example.com/ide.tar.gz" {
		t.Errorf("expected antigravity.ide_download_url to update to new url, got %s",
			viper.GetString("antigravity.ide_download_url"))
	}
}

func TestTelemetryCmd(t *testing.T) {
	setupTestViper(t)
	restore := mockAllPrompts()
	defer restore()

	// Test Case: Run telemetry enable and disable
	for _, sub := range TelemetryCmd.Commands() {
		if sub.Name() == "enable" {
			err := sub.RunE(sub, []string{})
			if err != nil {
				t.Fatalf("enable command failed: %v", err)
			}
			if !viper.GetBool("telemetry.enabled") {
				t.Error("expected telemetry.enabled to be true")
			}
		}
		if sub.Name() == "disable" {
			err := sub.RunE(sub, []string{})
			if err != nil {
				t.Fatalf("disable command failed: %v", err)
			}
			if viper.GetBool("telemetry.enabled") {
				t.Error("expected telemetry.enabled to be false")
			}
		}
	}
}

// ============================================================================
// E - Examples (Executable Documentation)
// ============================================================================

func ExampleConfigCmd() {
	// Showing structure configuration of parent command
	fmt.Println("Config Command Use:", ConfigCmd.Use)
	fmt.Println("Subcommand Count:", len(ConfigCmd.Commands()))
	// Output:
	// Config Command Use: config
	// Subcommand Count: 12
}

// ============================================================================
// B - Benchmarks (Performance Profiling)
// ============================================================================

func BenchmarkConfigCmd_Init(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ConfigCmd.Commands()
	}
}

func BenchmarkConfigCmd_SetViper(b *testing.B) {
	// Standard performance check of setting configuration variables
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		viper.Set("verbose", true)
		viper.Set("email", "perf@example.com")
	}
}
