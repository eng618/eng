package git

import (
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/log"
	"github.com/eng618/eng/utils/repo"
)

// PullAllCmd defines the cobra command for pulling all git repositories.
// It pulls with rebase for all repositories in the development folder (assumes fetch was already done).
var PullAllCmd = &cobra.Command{
	Use:   "pull-all",
	Short: "Pull all git repositories in development folder",
	Long:  `This command pulls with rebase for all git repositories found in your development folder. Use this after fetch-all for faster operations.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Pulling all git repositories")

		isVerbose := utils.IsVerbose(cmd)
		dryRun, _ := cmd.Flags().GetBool("dry-run")

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

		for _, repoPath := range repos {
			repoName := filepath.Base(repoPath)
			log.Info("Pulling repository: %s", repoName)

			if dryRun {
				log.Info("  [DRY RUN] Would pull repository at: %s", repoPath)
				successCount++
				continue
			}

			// Check if repository is dirty
			isDirty, err := repo.IsDirty(repoPath)
			if err != nil {
				log.Error("  Failed to check repository status: %s", err)
				failureCount++
				continue
			}

			if isDirty {
				log.Warn("  Repository has uncommitted changes, skipping...")
				failureCount++
				continue
			}

			// Ensure we're on default branch
			if err := repo.EnsureOnDefaultBranch(repoPath); err != nil {
				log.Error("  Failed to ensure on default branch: %s", err)
				failureCount++
				continue
			}

			// Pull with rebase
			if err := pullRepository(repoPath); err != nil {
				log.Error("  Failed to pull %s: %s", repoName, err)
				failureCount++
				continue
			}

			log.Success("  Successfully pulled %s", repoName)
			successCount++
		}

		log.Info("Pull completed: %d successful, %d failed", successCount, failureCount)

		if failureCount > 0 {
			log.Warn("Some repositories failed to pull. Check the output above for details.")
		} else {
			log.Success("All git repositories pulled successfully")
		}
	},
}

func init() {
	PullAllCmd.Flags().Bool("dry-run", false, "Perform a dry run without making actual changes")
}

// pullRepository performs a git pull with rebase operation on the given repository path.
func pullRepository(repoPath string) error {
	cmd := exec.Command("git", "-C", repoPath, "pull", "--rebase", "--autostash")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error("Git pull output: %s", string(output))
		return err
	}
	log.Info("Git pull output: %s", string(output))
	return nil
}
