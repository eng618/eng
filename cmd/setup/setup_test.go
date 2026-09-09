package setup

import (
	"testing"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/ui"
)

func TestRunSetup_NonInteractive(t *testing.T) {
	origPrereq := ensurePrerequisitesStep
	origZsh := setupOhMyZshStep
	origASDF := setupASDFStep
	origDotfiles := setupDotfilesStep
	origSoftware := setupSoftwareStep
	origGPG := setupGPGStep
	defer func() {
		ensurePrerequisitesStep = origPrereq
		setupOhMyZshStep = origZsh
		setupASDFStep = origASDF
		setupDotfilesStep = origDotfiles
		setupSoftwareStep = origSoftware
		setupGPGStep = origGPG
	}()

	var ran []string
	ensurePrerequisitesStep = func(_ bool) error { ran = append(ran, "prerequisites"); return nil }
	setupOhMyZshStep = func(_ bool) { ran = append(ran, "oh-my-zsh") }
	setupASDFStep = func(_ bool) { ran = append(ran, "asdf") }
	setupDotfilesStep = func(_ bool) error { ran = append(ran, "dotfiles"); return nil }
	setupSoftwareStep = func(_ bool) { ran = append(ran, "software") }
	setupGPGStep = func(_ bool) error { ran = append(ran, "gpg"); return nil }

	cmd := &cobra.Command{}
	cmd.Flags().BoolP("interactive", "i", false, "")

	if err := runSetup(cmd, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"prerequisites", "oh-my-zsh", "asdf", "dotfiles", "software", "gpg"}
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
	origGPG := setupGPGStep
	origWizard := runSetupWizard
	ui.DisableProgress = true
	defer func() {
		ensurePrerequisitesStep = origPrereq
		setupOhMyZshStep = origZsh
		setupASDFStep = origASDF
		setupDotfilesStep = origDotfiles
		setupSoftwareStep = origSoftware
		setupGPGStep = origGPG
		runSetupWizard = origWizard
		ui.DisableProgress = false
	}()

	var ran []string
	ensurePrerequisitesStep = func(_ bool) error { ran = append(ran, "prerequisites"); return nil }
	// oh-my-zsh will be skipped below
	setupOhMyZshStep = func(_ bool) { ran = append(ran, "oh-my-zsh") }
	setupASDFStep = func(_ bool) { ran = append(ran, "asdf") }
	setupDotfilesStep = func(_ bool) error { ran = append(ran, "dotfiles"); return nil }
	setupSoftwareStep = func(_ bool) { ran = append(ran, "software") }
	setupGPGStep = func(_ bool) error { ran = append(ran, "gpg"); return nil }

	runSetupWizard = func(steps []setupStep) ([]string, error) {
		actions := make([]string, len(steps))
		for i := range steps {
			actions[i] = setupActionContinue
		}
		actions[1] = setupActionSkip // skip Oh My Zsh
		return actions, nil
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
	for _, want := range []string{"prerequisites", "asdf", "dotfiles", "software", "gpg"} {
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
	origWizard := runSetupWizard
	ui.DisableProgress = true
	defer func() {
		ensurePrerequisitesStep = origPrereq
		setupOhMyZshStep = origZsh
		runSetupWizard = origWizard
		ui.DisableProgress = false
	}()

	zshRan := false
	ensurePrerequisitesStep = func(_ bool) error { return nil }
	setupOhMyZshStep = func(_ bool) { zshRan = true }

	runSetupWizard = func(steps []setupStep) ([]string, error) {
		actions := make([]string, len(steps))
		for i := range steps {
			actions[i] = setupActionContinue
		}
		actions[1] = setupActionExit // exit at Oh My Zsh
		return actions, nil
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
