package system

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestSetupASDF(t *testing.T) {
	tempDir := t.TempDir()
	toolVersionsPath := filepath.Join(tempDir, ".tool-versions")

	// Write a fake .tool-versions file
	plugins := []string{"nodejs 20.0.0", "python 3.11.0"}
	content := strings.Join(plugins, "\n")
	if err := os.WriteFile(toolVersionsPath, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to write .tool-versions: %v", err)
	}

	// Set HOME to tempDir for this test
	homeOrig := os.Getenv("HOME")
	_ = os.Setenv("HOME", tempDir)
	defer func() {
		if err := os.Setenv("HOME", homeOrig); err != nil {
			t.Fatalf("Failed to restore HOME: %v", err)
		}
	}()

	// Mock exec.Command for asdf plugin add and install
	// This is a simple test to check the file parsing and command invocation logic
	// For real integration, use a test double or exec wrapper
	called := []string{}
	execCommand = func(name string, args ...string) *exec.Cmd {
		called = append(called, name+" "+strings.Join(args, " "))
		return exec.Command("echo", "mock")
	}
	defer func() { execCommand = exec.Command }()

	setupASDF(false)

	if len(called) == 0 {
		t.Error("No commands were called, expected asdf plugin add and install")
	}
	foundInstall := false
	for _, c := range called {
		if strings.Contains(c, "asdf install") {
			foundInstall = true
		}
	}
	if !foundInstall {
		t.Error("asdf install was not called")
	}
}

func TestSetupOhMyZsh(t *testing.T) {
	tempDir := t.TempDir()
	homeOrig := os.Getenv("HOME")
	_ = os.Setenv("HOME", tempDir)
	defer func() { _ = os.Setenv("HOME", homeOrig) }()

	called := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "sh" && len(args) > 1 && strings.Contains(args[1], "ohmyzsh") {
			called = true
		}
		return exec.Command("echo", "mock")
	}
	defer func() { execCommand = exec.Command }()

	setupOhMyZsh(false)

	if !called {
		t.Error("Oh My Zsh installation command was not called")
	}
}

func TestSetupDotfiles(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	called := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		// setupDotfiles calls the executable with "dotfiles", "install"
		for _, arg := range args {
			if arg == "dotfiles" {
				called = true
			}
		}
		return exec.Command("echo", "mock")
	}
	defer func() { execCommand = exec.Command }()

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
	defer func() { execCommand = exec.Command }()

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
	defer func() { execCommand = exec.Command }()

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

func TestSetupSoftware(t *testing.T) {
	origLookPath := lookPath
	origAskOne := askOne
	origExec := execCommand
	defer func() {
		lookPath = origLookPath
		askOne = origAskOne
		execCommand = origExec
	}()

	lookPath = func(path string) (string, error) {
		return "/usr/bin/" + path, nil
	}
	// Mock select prompt
	askOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
		r := response.(*[]string)
		*r = []string{} // No optional software selected
		return nil
	}
	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("echo", "success")
	}

	setupSoftware(false)
	// If it doesn't panic and reaches here, basic flow works
}

func TestRunSetup_NonInteractive(t *testing.T) {
	origPrereq := ensurePrerequisitesStep
	origZsh := setupOhMyZshStep
	origASDF := setupASDFStep
	origDotfiles := setupDotfilesStep
	origSoftware := setupSoftwareStep
	defer func() {
		ensurePrerequisitesStep = origPrereq
		setupOhMyZshStep = origZsh
		setupASDFStep = origASDF
		setupDotfilesStep = origDotfiles
		setupSoftwareStep = origSoftware
	}()

	var ran []string
	ensurePrerequisitesStep = func(_ bool) error { ran = append(ran, "prerequisites"); return nil }
	setupOhMyZshStep = func(_ bool) { ran = append(ran, "oh-my-zsh") }
	setupASDFStep = func(_ bool) { ran = append(ran, "asdf") }
	setupDotfilesStep = func(_ bool) error { ran = append(ran, "dotfiles"); return nil }
	setupSoftwareStep = func(_ bool) { ran = append(ran, "software") }

	cmd := &cobra.Command{}
	cmd.Flags().BoolP("interactive", "i", false, "")

	if err := runSetup(cmd, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"prerequisites", "oh-my-zsh", "asdf", "dotfiles", "software"}
	if len(ran) != len(expected) {
		t.Fatalf("expected steps %v, got %v", expected, ran)
	}
	for i, want := range expected {
		if ran[i] != want {
			t.Errorf("step %d: want %q, got %q", i, want, ran[i])
		}
	}
}

