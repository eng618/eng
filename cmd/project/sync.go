package project

import (
	"context"
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

// SyncCmd defines the cobra command for syncing all project repositories.
// It fetches and pulls all repositories in configured projects.
var SyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync all project repositories (fetch + pull)",
	Long: `This command synchronizes all repositories in configured projects by:
  1. Fetching updates from remote (git fetch --all --prune)
  2. Pulling changes for the current branch (git pull)

Repositories with uncommitted changes will have fetch performed but pull will be skipped.

Example:
  eng project sync                  # Sync all projects
  eng project sync -p MyProject     # Sync only the specified project
  eng project sync --dry-run        # Preview what would be synced`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Syncing project repositories")

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

		var fetchSuccess atomic.Int32
		var fetchFailed atomic.Int32
		var pullSuccess atomic.Int32
		var pullFailed atomic.Int32
		var skippedCount atomic.Int32
		var dirtyCount atomic.Int32

		multi, err := ui.NewMultiSpinner()
		if err != nil {
			log.Error("Failed to initialize UI: %s", err)
			return
		}
		defer multi.Stop()

		cmdCtx := cmd.Context()
		if cmdCtx == nil {
			cmdCtx = context.Background()
		}
		eg, ctx := errgroup.WithContext(cmdCtx)

		eg.SetLimit(10) // Concurrent sync limit

		for _, project := range projects {
			projectPath := filepath.Join(devPath, project.Name)

			for _, r := range project.Repos {
				repoObj := r // capture loop variable
				eg.Go(func() error {
					repoPath, err := repoObj.GetEffectivePath()
					if err != nil {
						fetchFailed.Add(1)
						pullFailed.Add(1)
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
						spinner := multi.AddSpinner(fmt.Sprintf("[DRY RUN] Would sync: %s", repoPath))
						spinner.Success()
						fetchSuccess.Add(1)
						pullSuccess.Add(1)
						return nil
					}

					spinner := multi.AddSpinner(fmt.Sprintf("Syncing %s...", repoPath))

					// Fetch
					spinner.UpdateText(fmt.Sprintf("Fetching %s...", repoPath))
					if err := fetchRepo(fullRepoPath); err != nil {
						spinner.Fail(fmt.Sprintf("Fetch failed for %s: %s", repoPath, err))
						fetchFailed.Add(1)
						// don't pull if fetch failed, but log it as pull failed too
						pullFailed.Add(1)
						return nil
					}
					fetchSuccess.Add(1)

					// Check for uncommitted changes before pull
					spinner.UpdateText(fmt.Sprintf("Checking %s...", repoPath))
					isDirty, err := repo.IsDirty(ctx, fullRepoPath)
					if err != nil {
						spinner.Fail(fmt.Sprintf("Failed to check status for %s: %s", repoPath, err))
						pullFailed.Add(1)
						return nil
					}

					if isDirty {
						spinner.Warning(fmt.Sprintf("Skipped pull for %s (has uncommitted changes)", repoPath))
						dirtyCount.Add(1)
						return nil
					}

					// Pull
					spinner.UpdateText(fmt.Sprintf("Pulling %s...", repoPath))
					if err := repo.PullLatestCode(ctx, fullRepoPath); err != nil {
						if errors.Is(err, git.NoErrAlreadyUpToDate) {
							spinner.Info(fmt.Sprintf("Synced %s (already up to date)", repoPath))
							pullSuccess.Add(1)
							return nil
						}
						spinner.Fail(fmt.Sprintf("Pull failed for %s: %s", repoPath, err))
						pullFailed.Add(1)
						return nil
					}

					spinner.Success(fmt.Sprintf("Synced %s", repoPath))
					pullSuccess.Add(1)
					return nil
				})
			}
		}

		_ = eg.Wait()
		multi.Stop() // explicitly flush

		log.Info("")
		log.Info("Sync complete:")
		log.Info("  Fetch: %d successful, %d failed", fetchSuccess.Load(), fetchFailed.Load())
		log.Info(
			"  Pull:  %d successful, %d failed, %d dirty, %d skipped",
			pullSuccess.Load(),
			pullFailed.Load(),
			dirtyCount.Load(),
			skippedCount.Load(),
		)

		if dirtyCount.Load() > 0 {
			log.Warn("Some repositories were not pulled due to uncommitted changes.")
		}
	},
}
