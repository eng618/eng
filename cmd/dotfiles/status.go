package dotfiles

import (
	"fmt"
	"os/exec"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

// StatusCmd defines the cobra command for checking the status of the dotfiles repository.
// It shows any local changes, untracked files, or uncommitted modifications.
var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "check the status of your dotfiles repository",
	Long:  `This command checks the status of your local bare dotfiles repository to see if there are any local changes.`,
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

		theme.SuccessMessage("Dotfiles status check complete.")
	},
}

// checkStatus is injectable for tests to avoid executing git.
var checkStatus = func(repoPath, worktreePath string) error {
	gitCmd := exec.Command("git", "--git-dir="+repoPath, "--work-tree="+worktreePath, "status")
	gitCmd.Stdout = log.Writer()
	gitCmd.Stderr = log.ErrorWriter()

	return gitCmd.Run()
}
