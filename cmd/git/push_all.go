package git

import (
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/log"
	"github.com/eng618/eng/utils/repo"
)

// PushAllCmd defines the cobra command for pushing all git repositories.
// It pushes commits for all repositories in the development folder that have unpushed commits.
var PushAllCmd = &cobra.Command{
	Use:   "push-all",
	Short: "Push all git repositories in development folder",
	Long:  `This command pushes commits for all git repositories found in your development folder that have unpushed commits.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Pushing all git repositories")

		isVerbose := utils.IsVerbose(cmd)
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		force, _ := cmd.Flags().GetBool("force")

		devPath, err := getWorkingPath(cmd)
		if err != nil {
			log.Error("%s", err)
			return
		}

		log.Verbose(isVerbose, "Development path: %s", devPath)

		if dryRun {
			log.Info("Dry run mode - no actual git operations will be performed")
		}

		if force {
			log.Warn("Force push mode enabled - this will force push to remote")
		}

		repos, err := findGitRepositories(devPath)
		if err != nil {
			log.Error("Failed to find git repositories: %s", err)
			return
		}

		if len(repos) == 0 {
			log.Warn("No git repositories found in %s", devPath)
			return
		}

		log.Info("Found %d git repositories", len(repos))

		successCount := 0
		failureCount := 0
		skippedCount := 0

		for _, repoPath := range repos {
			repoName := filepath.Base(repoPath)
			log.Info("Checking repository: %s", repoName)

			if dryRun {
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
					log.Info("  [DRY RUN] No unpushed commits, skipping: %s", repoPath)
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
				log.Info("  No unpushed commits, skipping...")
				skippedCount++
				continue
			}

			// Check if repository is dirty (only warn, don't skip)
			isDirty, err := repo.IsDirty(repoPath)
			if err != nil {
				log.Error("  Failed to check repository status: %s", err)
				failureCount++
				continue
			}

			if isDirty {
				log.Warn("  Repository has uncommitted changes, but proceeding with push...")
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

		log.Info("Push completed: %d successful, %d failed, %d skipped", successCount, failureCount, skippedCount)

		if failureCount > 0 {
			log.Warn("Some repositories failed to push. Check the output above for details.")
		} else {
			log.Success("All git repositories with unpushed commits pushed successfully")
		}
	},
}

func init() {
	PushAllCmd.Flags().Bool("dry-run", false, "Perform a dry run without making actual changes")
	PushAllCmd.Flags().Bool("force", false, "Force push to remote (use with caution)")
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
		log.Error("Git push output: %s", string(output))
		return err
	}
	log.Info("Git push output: %s", string(output))
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
