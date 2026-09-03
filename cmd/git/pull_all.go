package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"sync/atomic"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/repo"
	"github.com/eng618/eng/internal/runlog"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

// PullAllCmd defines the cobra command for pulling all git repositories.
// It pulls with rebase for all repositories in the development folder (assumes fetch was already done).
var PullAllCmd = &cobra.Command{
	Use:   "pull-all",
	Short: "Pull all git repositories in development folder",
	Long:  `This command pulls with rebase for all git repositories found in your development folder. Use this after fetch-all for faster operations.`,
	Run: func(cmd *cobra.Command, args []string) {
		printHeader("📥 Pulling Git Repositories")

		setup, err := setupGitCommand(cmd)
		if err != nil {
			log.Error("%s", err)
			return
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

		logPath, stopLog := runlog.Start("git-pull-all")
		defer stopLog()
		defer runlog.Finish(logPath)

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

				if setup.DryRun {
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

		summaryMsg := fmt.Sprintf(
			"Pull completed: %d successful, %d failed across %d repositories.",
			successCount.Load(),
			failureCount.Load(),
			len(repos),
		)
		if failureCount.Load() > 0 {
			theme.WarningMessage(summaryMsg)
		} else {
			theme.SuccessMessage(summaryMsg)
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
