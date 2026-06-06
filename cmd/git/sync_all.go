package git

import (
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/repo"
	"github.com/eng618/eng/internal/ui"
)

// SyncAllCmd defines the cobra command for syncing all git repositories.
// It fetches and pulls with rebase for all repositories in the development folder.
var SyncAllCmd = &cobra.Command{
	Use:   "sync-all",
	Short: "Sync all git repositories in development folder",
	Long:  `This command fetches and pulls with rebase for all git repositories found in your development folder.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Syncing all git repositories")

		isVerbose := cmdutil.IsVerbose(cmd)
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

		var successCount atomic.Int32
		var failureCount atomic.Int32

		multi, err := ui.NewMultiSpinner()
		if err != nil {
			log.Error("Failed to initialize UI: %s", err)
			return
		}
		defer multi.Stop()

		var eg errgroup.Group
		eg.SetLimit(10) // Concurrent sync limit

		for _, repoPath := range repos {
			rPath := repoPath // capture loop variable
			eg.Go(func() error {
				repoName := filepath.Base(rPath)

				if dryRun {
					spinner := multi.AddSpinner(fmt.Sprintf("[DRY RUN] Would sync repository at: %s", rPath))
					spinner.Success()
					successCount.Add(1)
					return nil
				}

				spinner := multi.AddSpinner(fmt.Sprintf("Processing %s...", repoName))

				// Check if repository is dirty
				isDirty, err := repo.IsDirty(cmd.Context(), rPath)
				if err != nil {
					spinner.Fail(fmt.Sprintf("Failed to check status for %s: %s", repoName, err))
					failureCount.Add(1)
					return nil
				}

				if isDirty {
					spinner.Warning(fmt.Sprintf("Repository %s has uncommitted changes, skipping...", repoName))
					failureCount.Add(1)
					return nil
				}

				// Ensure we're on default branch
				if err := repo.EnsureOnDefaultBranch(cmd.Context(), rPath); err != nil {
					spinner.Fail(fmt.Sprintf("Not on default branch for %s: %s", repoName, err))
					failureCount.Add(1)
					return nil
				}

				// Pull latest code
				spinner.UpdateText(fmt.Sprintf("Pulling %s...", repoName))
				if err := repo.PullLatestCode(cmd.Context(), rPath); err != nil {
					spinner.Fail(fmt.Sprintf("Failed to pull latest code for %s: %s", repoName, err))
					failureCount.Add(1)
					return nil
				}

				spinner.Success(fmt.Sprintf("Successfully synced %s", repoName))
				successCount.Add(1)
				return nil
			})
		}

		_ = eg.Wait()
		multi.Stop()

		log.Info("Sync completed: %d successful, %d failed", successCount.Load(), failureCount.Load())

		if failureCount.Load() > 0 {
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
