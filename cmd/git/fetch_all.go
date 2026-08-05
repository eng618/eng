package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"sync/atomic"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

// FetchAllCmd defines the cobra command for fetching all git repositories.
// It fetches updates from remote for all repositories in the development folder.
var FetchAllCmd = &cobra.Command{
	Use:   "fetch-all",
	Short: "Fetch all git repositories in development folder",
	Long:  `This command fetches updates from remote for all git repositories found in your development folder.`,
	Run: func(cmd *cobra.Command, args []string) {
		printHeader("🔍 Fetching Git Repositories")

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

		var successCount atomic.Int32
		var failureCount atomic.Int32

		multi, err := ui.NewMultiSpinner()
		if err != nil {
			log.Error("Failed to initialize UI: %s", err)
			return
		}
		defer multi.Stop()

		var eg errgroup.Group
		eg.SetLimit(10) // Concurrent fetch limit

		for _, repoPath := range repos {
			rPath := repoPath // capture loop variable
			eg.Go(func() error {
				repoName := filepath.Base(rPath)

				if setup.DryRun {
					spinner := multi.AddSpinner(fmt.Sprintf("[DRY RUN] Would fetch repository at: %s", rPath))
					spinner.Success()
					successCount.Add(1)
					return nil
				}

				spinner := multi.AddSpinner(fmt.Sprintf("Fetching %s...", repoName))

				// Perform git fetch
				if err := fetchRepository(rPath); err != nil {
					spinner.Fail(fmt.Sprintf("Failed to fetch %s: %s", repoName, err))
					failureCount.Add(1)
					return nil
				}

				spinner.Success(fmt.Sprintf("Fetched %s", repoName))
				successCount.Add(1)
				return nil
			})
		}

		_ = eg.Wait()
		multi.Stop()

		summaryMsg := fmt.Sprintf(
			"Fetch completed: %d successful, %d failed across %d repositories.",
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
	FetchAllCmd.Flags().Bool("dry-run", false, "Perform a dry run without making actual changes")
}

func fetchRepository(repoPath string) error {
	cmd := exec.Command("git", "-C", repoPath, "fetch", "--all", "--prune")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}
	return nil
}
