package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/repo"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

// StatusAllCmd defines the cobra command for checking status of all git repositories.
var StatusAllCmd = &cobra.Command{
	Use:   "status-all",
	Short: "Check status of all git repositories in development folder",
	Long:  `Check working-tree status across all git repositories in your development folder. Use --current to scan the current directory.`,
	Example: `  eng git status-all
  eng git status-all --current`,
	Run: func(cmd *cobra.Command, _args []string) {
		printHeader("📊 Development Repositories Status")

		setup, err := setupGitCommand(cmd)
		if err != nil {
			log.Error("%s", err)
			return
		}

		var scanSpinner *ui.Spinner
		if !ui.DisableProgress {
			scanSpinner = ui.NewSpinner(fmt.Sprintf("Scanning git repositories in %s...", setup.DevPath))
			scanSpinner.Start()
		}

		repos, err := findGitRepositories(setup.DevPath)
		if err != nil {
			if scanSpinner != nil {
				scanSpinner.Fail("Scan failed")
			}
			log.Error("Failed to find git repositories: %s", err)
			return
		}

		if scanSpinner != nil {
			scanSpinner.Success(fmt.Sprintf("Scanned %d repositories", len(repos)))
		}

		if len(repos) == 0 {
			log.Warn("No git repositories found in %s", setup.DevPath)
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
			theme.BoldText.Render(setup.DevPath),
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
			fmt.Fprintln(log.Out, theme.InfoBox.Render(strings.Join(boxLines, "\n")))
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