func TestRunSetup_Interactive_SkipStep(t *testing.T) {
	origPrereq := ensurePrerequisitesStep
	origZsh := setupOhMyZshStep
	origASDF := setupASDFStep
	origDotfiles := setupDotfilesStep
	origSoftware := setupSoftwareStep
	origAskOne := askOne
	defer func() {
		ensurePrerequisitesStep = origPrereq
		setupOhMyZshStep = origZsh
		setupASDFStep = origASDF
		setupDotfilesStep = origDotfiles
		setupSoftwareStep = origSoftware
		askOne = origAskOne
	}()

	var ran []string
	ensurePrerequisitesStep = func(_ bool) error { ran = append(ran, "prerequisites"); return nil }
	// oh-my-zsh will be skipped below
	setupOhMyZshStep = func(_ bool) { ran = append(ran, "oh-my-zsh") }
	setupASDFStep = func(_ bool) { ran = append(ran, "asdf") }
	setupDotfilesStep = func(_ bool) error { ran = append(ran, "dotfiles"); return nil }
	setupSoftwareStep = func(_ bool) { ran = append(ran, "software") }

	promptIdx := 0
	askOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
		resp := response.(*string)
		if promptIdx == 1 { // second step = Oh My Zsh → skip it
			*resp = setupActionSkip
		} else {
			*resp = setupActionContinue
		}
		promptIdx++
		return nil
	}

	cmd := &cobra.Command{}
	cmd.Flags().BoolP("interactive", "i", false, "")
	if err := cmd.Flags().Set("interactive", "true"); err != nil {
		t.Fatalf("could not set interactive flag: %v", err)
	}

	if err := runSetup(cmd, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, step := range ran {
		if step == "oh-my-zsh" {
			t.Error("oh-my-zsh step should have been skipped")
		}
	}

	// All other steps should have run
	for _, want := range []string{"prerequisites", "asdf", "dotfiles", "software", "gpg-permissions"} {
		found := false
		for _, got := range ran {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected step %q to run, but it did not", want)
		}
	}
}

func TestRunSetup_Interactive_ExitEarly(t *testing.T) {
	origPrereq := ensurePrerequisitesStep
	origZsh := setupOhMyZshStep
	origAskOne := askOne
	defer func() {
		ensurePrerequisitesStep = origPrereq
		setupOhMyZshStep = origZsh
		askOne = origAskOne
	}()

	zshRan := false
	ensurePrerequisitesStep = func(_ bool) error { return nil }
	setupOhMyZshStep = func(_ bool) { zshRan = true }

	promptIdx := 0
	askOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
		resp := response.(*string)
		if promptIdx == 1 { // second step = Oh My Zsh → exit
			*resp = setupActionExit
		} else {
			*resp = setupActionContinue
		}
		promptIdx++
		return nil
	}

	cmd := &cobra.Command{}
	cmd.Flags().BoolP("interactive", "i", false, "")
	if err := cmd.Flags().Set("interactive", "true"); err != nil {
		t.Fatalf("could not set interactive flag: %v", err)
	}

	if err := runSetup(cmd, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if zshRan {
		t.Error("oh-my-zsh step should not have run after user chose exit")
	}
}
