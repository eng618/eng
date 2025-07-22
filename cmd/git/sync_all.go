package git

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/log"
	"github.com/eng618/eng/utils/repo"
	"github.com/spf13/cobra"
)

// SyncAllCmd defines the cobra command for syncing all git repositories.
// It fetches and pulls with rebase for all repositories in the development folder.
var SyncAllCmd = &cobra.Command{
	Use:   "sync-all",
	Short: "Sync all git repositories in development folder",
	Long:  `This command fetches and pulls with rebase for all git repositories found in your development folder.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Syncing all git repositories")

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
			log.Info("Processing repository: %s", repoName)

			if dryRun {
				log.Info("  [DRY RUN] Would sync repository at: %s", repoPath)
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

			// Pull latest code
			if err := repo.PullLatestCode(repoPath); err != nil {
				log.Error("  Failed to pull latest code: %s", err)
				failureCount++
				continue
			}

			log.Success("  Successfully synced %s", repoName)
			successCount++
		}

		log.Info("Sync completed: %d successful, %d failed", successCount, failureCount)

		if failureCount > 0 {
			log.Warn("Some repositories failed to sync. Check the output above for details.")
		} else {
			log.Success("All git repositories synced successfully")
		}
	},
}

func init() {
	SyncAllCmd.Flags().Bool("dry-run", false, "Perform a dry run without making actual changes")
}

// findGitRepositories scans the given directory for git repositories.
// It returns a slice of absolute paths to directories containing .git folders.
func findGitRepositories(devPath string) ([]string, error) {
	var repos []string

	// Check if the development path exists
	if _, err := os.Stat(devPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("development path does not exist: %s", devPath)
	}

	// Read the development directory
	entries, err := os.ReadDir(devPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read development directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		repoPath := filepath.Join(devPath, entry.Name())
		gitPath := filepath.Join(repoPath, ".git")

		// Check if .git exists (either as directory or file for worktrees)
		if _, err := os.Stat(gitPath); err == nil {
			repos = append(repos, repoPath)
		}
	}

	return repos, nil
}
