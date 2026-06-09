package project

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"

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

	var mu sync.Mutex
	var fetchSuccess, fetchFailed, pullSuccess, pullFailed, skippedCount, dirtyCount int
	var fetchFailedRepos, pullFailedRepos, skippedRepos, dirtyRepos []string

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
					mu.Lock()
					fetchFailed++
					pullFailed++
					fetchFailedRepos = append(fetchFailedRepos, repoPath)
					pullFailedRepos = append(pullFailedRepos, repoPath)
					mu.Unlock()
					return nil
				}

				fullRepoPath := filepath.Join(projectPath, repoPath)

				// Check if repo exists
				if !isRepoCloned(fullRepoPath) {
					if opts.IsVerbose {
						spinner := multi.AddSpinner(fmt.Sprintf("Skipping %s (not cloned)", repoPath))
						spinner.Warning()
					}
					mu.Lock()
					skippedCount++
					skippedRepos = append(skippedRepos, repoPath)
					mu.Unlock()
					return nil
				}

				if opts.DryRun {
					spinner := multi.AddSpinner(fmt.Sprintf("[DRY RUN] Would sync: %s", repoPath))
					spinner.Success()
					mu.Lock()
					fetchSuccess++
					mu.Unlock()
					mu.Lock()
					pullSuccess++
					mu.Unlock()
					return nil
				}

				spinner := multi.AddSpinner(fmt.Sprintf("Syncing %s...", repoPath))

				// Fetch
				spinner.UpdateText(fmt.Sprintf("Fetching %s...", repoPath))
				if err := opts.RepoClient.FetchAllPrune(egCtx, fullRepoPath); err != nil {
					spinner.Fail(fmt.Sprintf("Fetch failed for %s: %s", repoPath, err))
					mu.Lock()
					fetchFailed++
					fetchFailedRepos = append(fetchFailedRepos, repoPath)
					mu.Unlock()
					// don't pull if fetch failed, but log it as pull failed too
					mu.Lock()
					pullFailed++
					pullFailedRepos = append(pullFailedRepos, repoPath)
					mu.Unlock()
					return nil
				}
				mu.Lock()
				fetchSuccess++
				mu.Unlock()

				// Check for uncommitted changes before pull
				spinner.UpdateText(fmt.Sprintf("Checking %s...", repoPath))
				isDirty, err := opts.RepoClient.IsDirty(egCtx, fullRepoPath)
				if err != nil {
					spinner.Fail(fmt.Sprintf("Failed to check status for %s: %s", repoPath, err))
					mu.Lock()
					pullFailed++
					pullFailedRepos = append(pullFailedRepos, repoPath)
					mu.Unlock()
					return nil
				}

				if isDirty {
					spinner.Warning(fmt.Sprintf("Skipped pull for %s (has uncommitted changes)", repoPath))
					mu.Lock()
					dirtyCount++
					dirtyRepos = append(dirtyRepos, repoPath)
					mu.Unlock()
					return nil
				}

				// Pull
				spinner.UpdateText(fmt.Sprintf("Pulling %s...", repoPath))
				if err := opts.RepoClient.PullLatestCode(egCtx, fullRepoPath); err != nil {
					if errors.Is(err, git.NoErrAlreadyUpToDate) {
						spinner.Info(fmt.Sprintf("Synced %s (already up to date)", repoPath))
						mu.Lock()
						pullSuccess++
						mu.Unlock()
						return nil
					}
					spinner.Fail(fmt.Sprintf("Pull failed for %s: %s", repoPath, err))
					mu.Lock()
					pullFailed++
					pullFailedRepos = append(pullFailedRepos, repoPath)
					mu.Unlock()
					return nil
				}

				spinner.Success(fmt.Sprintf("Synced %s", repoPath))
				mu.Lock()
				pullSuccess++
				mu.Unlock()
				return nil
			})
		}
	}

	_ = eg.Wait()
	multi.Stop() // explicitly flush

	log.Info("")
	log.Info("Sync complete:")
	log.Info("  Fetch: %d successful, %d failed", fetchSuccess, fetchFailed)
	log.Info(
		"  Pull:  %d successful, %d failed, %d dirty, %d skipped",
		pullSuccess,
		pullFailed,
		dirtyCount,
		skippedCount,
	)

	if len(dirtyRepos) > 0 {
		log.Warn("Dirty repositories (skipped pull, require manual commit/stash):")
		for _, r := range dirtyRepos {
			log.Warn("  - %s", r)
		}
	}
	if len(skippedRepos) > 0 {
		log.Warn("Skipped repositories (not cloned):")
		for _, r := range skippedRepos {
			log.Warn("  - %s", r)
		}
	}
	if len(fetchFailedRepos) > 0 || len(pullFailedRepos) > 0 {
		log.Error("Failed repositories (require manual resolution):")

		// We'll build a unique list to avoid printing duplicates
		failedSet := make(map[string]bool)
		for _, r := range fetchFailedRepos {
			failedSet[r] = true
		}
		for _, r := range pullFailedRepos {
			failedSet[r] = true
		}

		for r := range failedSet {
			log.Error("  - %s", r)
		}
	}
}
