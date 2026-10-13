package dotfiles

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/ui"
)

func TestStatusCmd_MissingConfig(t *testing.T) {
	// Ensure viper has no config for dotfiles
	viper.Reset()

	// Run command; should return early without panic
	cmd := &cobra.Command{}
	StatusCmd.Run(cmd, []string{})
}

func TestStatusCmd_SuccessAndFailure(t *testing.T) {
	viper.Reset()
	viper.Set("dotfiles.bare_repo_path", "/tmp/repo")
	viper.Set("dotfiles.worktree_path", "/tmp/worktree")

	origStatus := checkStatus
	origStat := diffStatSummary
	origCanPrompt := statusCanPrompt
	t.Cleanup(func() {
		checkStatus = origStatus
		diffStatSummary = origStat
		statusCanPrompt = origCanPrompt
	})
	// Never prompt in this test; stat path returns empty (clean tree).
	statusCanPrompt = func() bool { return false }
	diffStatSummary = func(_, _ string) (string, error) { return "", nil }

	called := 0
	// Override to simulate failure then success
	checkStatus = func(repoPath, worktreePath string) error {
		called++
		if called == 1 {
			return errors.New("simulated status check failure")
		}
		return nil
	}

	cmd := &cobra.Command{}
	// First call => failure path
	StatusCmd.Run(cmd, []string{})
	// Second call => success path
	StatusCmd.Run(cmd, []string{})
}

func TestStatusCmd_PromptConfirmedRunsDiff(t *testing.T) {
	viper.Reset()
	viper.Set("dotfiles.bare_repo_path", "/tmp/repo")
	viper.Set("dotfiles.worktree_path", "/tmp/worktree")
	ui.DisableProgress = true
	defer func() { ui.DisableProgress = false }()

	origStatus := checkStatus
	origStat := diffStatSummary
	origDiff := runDiff
	origConfirm := statusConfirmPrompt
	origCanPrompt := statusCanPrompt
	defer func() {
		checkStatus = origStatus
		diffStatSummary = origStat
		runDiff = origDiff
		statusConfirmPrompt = origConfirm
		statusCanPrompt = origCanPrompt
	}()

	checkStatus = func(_, _ string) error { return nil }
	diffStatSummary = func(_, _ string) (string, error) {
		return " .zshrc | 2 +-\n 1 file changed", nil
	}
	statusCanPrompt = func() bool { return true }
	prompted := false
	statusConfirmPrompt = func(_ string, def bool) (bool, error) {
		prompted = true
		if def {
			t.Error("expected prompt default to be false ([y/N])")
		}
		return true, nil
	}
	diffCalls := 0
	var gotOpts diffOptions
	var gotPaths []string
	runDiff = func(_, _ string, opts diffOptions) error {
		diffCalls++
		gotOpts = opts
		gotPaths = opts.Paths
		return nil
	}

	cmd := &cobra.Command{}
	cmd.Flags().BoolP("diff", "d", false, "")
	cmd.Flags().Bool("staged", false, "")
	cmd.Flags().BoolP("yes", "y", false, "")

	StatusCmd.Run(cmd, []string{".zshrc"})

	if !prompted {
		t.Fatal("expected diff prompt when changes exist")
	}
	if diffCalls != 1 {
		t.Fatalf("expected runDiff once after confirm, got %d", diffCalls)
	}
	if gotOpts.Staged {
		t.Error("expected unstaged diff by default")
	}
	if len(gotPaths) != 1 || gotPaths[0] != ".zshrc" {
		t.Errorf("expected paths [.zshrc] forwarded to runDiff, got %v", gotPaths)
	}
}

func TestStatusCmd_PromptDeclinedSkipsDiff(t *testing.T) {
	viper.Reset()
	viper.Set("dotfiles.bare_repo_path", "/tmp/repo")
	viper.Set("dotfiles.worktree_path", "/tmp/worktree")
	ui.DisableProgress = true
	defer func() { ui.DisableProgress = false }()

	origStatus := checkStatus
	origStat := diffStatSummary
	origDiff := runDiff
	origConfirm := statusConfirmPrompt
	origCanPrompt := statusCanPrompt
	defer func() {
		checkStatus = origStatus
		diffStatSummary = origStat
		runDiff = origDiff
		statusConfirmPrompt = origConfirm
		statusCanPrompt = origCanPrompt
	}()

	checkStatus = func(_, _ string) error { return nil }
	diffStatSummary = func(_, _ string) (string, error) {
		return " .zshrc | 2 +-\n 1 file changed", nil
	}
	statusCanPrompt = func() bool { return true }
	statusConfirmPrompt = func(_ string, _ bool) (bool, error) { return false, nil }
	runDiff = func(_, _ string, _ diffOptions) error {
		t.Error("runDiff must not run when prompt is declined")
		return nil
	}

	cmd := &cobra.Command{}
	cmd.Flags().BoolP("diff", "d", false, "")
	cmd.Flags().Bool("staged", false, "")
	cmd.Flags().BoolP("yes", "y", false, "")

	StatusCmd.Run(cmd, []string{})
}

