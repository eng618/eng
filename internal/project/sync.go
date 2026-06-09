package project

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync/atomic"

	"github.com/go-git/go-git/v5"
	"golang.org/x/sync/errgroup"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
)

// SyncOptions holds the configuration for syncing projects.
type SyncOptions struct {
	DryRun        bool
	IsVerbose     bool
	ProjectFilter string
	DevPath       string
	Projects      []config.Project
	RepoClient    RepoClient
}

// Sync fetches and pulls all repositories in configured projects.
func Sync(ctx context.Context, opts SyncOptions) {
	if opts.RepoClient == nil {
		opts.RepoClient = &defaultRepoClient{}
	}
	log.Start("Syncing project repositories")

	if opts.DryRun {
		log.Info("Dry run mode - no actual git operations will be performed")
	}

	projects := opts.Projects
	if len(projects) == 0 {
		log.Warn("No projects configured. Use 'eng project add' to add a project.")
		return
	}

	// Filter by project if specified
	projects = filterProjects(projects, opts.ProjectFilter)
	if len(projects) == 0 {
		return
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

	if ctx == nil {
		ctx = context.Background()
	}
	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(10) // Concurrent sync limit

	for _, project := range projects {
		projectPath := filepath.Join(opts.DevPath, project.Name)

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
					if opts.IsVerbose {
						spinner := multi.AddSpinner(fmt.Sprintf("Skipping %s (not cloned)", repoPath))
						spinner.Warning()
					}
					skippedCount.Add(1)
					return nil
				}

				if opts.DryRun {
					spinner := multi.AddSpinner(fmt.Sprintf("[DRY RUN] Would sync: %s", repoPath))
					spinner.Success()
					fetchSuccess.Add(1)
					pullSuccess.Add(1)
					return nil
				}

				spinner := multi.AddSpinner(fmt.Sprintf("Syncing %s...", repoPath))

				// Fetch
				spinner.UpdateText(fmt.Sprintf("Fetching %s...", repoPath))
				if err := opts.RepoClient.FetchAllPrune(egCtx, fullRepoPath); err != nil {
					spinner.Fail(fmt.Sprintf("Fetch failed for %s: %s", repoPath, err))
					fetchFailed.Add(1)
					// don't pull if fetch failed, but log it as pull failed too
					pullFailed.Add(1)
					return nil
				}
				fetchSuccess.Add(1)

				// Check for uncommitted changes before pull
				spinner.UpdateText(fmt.Sprintf("Checking %s...", repoPath))
				isDirty, err := opts.RepoClient.IsDirty(egCtx, fullRepoPath)
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
				if err := opts.RepoClient.PullLatestCode(egCtx, fullRepoPath); err != nil {
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
}
