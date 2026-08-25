package dotfiles

import (
	"context"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/dotfiles"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

// SyncCmd defines the cobra command for syncing the dotfiles repository.
var SyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "sync your local bare repository",
	Long:  `This command fetches and pulls in remote changes to the local bare dot repository.`,
	Run: func(cmd *cobra.Command, args []string) {
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary).
			MarginBottom(1)
		if !ui.DisableProgress {
			fmt.Fprintln(log.Out, headerStyle.Render("🔄 Syncing Dotfiles Repository"))
		}

		isVerbose := cmdutil.IsVerbose(cmd)

		repoPath, worktreePath, err := getDotfilesConfig()
		if err != nil || repoPath == "" {
			log.Error("Dotfiles repository path is not set in configuration")
			return
		}
		log.Verbose(isVerbose, "Repository path: %s", repoPath)
		log.Verbose(isVerbose, "Worktree path:   %s", worktreePath)

		ctx := cmd.Context()
		if ctx == nil {
			ctx = context.Background()
		}
		if err = dotfiles.SyncRepo(ctx, repoPath, worktreePath, isVerbose); err != nil {
			log.Error("Sync failed: %v", err)
			return
		}

		theme.SuccessMessage("Dotfiles synced successfully.")
	},
}
