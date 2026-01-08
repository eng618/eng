package git

import (
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/utils"
	"github.com/eng618/eng/internal/utils/log"
)

// StashAllCmd defines the cobra command for stashing changes in all git repositories.
// It stashes uncommitted changes for all repositories in the development folder that have changes.
var StashAllCmd = &cobra.Command{
	Use:   "stash-all",
	Short: "Stash changes in all git repositories in development folder",
	Long:  `This command stashes uncommitted changes for all git repositories found in your development folder that have uncommitted changes.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Stashing changes in all git repositories")

		isVerbose := utils.IsVerbose(cmd)
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		message, _ := cmd.Flags().GetString("message")

		devPath, err := getWorkingPath(cmd)
		if err != nil {
			log.Error("%s", err)
			return
		}

		log.Verbose(isVerbose, "Development path: %s", devPath)

		if dryRun {
			log.Info("Dry run mode - no actual git operations will be performed")
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
				hasChanges, err := hasUncommittedChanges(repoPath)
				if err != nil {
					log.Error("  [DRY RUN] Failed to check for changes: %s", err)
					failureCount++
					continue
				}
				if hasChanges {
					log.Info("  [DRY RUN] Would stash changes in: %s", repoPath)
					successCount++
				} else {
					log.Info("  [DRY RUN] No changes to stash, skipping: %s", repoPath)
					skippedCount++
				}
				continue
			}

			// Check if repository has uncommitted changes
			hasChanges, err := hasUncommittedChanges(repoPath)
			if err != nil {
				log.Error("  Failed to check for uncommitted changes: %s", err)
				failureCount++
				continue
			}

			if !hasChanges {
				log.Info("  No uncommitted changes, skipping...")
				skippedCount++
				continue
			}

			// Stash changes
			if err := stashRepository(repoPath, message); err != nil {
				log.Error("  Failed to stash %s: %s", repoName, err)
				failureCount++
				continue
			}

			log.Success("  Successfully stashed changes in %s", repoName)
			successCount++
		}

		log.Info("Stash completed: %d successful, %d failed, %d skipped", successCount, failureCount, skippedCount)

		if failureCount > 0 {
			log.Warn("Some repositories failed to stash. Check the output above for details.")
		} else {
			log.Success("All git repositories with changes stashed successfully")
		}
	},
}

func init() {
	StashAllCmd.Flags().Bool("dry-run", false, "Perform a dry run without making actual changes")
	StashAllCmd.Flags().StringP("message", "m", "", "Stash message (optional)")
}

// stashRepository performs a git stash operation on the given repository path.
func stashRepository(repoPath, message string) error {
	args := []string{"-C", repoPath, "stash", "push"}
	if message != "" {
		args = append(args, "-m", message)
	}

	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error("Git stash output: %s", string(output))
		return err
	}
	log.Info("Git stash output: %s", string(output))
	return nil
}

// hasUncommittedChanges checks if the repository has uncommitted changes.
func hasUncommittedChanges(repoPath string) (bool, error) {
	cmd := exec.Command("git", "-C", repoPath, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	return len(output) > 0, nil
}
