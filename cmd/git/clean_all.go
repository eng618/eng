package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/asdf"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

// CleanAllCmd defines the cobra command for cleaning untracked files in all git repositories.
var CleanAllCmd = &cobra.Command{
	Use:   "clean-all",
	Short: "Clean untracked files in all git repositories in development folder",
	Long:  `Preview untracked files across all repos, confirm, then clean. Use --dry-run to preview only, --force with --yes to skip confirmation.`,
	Example: `  eng git clean-all --dry-run
  eng git clean-all
  eng git clean-all --force --yes -d`,
	Run: func(cmd *cobra.Command, _args []string) {
		printHeader("🧹 Clean Development Repositories")

		setup, err := setupGitCommand(cmd)
		if err != nil {
			log.Error("%s", err)
			return
		}

		force, _ := cmd.Flags().GetBool("force")
		assumeYes, _ := cmd.Flags().GetBool("yes")
		skipConfirm := force || assumeYes
		directories, _ := cmd.Flags().GetBool("directories")

		var scanSpinner *ui.Spinner
		if !ui.DisableProgress {
			scanSpinner = ui.NewSpinner(fmt.Sprintf(
				"Scanning git repositories in %s for untracked files...",
				setup.DevPath,
			))
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

		type RepoCleanSummary struct {
			Path           string
			Name           string
			UntrackedCount int
			SizeBytes      int64
		}

		var reposToClean []RepoCleanSummary
		var totalUntrackedBytes int64

		for _, repoPath := range repos {
			untrackedFiles, sizeBytes := getUntrackedFilesAndSize(repoPath)
			if len(untrackedFiles) > 0 {
				repoName := filepath.Base(repoPath)
				reposToClean = append(reposToClean, RepoCleanSummary{
					Path:           repoPath,
					Name:           repoName,
					UntrackedCount: len(untrackedFiles),
					SizeBytes:      sizeBytes,
				})
				totalUntrackedBytes += sizeBytes
			}
		}

		if scanSpinner != nil {
			scanSpinner.Success(fmt.Sprintf("Scanned %d repositories", len(repos)))
		}

		if len(reposToClean) == 0 {
			theme.SuccessMessage("No untracked files found across any development repositories. Everything is clean!")
			return
		}

		totalSizeStr := humanize.Bytes(uint64(totalUntrackedBytes))

		// Render Callout Box
		var boxLines []string
		boxLines = append(boxLines, fmt.Sprintf(
			"Found %s repository(ies) with untracked files reclaiming %s space:",
			theme.PrimaryText.Bold(true).Render(fmt.Sprintf("%d", len(reposToClean))),
			theme.SuccessText.Bold(true).Render(totalSizeStr),
		))

		for _, r := range reposToClean {
			sizeStr := ""
			if r.SizeBytes > 0 {
				sizeStr = fmt.Sprintf(" (%s)", humanize.Bytes(uint64(r.SizeBytes)))
			}
			boxLines = append(boxLines, fmt.Sprintf(
				"  • %s: %s untracked file(s)%s",
				theme.BoldText.Render(r.Name),
				theme.MutedText.Render(fmt.Sprintf("%d", r.UntrackedCount)),
				theme.MutedText.Render(sizeStr),
			))
		}

		if !ui.DisableProgress {
			fmt.Fprintln(log.Out, theme.InfoBox.Render(strings.Join(boxLines, "\n")))
		}

		if setup.DryRun {
			theme.InfoMessage(
				fmt.Sprintf(
					"Dry run complete. Previewed cleaning %d repository(ies) reclaiming %s disk space.",
					len(reposToClean),
					totalSizeStr,
				),
			)
			return
		}

		if !skipConfirm {
			confirmMsg := fmt.Sprintf(
				"Clean untracked files across %d repository(ies) to reclaim %s?",
				len(reposToClean),
				totalSizeStr,
			)
			confirmed, err := ui.Confirm(confirmMsg, false)
			if err != nil || !confirmed {
				log.Info("Clean operation canceled.")
				return
			}
		}

		totalCount := len(reposToClean)
		var progressSpinner *ui.Spinner
		if !ui.DisableProgress {
			progressSpinner = ui.NewProgressSpinner(fmt.Sprintf("Cleaning %d git repository(ies)...", totalCount))
		}

		var successCount int
		var failureCount int
		var freedBytes int64

		for i, r := range reposToClean {
			ratio := float64(i+1) / float64(totalCount)
			statusMsg := fmt.Sprintf("[%d/%d] Cleaning %s", i+1, totalCount, r.Name)

			if progressSpinner != nil {
				progressSpinner.SetProgressBar(ratio, statusMsg)
			}

			if err := cleanRepository(r.Path, directories); err != nil {
				failureCount++
				if progressSpinner != nil {
					progressSpinner.Logf("  %s Failed to clean %s: %v\n", theme.ErrorText.Render("✗"), r.Name, err)
				} else {
					log.Error("Failed to clean %s: %v", r.Name, err)
				}
			} else {
				successCount++
				freedBytes += r.SizeBytes
				sizeStr := ""
				if r.SizeBytes > 0 {
					sizeStr = fmt.Sprintf(" (freed %s)", humanize.Bytes(uint64(r.SizeBytes)))
				}
				if progressSpinner != nil {
					progressSpinner.Logf("  %s Cleaned %s%s\n", theme.SuccessText.Render("✓"), r.Name, sizeStr)
				} else {
					log.Success("Cleaned %s%s", r.Name, sizeStr)
				}
			}
		}

		if progressSpinner != nil {
			progressSpinner.SetProgressBar(1.0, "Cleaning complete")
			progressSpinner.Stop()
		}

		freedStr := humanize.Bytes(uint64(freedBytes))
		if failureCount > 0 {
			theme.WarningMessage(
				fmt.Sprintf(
					"Cleaned %d repository(ies) freeing %s, but %d failed.",
					successCount,
					freedStr,
					failureCount,
				),
			)
		} else {
			theme.SuccessMessage(
				fmt.Sprintf(
					"Successfully cleaned %d git repository(ies) freeing %s of disk space!",
					successCount,
					freedStr,
				),
			)
		}
	},
}

func init() {
	CleanAllCmd.Flags().BoolP("dry-run", "n", false, "Preview what would be cleaned without making changes")
	CleanAllCmd.Flags().Bool("force", false, "Skip confirmation prompt (use with --yes in scripts)")
	CleanAllCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	CleanAllCmd.Flags().BoolP("directories", "d", false, "Also remove untracked directories")
}

func cleanRepository(repoPath string, directories bool) error {
	args := []string{"-C", repoPath, "clean", "-f"}
	if directories {
		args = append(args, "-d")
	}

	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error("Git clean output: %s", string(output))
		return err
	}
	return nil
}

func getUntrackedFilesAndSize(repoPath string) ([]string, int64) {
	cmd := exec.Command("git", "-C", repoPath, "ls-files", "--others", "--exclude-standard")
	output, err := cmd.Output()
	if err != nil || len(output) == 0 {
		return nil, 0
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var files []string
	var totalBytes int64

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		fullPath := filepath.Join(repoPath, trimmed)
		files = append(files, fullPath)
		if info, err := os.Stat(fullPath); err == nil {
			if !info.IsDir() {
				totalBytes += info.Size()
			} else {
				totalBytes += asdf.CalculateDirSize(fullPath)
			}
		}
	}

	return files, totalBytes
}

func hasUntrackedFiles(repoPath string) (bool, error) {
	files, _ := getUntrackedFilesAndSize(repoPath)
	return len(files) > 0, nil
}

func showUntrackedFiles(repoPath string) error {
	cmd := exec.Command("git", "-C", repoPath, "clean", "-n")
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	if len(output) > 0 {
		log.Info("    Files to be cleaned: %s", string(output))
	}
	return nil
}