func TestStatusCmd_NoPromptWhenClean(t *testing.T) {
	viper.Reset()
	viper.Set("dotfiles.bare_repo_path", "/tmp/repo")
	viper.Set("dotfiles.worktree_path", "/tmp/worktree")
	ui.DisableProgress = true
	defer func() { ui.DisableProgress = false }()

	origStatus := checkStatus
	origStat := diffStatSummary
	origConfirm := statusConfirmPrompt
	origCanPrompt := statusCanPrompt
	defer func() {
		checkStatus = origStatus
		diffStatSummary = origStat
		statusConfirmPrompt = origConfirm
		statusCanPrompt = origCanPrompt
	}()

	checkStatus = func(_, _ string) error { return nil }
	diffStatSummary = func(_, _ string) (string, error) { return "", nil }
	statusCanPrompt = func() bool { return true }
	statusConfirmPrompt = func(_ string, _ bool) (bool, error) {
		t.Error("must not prompt when working tree is clean")
		return false, nil
	}

	cmd := &cobra.Command{}
	cmd.Flags().BoolP("diff", "d", false, "")
	cmd.Flags().Bool("staged", false, "")
	cmd.Flags().BoolP("yes", "y", false, "")

	StatusCmd.Run(cmd, []string{})
}

func TestStatusCmd_YesFlagSkipsPrompt(t *testing.T) {
	viper.Reset()
	viper.Set("dotfiles.bare_repo_path", "/tmp/repo")
	viper.Set("dotfiles.worktree_path", "/tmp/worktree")
	ui.DisableProgress = true
	defer func() { ui.DisableProgress = false }()

	origStatus := checkStatus
	origDiff := runDiff
	origStat := diffStatSummary
	origConfirm := statusConfirmPrompt
	origCanPrompt := statusCanPrompt
	defer func() {
		checkStatus = origStatus
		runDiff = origDiff
		diffStatSummary = origStat
		statusConfirmPrompt = origConfirm
		statusCanPrompt = origCanPrompt
	}()

	checkStatus = func(_, _ string) error { return nil }
	diffStatSummary = func(_, _ string) (string, error) {
		t.Error("stat summary must not run when --yes is set")
		return "", nil
	}
	statusConfirmPrompt = func(_ string, _ bool) (bool, error) {
		t.Error("prompt must not run when --yes is set")
		return false, nil
	}
	// Non-interactive guard must not matter for --yes.
	statusCanPrompt = func() bool { return false }
	var got diffOptions
	diffCalls := 0
	runDiff = func(_, _ string, opts diffOptions) error {
		diffCalls++
		got = opts
		return nil
	}

	cmd := &cobra.Command{}
	cmd.Flags().BoolP("diff", "d", false, "")
	cmd.Flags().Bool("staged", false, "")
	cmd.Flags().BoolP("yes", "y", false, "")
	if err := cmd.Flags().Set("yes", "true"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("staged", "true"); err != nil {
		t.Fatal(err)
	}

	StatusCmd.Run(cmd, []string{})

	if diffCalls != 1 {
		t.Fatalf("expected runDiff once with --yes, got %d", diffCalls)
	}
	if !got.Staged {
		t.Error("expected --yes to respect --staged")
	}
}

func TestStatusCmd_PromptErrorIsNonFatal(t *testing.T) {
	viper.Reset()
	viper.Set("dotfiles.bare_repo_path", "/tmp/repo")
	viper.Set("dotfiles.worktree_path", "/tmp/worktree")
	ui.DisableProgress = true
	defer func() { ui.DisableProgress = false }()

	origStatus := checkStatus
	origStat := diffStatSummary
	origDiff := runDiff
	origConfirm := statusConfirmPrompt
	origCanPrompt := statusCanPrompt
	defer func() {
		checkStatus = origStatus
		diffStatSummary = origStat
		runDiff = origDiff
		statusConfirmPrompt = origConfirm
		statusCanPrompt = origCanPrompt
	}()

	checkStatus = func(_, _ string) error { return nil }
	diffStatSummary = func(_, _ string) (string, error) {
		return " .zshrc | 2 +-\n 1 file changed", nil
	}
	statusCanPrompt = func() bool { return true }
	statusConfirmPrompt = func(_ string, _ bool) (bool, error) {
		return false, errors.New("user aborted")
	}
	runDiff = func(_, _ string, _ diffOptions) error {
		t.Error("runDiff must not run when prompt errors")
		return nil
	}

	cmd := &cobra.Command{}
	cmd.Flags().BoolP("diff", "d", false, "")
	cmd.Flags().Bool("staged", false, "")
	cmd.Flags().BoolP("yes", "y", false, "")

	// Should return normally (non-fatal) without panic.
	StatusCmd.Run(cmd, []string{})
}
