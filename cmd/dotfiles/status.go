package dotfiles

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/execx"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

// StatusCmd defines the cobra command for checking the status of the dotfiles repository.
// It shows any local changes, untracked files, or uncommitted modifications.
var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "check the status of your dotfiles repository",
	Long: `This command checks the status of your local bare dotfiles repository to see if there are any local changes.

When changes exist it prints a diff --stat summary with a hint to run
'eng cfg diff' for the full patch. Use --diff to pretty-print the full
diff inline.

When run interactively without --diff and changes are detected, it
offers to show the full diff inline (same rendering as 'eng cfg diff').`,
	Example: `  eng dotfiles status
  eng cfg status --diff
  eng cfg status --diff --staged
  eng cfg status --yes`,
	Run: func(cmd *cobra.Command, args []string) {
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary).
			MarginBottom(1)
		if !ui.DisableProgress {
			fmt.Fprintln(log.Out, headerStyle.Render("📊 Dotfiles Repository Status"))
		}

		isVerbose := cmdutil.IsVerbose(cmd)

		repoPath, worktreePath, err := getDotfilesConfig()
		if err != nil || repoPath == "" {
			log.Error("Dotfiles repository path is not set in configuration")
			return
		}
		log.Verbose(isVerbose, "Repository path: %s", repoPath)
		log.Verbose(isVerbose, "Worktree path:   %s", worktreePath)

		// Use injectable function so tests can override and avoid executing git.
		err = checkStatus(repoPath, worktreePath)
		if err != nil {
			log.Error("Failed to check status: %s", err)
			return
		}

		showDiff, _ := cmd.Flags().GetBool("diff")
		showStaged, _ := cmd.Flags().GetBool("staged")
		autoYes, _ := cmd.Flags().GetBool("yes")
		if showDiff || autoYes {
			opts := diffOptions{Staged: showStaged, Paths: args}
			if err := runDiff(repoPath, worktreePath, opts); err != nil {
				log.Error("Failed to show diff: %s", err)
				return
			}
		} else {
			// Summarize dirtiness so `status` alone answers "what changed?".
			// Stat failures are non-fatal (e.g. fresh bare repo without HEAD).
			if stat, statErr := diffStatSummary(repoPath, worktreePath); statErr != nil {
				log.Verbose(isVerbose, "Skipping diff summary: %s", statErr)
			} else if stat != "" {
				fmt.Fprintln(log.Out, "")
				fmt.Fprintln(log.Out, "Changes:")
				fmt.Fprintln(log.Out, ui.RenderDiffStat(stat))
				log.Info("Run `eng cfg diff` for the full patch, or `eng cfg status --diff`.")
				offerDiffPrompt(isVerbose, repoPath, worktreePath, showStaged, args)
			}
		}

		theme.SuccessMessage("Dotfiles status check complete.")
	},
}

func init() {
	StatusCmd.Flags().BoolP("diff", "d", false, "pretty-print the diff of local changes after status")
	StatusCmd.Flags().Bool("staged", false, "with --diff, show staged changes (git diff --cached)")
	StatusCmd.Flags().
		BoolP("yes", "y", false, "automatically show the full diff when changes exist (skips the prompt)")
}

// offerDiffPrompt intelligently recommends the diff flag when changes were
// detected by diffStatSummary. It prompts interactively (TTY only) and reuses
// runDiff from diff.go so status and diff render identically.
func offerDiffPrompt(isVerbose bool, repoPath, worktreePath string, staged bool, args []string) {
	if !statusCanPrompt() {
		log.Verbose(isVerbose, "Skipping diff prompt: non-interactive terminal.")
		return
	}
	prompt := "Show full diff now? (same as `eng cfg status --diff`)"
	if staged {
		prompt = "Show staged diff now? (same as `eng cfg status --diff --staged`)"
	}
	confirmed, err := statusConfirmPrompt(prompt, false)
	if err != nil {
		log.Verbose(isVerbose, "Diff prompt dismissed: %s", err)
		return
	}
	if !confirmed {
		return
	}
	opts := diffOptions{Staged: staged, Paths: args}
	if err := runDiff(repoPath, worktreePath, opts); err != nil {
		log.Error("Failed to show diff: %s", err)
	}
}

// statusConfirmPrompt asks the diff recommendation question. Injectable for tests.
var statusConfirmPrompt = ui.Confirm

// statusCanPrompt reports whether it is safe to prompt. Injectable for tests.
var statusCanPrompt = func() bool {
	return ui.IsTerminal(os.Stdin) && ui.IsTerminal(os.Stdout)
}

// checkStatus is injectable for tests to avoid executing git.
var checkStatus = func(repoPath, worktreePath string) error {
	gitCmd := execx.Command("git", "--git-dir="+repoPath, "--work-tree="+worktreePath, "status")
	gitCmd.Stdout = log.Writer()
	gitCmd.Stderr = log.ErrorWriter()

	return gitCmd.Run()
}
