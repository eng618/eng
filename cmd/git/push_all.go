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

// PushAllCmd defines the cobra command for pushing all git repositories.
// It pushes commits for all repositories in the development folder that have unpushed commits.
var PushAllCmd = &cobra.Command{
	Use:   "push-all",
	Short: "Push all git repositories in development folder",
	Long:  `Push commits for all git repositories in your development folder that have unpushed commits. Use --dry-run to preview, --force only with --yes confirmation for force-with-lease pushes.`,
	Example: `  eng git push-all --dry-run
  eng git push-all
  eng git push-all --force --yes`,
	Run: func(cmd *cobra.Command, args []string) {
		printHeader("📤 Pushing Git Repositories")

		setup, err := setupGitCommand(cmd)
		if err != nil {
			log.Error("%s", err)
			return
		}

		force, _ := cmd.Flags().GetBool("force")
		assumeYes, _ := cmd.Flags().GetBool("yes")

		if force {
			log.Warn("Force push mode enabled — uses --force-with-lease across all repos")
			if !assumeYes && !setup.DryRun {
				confirmed, err := ui.Confirm(
					"Force-push all repos with unpushed commits? This rewrites remote history.",
					false,
				)
				if err != nil || !confirmed {
					log.Info("Push operation canceled.")
					return
				}
			}
		}

		repos, err := findGitRepositories(setup.DevPath)
		if err != nil {
			log.Error("Failed to find git repositories: %s", err)
			return
		}

		if len(repos) == 0 {
			log.Warn("No git repositories found in %s", setup.DevPath)
			return
		}

		log.Info("Found %d git repositories", len(repos))

		successCount := 0
		failureCount := 0
		skippedCount := 0

		for _, repoPath := range repos {
			repoName := filepath.Base(repoPath)
			log.Verbose(setup.IsVerbose, "Checking repository: %s", repoName)

			if setup.DryRun {
				hasUnpushed, err := hasUnpushedCommits(repoPath)
				if err != nil {
					log.Error("  [DRY RUN] Failed to check for unpushed commits: %s", err)
					failureCount++
					continue
				}
				if hasUnpushed {
					log.Info("  [DRY RUN] Would push repository at: %s", repoPath)
					successCount++
				} else {
					log.Verbose(setup.IsVerbose, "  [DRY RUN] No unpushed commits, skipping: %s", repoPath)
					skippedCount++
				}
				continue
			}

			// Check if repository has unpushed commits
			hasUnpushed, err := hasUnpushedCommits(repoPath)
			if err != nil {
				log.Error("  Failed to check for unpushed commits: %s", err)
				failureCount++
				continue
			}

			if !hasUnpushed {
				log.Verbose(setup.IsVerbose, "  No unpushed commits, skipping: %s", repoName)
				skippedCount++
				continue
			}

			// Check if repository is dirty (only warn, don't skip)
			isDirty, err := repo.IsDirty(cmd.Context(), repoPath)
			if err != nil {
				log.Error("  Failed to check repository status: %s", err)
				failureCount++
				continue
			}

			if isDirty {
				log.Warn("  Repository %s has uncommitted changes, but proceeding with push...", repoName)
			}

			// Push commits
			if err := pushRepository(repoPath, force); err != nil {
				log.Error("  Failed to push %s: %s", repoName, err)
				failureCount++
				continue
			}

			log.Success("  Successfully pushed %s", repoName)
			successCount++
		}

		summaryMsg := fmt.Sprintf(
			"Push completed: %d successful, %d failed, %d skipped across %d repositories.",
			successCount,
			failureCount,
			skippedCount,
			len(repos),
		)
		if failureCount > 0 {
			theme.WarningMessage(summaryMsg)
		} else {
			theme.SuccessMessage(summaryMsg)
		}
	},
}

func init() {
	PushAllCmd.Flags().BoolP("dry-run", "n", false, "Preview what would be pushed without making changes")
	PushAllCmd.Flags().Bool("force", false, "Force push with --force-with-lease (requires confirmation unless --yes)")
	PushAllCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompts")
}

// pushRepository performs a git push operation on the given repository path.
func pushRepository(repoPath string, force bool) error {
	args := []string{"-C", repoPath, "push"}
	if force {
		args = append(args, "--force-with-lease")
	}

	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error("Git push failed for %s: %s", filepath.Base(repoPath), strings.TrimSpace(string(output)))
		return err
	}
	return nil
}

// hasUnpushedCommits checks if the repository has commits that haven't been pushed to remote.
func hasUnpushedCommits(repoPath string) (bool, error) {
	// Check if there are commits ahead of origin
	cmd := exec.Command("git", "-C", repoPath, "rev-list", "--count", "@{upstream}..HEAD")
	output, err := cmd.Output()
	if err != nil {
		// If there's no upstream, consider it as having unpushed commits
		return true, nil
	}

	count := strings.TrimSpace(string(output))
	return count != "0", nil
}
