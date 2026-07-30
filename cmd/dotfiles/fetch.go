package dotfiles

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/dotfiles"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

// FetchCmd defines the cobra command for fetching the dotfiles repository.
var FetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "fetch your local bare repository",
	Long:  `This command fetches remote changes to the local bare dot repository.`,
	Run: func(cmd *cobra.Command, args []string) {
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary).
			MarginBottom(1)
		if !ui.DisableProgress {
			fmt.Fprintln(log.Out, headerStyle.Render("🔍 Fetching Dotfiles Repository"))
		}

		repoPath, worktreePath, err := getDotfilesConfig()
		if err != nil || repoPath == "" {
			log.Error("Dotfiles repository path is not set in configuration")
			return
		}

		err = dotfiles.FetchRepo(cmd.Context(), repoPath, worktreePath)
		if err != nil {
			log.Error("Failed to fetch dotfiles: %s", err)
			return
		}

		theme.SuccessMessage("Dotfiles fetched successfully.")
	},
}
