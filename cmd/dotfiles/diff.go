package dotfiles

import (
	"bytes"
	"fmt"
	"io"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/execx"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

// DiffCmd defines the cobra command for showing local dotfiles changes.
// It wraps `git --git-dir --work-tree diff` so bare-repo dotfiles diff
// like a normal checkout.
var DiffCmd = &cobra.Command{
	Use:   "diff [path...]",
	Short: "show local dotfiles changes",
	Long: `Show unstaged and staged changes in your local bare dotfiles repository.

This is the dotfiles equivalent of running:
  git --git-dir=<repo> --work-tree=<worktree> diff

By default it shows unstaged changes. Use --staged to show staged
changes, --stat for a summary, or --name-only for just file names.
Pass optional paths to limit the diff, e.g. 'eng cfg diff -- .zshrc'.`,
	Example: `  eng dotfiles diff
  eng cfg diff --staged
  eng cfg diff --stat
  eng cfg diff --name-only
  eng cfg diff -- .zshrc .config/nvim`,
	Run: func(cmd *cobra.Command, args []string) {
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary).
			MarginBottom(1)
		if !ui.DisableProgress {
			fmt.Fprintln(log.Out, headerStyle.Render("📝 Dotfiles Changes"))
		}

		isVerbose := cmdutil.IsVerbose(cmd)

		repoPath, worktreePath, err := getDotfilesConfig()
		if err != nil || repoPath == "" {
			log.Error("Dotfiles repository path is not set in configuration")
			return
		}

		log.Verbose(isVerbose, "Repository path: %s", repoPath)
		log.Verbose(isVerbose, "Worktree path:   %s", worktreePath)

		staged, _ := cmd.Flags().GetBool("staged")
		// --cached is the git-native alias for --staged.
		cached, _ := cmd.Flags().GetBool("cached")
		stat, _ := cmd.Flags().GetBool("stat")
		nameOnly, _ := cmd.Flags().GetBool("name-only")

		opts := diffOptions{
			Staged:   staged || cached,
			Stat:     stat,
			NameOnly: nameOnly,
			Paths:    args,
		}

		// Use injectable function so tests can override and avoid executing git.
		if err := runDiff(repoPath, worktreePath, opts); err != nil {
			log.Error("Failed to show diff: %s", err)
			return
		}

		theme.SuccessMessage("Dotfiles diff complete.")
	},
}

func init() {
	DiffCmd.Flags().Bool("staged", false, "show staged changes (git diff --cached)")
	DiffCmd.Flags().
		Bool("cached", false, "alias for --staged (git-native spelling)")
	DiffCmd.Flags().Bool("stat", false, "show diffstat summary instead of full diff")
	DiffCmd.Flags().Bool("name-only", false, "show only names of changed files")
}

// diffOptions captures the supported `git diff` switches for the bare repo.
type diffOptions struct {
	Staged   bool
	Stat     bool
	NameOnly bool
	Paths    []string
}

// buildDiffArgs constructs the git CLI args for a bare-repo diff.
func buildDiffArgs(repoPath, worktreePath string, opts diffOptions) []string {
	args := []string{"--git-dir=" + repoPath, "--work-tree=" + worktreePath, "diff"}
	if opts.Staged {
		args = append(args, "--cached")
	}
	if opts.Stat {
		args = append(args, "--stat")
	}
	if opts.NameOnly {
		args = append(args, "--name-only")
	}
	if len(opts.Paths) > 0 {
		args = append(args, "--")
		args = append(args, opts.Paths...)
	}
	return args
}

// runDiff streams `git diff` output for the bare repo. Injectable for tests.
var runDiff = func(repoPath, worktreePath string, opts diffOptions) error {
	gitCmd := execx.Command("git", buildDiffArgs(repoPath, worktreePath, opts)...)
	gitCmd.Stdout = log.Writer()
	gitCmd.Stderr = log.ErrorWriter()
	return gitCmd.Run()
}

// diffStatSummary returns `git diff HEAD --stat` output for the bare repo.
// An empty string means no uncommitted changes. Injectable for tests and
// reused by `status` to summarize dirtiness.
var diffStatSummary = func(repoPath, worktreePath string) (string, error) {
	var buf bytes.Buffer
	gitCmd := execx.Command(
		"git",
		"--git-dir="+repoPath,
		"--work-tree="+worktreePath,
		"diff",
		"HEAD",
		"--stat",
	)
	gitCmd.Stdout = &buf
	// Discard stderr: a stat failure (e.g. bare repo without HEAD yet)
	// is non-fatal for `status`, which only wants a best-effort summary.
	gitCmd.Stderr = io.Discard
	if err := gitCmd.Run(); err != nil {
		return "", err
	}
	return buf.String(), nil
}
