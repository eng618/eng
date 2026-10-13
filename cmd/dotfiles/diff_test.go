package dotfiles

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
)

func TestBuildDiffArgs(t *testing.T) {
	tests := []struct {
		name string
		opts diffOptions
		want []string
	}{
		{
			name: "default unstaged",
			opts: diffOptions{},
			want: []string{"--git-dir=/tmp/repo", "--work-tree=/tmp/wt", "diff"},
		},
		{
			name: "staged",
			opts: diffOptions{Staged: true},
			want: []string{"--git-dir=/tmp/repo", "--work-tree=/tmp/wt", "diff", "--cached"},
		},
		{
			name: "stat",
			opts: diffOptions{Stat: true},
			want: []string{"--git-dir=/tmp/repo", "--work-tree=/tmp/wt", "diff", "--stat"},
		},
		{
			name: "name-only with paths",
			opts: diffOptions{NameOnly: true, Paths: []string{".zshrc", ".config/nvim"}},
			want: []string{
				"--git-dir=/tmp/repo",
				"--work-tree=/tmp/wt",
				"diff",
				"--name-only",
				"--",
				".zshrc",
				".config/nvim",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildDiffArgs("/tmp/repo", "/tmp/wt", tt.opts)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiffCmd_MissingConfig(t *testing.T) {
	viper.Reset()
	ui.DisableProgress = true
	defer func() { ui.DisableProgress = false }()

	cmd := &cobra.Command{}
	DiffCmd.Run(cmd, []string{})
}

func TestDiffCmd_SuccessAndFailure(t *testing.T) {
	viper.Reset()
	viper.Set("dotfiles.bare_repo_path", "/tmp/repo")
	viper.Set("dotfiles.worktree_path", "/tmp/worktree")
	ui.DisableProgress = true
	defer func() { ui.DisableProgress = false }()

	original := runDiff
	defer func() { runDiff = original }()

	called := 0
	var gotOpts []diffOptions
	runDiff = func(repoPath, worktreePath string, opts diffOptions) error {
		called++
		gotOpts = append(gotOpts, opts)
		if called == 1 {
			return errors.New("simulated diff failure")
		}
		return nil
	}

	cmd := &cobra.Command{}
	// First call => failure path
	DiffCmd.Run(cmd, []string{})
	// Second call => success path
	DiffCmd.Run(cmd, []string{})

	if called != 2 {
		t.Fatalf("expected 2 calls, got %d", called)
	}
}

func TestDiffCmd_StagedFlagMapping(t *testing.T) {
	viper.Reset()
	viper.Set("dotfiles.bare_repo_path", "/tmp/repo")
	viper.Set("dotfiles.worktree_path", "/tmp/worktree")
	ui.DisableProgress = true
	defer func() { ui.DisableProgress = false }()

	original := runDiff
	defer func() { runDiff = original }()

	var got diffOptions
	runDiff = func(_, _ string, opts diffOptions) error {
		got = opts
		return nil
	}

	// Build a command that mirrors DiffCmd flags so GetBool works.
	// Run reads flags from the passed-in cmd, so define them there.
	cmd := &cobra.Command{}
	cmd.Flags().Bool("staged", false, "")
	cmd.Flags().Bool("cached", false, "")
	cmd.Flags().Bool("stat", false, "")
	cmd.Flags().Bool("name-only", false, "")

	if err := cmd.Flags().Set("cached", "true"); err != nil {
		t.Fatal(err)
	}
	DiffCmd.Run(cmd, []string{".zshrc"})
	if !got.Staged {
		t.Error("expected --cached to map to staged diff")
	}
	if len(got.Paths) != 1 || got.Paths[0] != ".zshrc" {
		t.Errorf("expected paths [.zshrc], got %v", got.Paths)
	}
}

func TestStatusCmd_DiffFlag(t *testing.T) {
	viper.Reset()
	viper.Set("dotfiles.bare_repo_path", "/tmp/repo")
	viper.Set("dotfiles.worktree_path", "/tmp/worktree")
	ui.DisableProgress = true
	defer func() { ui.DisableProgress = false }()

	origStatus := checkStatus
	origDiff := runDiff
	origStat := diffStatSummary
	defer func() {
		checkStatus = origStatus
		runDiff = origDiff
		diffStatSummary = origStat
	}()

	checkStatus = func(_, _ string) error { return nil }
	diffCalls := 0
	runDiff = func(_, _ string, _ diffOptions) error {
		diffCalls++
		return nil
	}
	diffStatSummary = func(_, _ string) (string, error) {
		t.Error("stat summary must not run when --diff is set")
		return "", nil
	}

	cmd := &cobra.Command{}
	cmd.Flags().BoolP("diff", "d", false, "")
	cmd.Flags().Bool("staged", false, "")
	cmd.Flags().BoolP("yes", "y", false, "")

	if err := cmd.Flags().Set("diff", "true"); err != nil {
		t.Fatal(err)
	}
	StatusCmd.Run(cmd, []string{})
	if diffCalls != 1 {
		t.Fatalf("expected runDiff once, got %d", diffCalls)
	}
}

func TestStatusCmd_StatHint(t *testing.T) {
	var buf bytes.Buffer
	log.SetWriters(&buf, &buf)
	defer log.ResetWriters()

	viper.Reset()
	viper.Set("dotfiles.bare_repo_path", "/tmp/repo")
	viper.Set("dotfiles.worktree_path", "/tmp/worktree")
	ui.DisableProgress = true
	defer func() { ui.DisableProgress = false }()

	origStatus := checkStatus
	origStat := diffStatSummary
	origCanPrompt := statusCanPrompt
	defer func() {
		checkStatus = origStatus
		diffStatSummary = origStat
		statusCanPrompt = origCanPrompt
	}()

	checkStatus = func(_, _ string) error { return nil }
	diffStatSummary = func(_, _ string) (string, error) {
		return " .zshrc | 2 +-\n 1 file changed", nil
	}
	// Hint path must not prompt in this test.
	statusCanPrompt = func() bool { return false }

	cmd := &cobra.Command{}
	cmd.Flags().BoolP("diff", "d", false, "")
	cmd.Flags().Bool("staged", false, "")
	cmd.Flags().BoolP("yes", "y", false, "")

	StatusCmd.Run(cmd, []string{})
	out := buf.String()
	if !strings.Contains(out, "eng cfg diff") {
		t.Fatalf("expected hint to run eng cfg diff, got: %s", out)
	}
	if !strings.Contains(out, ".zshrc") {
		t.Fatalf("expected stat output, got: %s", out)
	}
}
