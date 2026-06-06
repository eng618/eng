package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/repo"
	"github.com/eng618/eng/internal/ui"
)

// PullCmd defines the cobra command for pulling all project repositories.
// It runs git pull on all repositories in configured projects.
var PullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull updates for all project repositories",
	Long: `This command pulls the latest changes from remote for all repositories in configured projects.

Note: Repositories with uncommitted changes will be skipped.

Example:
  eng project pull                  # Pull all projects
  eng project pull -p MyProject     # Pull only the specified project
  eng project pull --dry-run        # Preview what would be pulled`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Pulling project repositories")

		isVerbose := cmdutil.IsVerbose(cmd)
		dryRun, _ := cmd.Parent().PersistentFlags().GetBool("dry-run")
		projectFilter, _ := cmd.Parent().PersistentFlags().GetString("project")

		devPath := viper.GetString("git.dev_path")
		if devPath == "" {
			log.Error("Development folder path is not set. Use 'eng config git-dev-path' to set it.")
			return
		}
		devPath = os.ExpandEnv(devPath)

		if dryRun {
			log.Info("Dry run mode - no actual git operations will be performed")
		}

		projects := config.GetProjects()
		if len(projects) == 0 {
			log.Warn("No projects configured. Use 'eng project add' to add a project.")
			return
		}

		// Filter by project if specified
		if projectFilter != "" {
			filtered := make([]config.Project, 0)
			for _, p := range projects {
				if p.Name == projectFilter {
					filtered = append(filtered, p)
					break
				}
			}
			if len(filtered) == 0 {
				log.Error("Project '%s' not found in configuration", projectFilter)
				return
			}
			projects = filtered
		}

		var successCount atomic.Int32
		var failedCount atomic.Int32
		var skippedCount atomic.Int32
		var dirtyCount atomic.Int32

		multi, err := ui.NewMultiSpinner()
		if err != nil {
			log.Error("Failed to initialize UI: %s", err)
			return
		}
		defer multi.Stop()

		var eg errgroup.Group
		eg.SetLimit(10) // Concurrent pull limit

		for _, project := range projects {
			projectPath := filepath.Join(devPath, project.Name)

			for _, r := range project.Repos {
				repoObj := r // capture loop variable
				eg.Go(func() error {
					repoPath, err := repoObj.GetEffectivePath()
					if err != nil {
						failedCount.Add(1)
						return nil
					}

					fullRepoPath := filepath.Join(projectPath, repoPath)

					// Check if repo exists
					if !isRepoCloned(fullRepoPath) {
						if isVerbose {
							spinner := multi.AddSpinner(fmt.Sprintf("Skipping %s (not cloned)", repoPath))
							spinner.Warning()
						}
						skippedCount.Add(1)
						return nil
					}

					if dryRun {
						spinner := multi.AddSpinner(fmt.Sprintf("[DRY RUN] Would pull: %s", repoPath))
						spinner.Success()
						successCount.Add(1)
						return nil
					}

					// Check for uncommitted changes
					isDirty, err := repo.IsDirty(fullRepoPath)
					if err != nil {
						spinner := multi.AddSpinner(fmt.Sprintf("Checking %s...", repoPath))
						spinner.Fail(fmt.Sprintf("Failed to check status of %s: %s", repoPath, err))
						failedCount.Add(1)
						return nil
					}

					if isDirty {
						spinner := multi.AddSpinner(fmt.Sprintf("Skipping %s (has uncommitted changes)", repoPath))
						spinner.Warning()
						dirtyCount.Add(1)
						return nil
					}

					spinner := multi.AddSpinner(fmt.Sprintf("Pulling %s...", repoPath))
					if err := repo.PullLatestCode(fullRepoPath); err != nil {
						// Check if it's just "already up to date"
						if errors.Is(err, git.NoErrAlreadyUpToDate) {
							spinner.Info(fmt.Sprintf("%s is already up to date", repoPath))
							successCount.Add(1)
							return nil
						}
						spinner.Fail(fmt.Sprintf("Failed to pull %s: %s", repoPath, err))
						failedCount.Add(1)
						return nil
					}

					spinner.Success(fmt.Sprintf("Pulled %s", repoPath))
					successCount.Add(1)
					return nil
				})
			}
		}

		_ = eg.Wait()
		multi.Stop() // explicitly stop to flush output before summary logs

		log.Info("")
		log.Info(
			"Pull complete: %d successful, %d skipped, %d dirty, %d failed",
			successCount.Load(),
			skippedCount.Load(),
			dirtyCount.Load(),
			failedCount.Load(),
		)

		if dirtyCount.Load() > 0 {
			log.Warn("Some repositories were skipped due to uncommitted changes.")
			log.Info("Commit or stash your changes, then run again.")
		}
	},
}
