package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/ui"
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
	oldConfirm := ui.Confirm
	oldInput := ui.Input
	oldSelect := ui.Select
	oldMultiSelect := ui.MultiSelect
	oldPassword := ui.Password

	ui.Confirm = func(_ string, defaultVal bool) (bool, error) {
		return defaultVal, nil
	}
	ui.Input = func(_, defaultVal string) (string, error) {
		return defaultVal, nil
	}
	ui.Select = func(_ string, _ []string, defaultVal string) (string, error) {
		return defaultVal, nil
	}
	ui.MultiSelect = func(_ string, _, defaultSelected []string) ([]string, error) {
		return defaultSelected, nil
	}
	ui.Password = func(_ string) (string, error) {
		return "secret", nil
	}

	return func() {
		ui.Confirm = oldConfirm
		ui.Input = oldInput
		ui.Select = oldSelect
		ui.MultiSelect = oldMultiSelect
		ui.Password = oldPassword
	}
}

// ============================================================================
// T - Tests (Unit and Table-driven Tests)
// ============================================================================

func TestConfigCmd_Subcommands(t *testing.T) {
	subcommands := ConfigCmd.Commands()
	expected := map[string]bool{
		"list":                    false,
		"edit":                    false,
		"email":                   false,
		"dotfiles-repo":           false,
		"dotfiles-repo-url":       false,
		"dotfiles-branch":         false,
		"dotfiles-bare-repo-path": false,
		"git-dev-path":            false,
		"ide-url [url]":           false,
		"verbose":                 false,
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
	ui.Confirm = func(_ string, _ bool) (bool, error) {
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
	ui.Confirm = func(msg string, _ bool) (bool, error) {
		return true, nil // User confirms
	}

	VerboseCmd.Run(VerboseCmd, []string{})
	if !viper.GetBool("verbose") {
		t.Error("Expected verbose to remain true")
	}

	// Test Case 2: User denies, changes setting to false
	ui.Confirm = func(msg string, _ bool) (bool, error) {
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
	ui.Confirm = func(_ string, _ bool) (bool, error) {
		return true, nil
	}

	EmailCmd.Run(EmailCmd, []string{})
	if viper.GetString("user-email") != "old@example.com" {
		t.Errorf("Expected email to remain old@example.com, got %s", viper.GetString("user-email"))
	}

	// Test Case 2: Reject existing, update to new
	ui.Confirm = func(_ string, _ bool) (bool, error) {
		return false, nil
	}
	ui.Input = func(_, _ string) (string, error) {
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
	ui.Input = func(_, _ string) (string, error) {
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
	ui.Input = func(_, _ string) (string, error) {
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
	ui.Select = func(_ string, _ []string, _ string) (string, error) {
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
	ui.Input = func(_, _ string) (string, error) {
		return "~/.new-bare", nil
	}

	DotfilesBareRepoPathCmd.Run(DotfilesBareRepoPathCmd, []string{})
	if viper.GetString("dotfiles.bare_repo_path") != "~/.new-bare" {
		t.Errorf("expected bare repo path to update, got %s", viper.GetString("dotfiles.bare_repo_path"))
	}
}

func TestGitDevPathCmd(t *testing.T) {
	setupTestViper(t)
	restore := mockAllPrompts()
	defer restore()

	viper.Set("git.dev_path", "/old/dev")
	// Confirm prompt returns false, triggering update prompt
	ui.Confirm = func(_ string, _ bool) (bool, error) {
		return false, nil
	}
	ui.Input = func(_, _ string) (string, error) {
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
	ui.Confirm = func(_ string, _ bool) (bool, error) {
		return false, nil // reject existing
	}
	ui.Input = func(_, _ string) (string, error) {
		return "https://new.example.com/ide.tar.gz", nil
	}

	IdeURLCmd.Run(IdeURLCmd, []string{})
	if viper.GetString("antigravity.ide_download_url") != "https://new.example.com/ide.tar.gz" {
		t.Errorf("expected antigravity.ide_download_url to update to new url, got %s",
			viper.GetString("antigravity.ide_download_url"))
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
	// Subcommand Count: 10
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
