package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"sync/atomic"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/repo"
	"github.com/eng618/eng/internal/ui"
)

// PullAllCmd defines the cobra command for pulling all git repositories.
// It pulls with rebase for all repositories in the development folder (assumes fetch was already done).
var PullAllCmd = &cobra.Command{
	Use:   "pull-all",
	Short: "Pull all git repositories in development folder",
	Long:  `This command pulls with rebase for all git repositories found in your development folder. Use this after fetch-all for faster operations.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Pulling all git repositories")

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
		eg.SetLimit(10) // Concurrent pull limit

		for _, repoPath := range repos {
			rPath := repoPath // capture loop variable
			eg.Go(func() error {
				repoName := filepath.Base(rPath)

				if dryRun {
					spinner := multi.AddSpinner(fmt.Sprintf("[DRY RUN] Would pull repository at: %s", rPath))
					spinner.Success()
					successCount.Add(1)
					return nil
				}

				spinner := multi.AddSpinner(fmt.Sprintf("Pulling %s...", repoName))

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

				// Pull with rebase
				if err := pullRepository(rPath); err != nil {
					spinner.Fail(fmt.Sprintf("Failed to pull %s: %s", repoName, err))
					failureCount.Add(1)
					return nil
				}

				spinner.Success(fmt.Sprintf("Successfully pulled %s", repoName))
				successCount.Add(1)
				return nil
			})
		}

		_ = eg.Wait()
		multi.Stop()

		log.Info("Pull completed: %d successful, %d failed", successCount.Load(), failureCount.Load())

		if failureCount.Load() > 0 {
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
		return fmt.Errorf("%w: %s", err, string(output))
	}
	return nil
}
