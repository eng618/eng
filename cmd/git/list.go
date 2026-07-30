package git

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

// ListCmd defines the cobra command for listing all git repositories.
// It lists all git repositories found in the development folder.
var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all git repositories in development folder",
	Long:  `This command lists all git repositories found in your development folder.`,
	Run: func(cmd *cobra.Command, _args []string) {
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary).
			MarginBottom(1)
		if !ui.DisableProgress {
			fmt.Fprintln(log.Out, headerStyle.Render("📁 Development Git Repositories"))
		}

		isVerbose := cmdutil.IsVerbose(cmd)
		showPaths, _ := cmd.Flags().GetBool("paths")

		devPath, err := getWorkingPath(cmd)
		if err != nil {
			log.Error("%s", err)
			return
		}

		log.Verbose(isVerbose, "Development path: %s", devPath)

		repos, err := findGitRepositories(devPath)
		if err != nil {
			log.Error("Failed to find git repositories: %s", err)
			return
		}

		if len(repos) == 0 {
			log.Warn("No git repositories found in %s", devPath)
			return
		}

		var cardLines []string
		cardLines = append(cardLines, fmt.Sprintf("Found %s repositories in %s:",
			theme.PrimaryText.Bold(true).Render(fmt.Sprintf("%d", len(repos))),
			theme.BoldText.Render(devPath),
		))
		cardLines = append(cardLines, "")

		for i, repoPath := range repos {
			repoName := filepath.Base(repoPath)
			branch := getRepoBranch(repoPath)
			if showPaths {
				cardLines = append(cardLines, fmt.Sprintf("  %-3d %-25s %-15s %s",
					i+1,
					theme.PrimaryText.Render(repoName),
					theme.MutedText.Render(branch),
					theme.BaseText.Render(repoPath),
				))
			} else {
				cardLines = append(cardLines, fmt.Sprintf("  %-3d %-25s %s",
					i+1,
					theme.PrimaryText.Render(repoName),
					theme.MutedText.Render(branch),
				))
			}
		}

		if !ui.DisableProgress {
			boxStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(theme.Primary).
				Padding(0, 1).
				MarginBottom(1)
			fmt.Fprintln(log.Out, boxStyle.Render(strings.Join(cardLines, "\n")))
		} else {
			for _, line := range cardLines {
				log.Info("%s", line)
			}
		}

		theme.SuccessMessage(fmt.Sprintf("Listed %d git repositories", len(repos)))
	},
}

func init() {
	ListCmd.Flags().BoolP("paths", "p", false, "Show full paths for each repository")
}
