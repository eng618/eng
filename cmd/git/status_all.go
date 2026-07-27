package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/repo"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

// StatusAllCmd defines the cobra command for checking status of all git repositories.
var StatusAllCmd = &cobra.Command{
	Use:   "status-all",
	Short: "Check status of all git repositories in development folder",
	Long:  `This command checks the status of all git repositories found in your development folder.`,
	Run: func(cmd *cobra.Command, _args []string) {
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary).
			MarginBottom(1)
		if !ui.DisableProgress {
			fmt.Println(headerStyle.Render("📊 Development Repositories Status"))
		}

		isVerbose := cmdutil.IsVerbose(cmd)

		devPath, err := getWorkingPath(cmd)
		if err != nil {
			log.Error("%s", err)
			return
		}

		log.Verbose(isVerbose, "Development path: %s", devPath)

		var scanSpinner *ui.Spinner
		if !ui.DisableProgress {
			scanSpinner = ui.NewSpinner(fmt.Sprintf("Scanning git repositories in %s...", devPath))
		}

		repos, err := findGitRepositories(devPath)
		if err != nil {
			if scanSpinner != nil {
				scanSpinner.Stop()
			}
			log.Error("Failed to find git repositories: %s", err)
			return
		}

		if scanSpinner != nil {
			scanSpinner.Stop()
		}

		if len(repos) == 0 {
			log.Warn("No git repositories found in %s", devPath)
			return
		}

		type RepoStatusSummary struct {
			Name    string
			Branch  string
			IsDirty bool
		}

		var summaries []RepoStatusSummary
		cleanCount := 0
		dirtyCount := 0

		for _, repoPath := range repos {
			repoName := filepath.Base(repoPath)
			branch := getRepoBranch(repoPath)

			isDirty, err := repo.IsDirty(cmd.Context(), repoPath)
			if err != nil {
				log.Error("  %s: Failed to check status - %s", repoName, err)
				continue
			}

			if isDirty {
				dirtyCount++
			} else {
				cleanCount++
			}

			summaries = append(summaries, RepoStatusSummary{
				Name:    repoName,
				Branch:  branch,
				IsDirty: isDirty,
			})
		}

		// Render Lipgloss Table Box
		var boxLines []string
		boxLines = append(boxLines, fmt.Sprintf("Checked %s repository(ies) in %s:",
			theme.PrimaryText.Bold(true).Render(fmt.Sprintf("%d", len(repos))),
			theme.BoldText.Render(devPath),
		))
		boxLines = append(boxLines, "")
		boxLines = append(boxLines, fmt.Sprintf("  %-30s %-20s %s",
			theme.BoldText.Render("Repository"),
			theme.BoldText.Render("Branch"),
			theme.BoldText.Render("Status"),
		))
		boxLines = append(boxLines, "  "+strings.Repeat("─", 65))

		for _, s := range summaries {
			statusTag := theme.SuccessText.Render("✓ Clean")
			if s.IsDirty {
				statusTag = theme.ErrorText.Render("⚠️ Uncommitted Changes")
			}

			boxLines = append(boxLines, fmt.Sprintf("  %-30s %-20s %s",
				theme.PrimaryText.Render(s.Name),
				theme.MutedText.Render(s.Branch),
				statusTag,
			))
		}

		if !ui.DisableProgress {
			fmt.Println(theme.InfoBox.Render(strings.Join(boxLines, "\n")))
		}

		summaryMsg := fmt.Sprintf(
			"Status summary: %d clean, %d with uncommitted changes across %d repositories.",
			cleanCount,
			dirtyCount,
			len(repos),
		)
		if dirtyCount > 0 {
			theme.WarningMessage(summaryMsg)
		} else {
			theme.SuccessMessage(summaryMsg)
		}
	},
}

func getRepoBranch(repoPath string) string {
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "main"
	}
	return strings.TrimSpace(string(out))
}
