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

// PullOptions holds the configuration for pulling projects.
type PullOptions struct {
	DryRun        bool
	IsVerbose     bool
	ProjectFilter string
	DevPath       string
	Projects      []config.Project
	RepoClient    RepoClient
}

// Pull updates from remote for all repositories in configured projects.
func Pull(ctx context.Context, opts PullOptions) {
	if opts.RepoClient == nil {
		opts.RepoClient = &defaultRepoClient{}
	}
	log.Start("Pulling project repositories")

	if opts.DryRun {
		log.Info("Dry run mode - no actual git operations will be performed")
	}

	projects := opts.Projects
	if len(projects) == 0 {
		log.Warn("No projects configured. Use 'eng project add' to add a project.")
		return
	}

	projects = filterProjects(projects, opts.ProjectFilter)
	if len(projects) == 0 {
		return
	}

	var mu sync.Mutex
	var successCount, failedCount, skippedCount, dirtyCount int
	var failedRepos, skippedRepos, dirtyRepos []string

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
	eg.SetLimit(10) // Concurrent pull limit

	for _, project := range projects {
		projectPath := filepath.Join(opts.DevPath, project.Name)

		for _, r := range project.Repos {
			repoObj := r // capture loop variable
			eg.Go(func() error {
				repoPath, err := repoObj.GetEffectivePath()
				if err != nil {
					mu.Lock()
					failedCount++
					failedRepos = append(failedRepos, repoPath)
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
					spinner := multi.AddSpinner(fmt.Sprintf("[DRY RUN] Would pull: %s", repoPath))
					spinner.Success()
					mu.Lock()
					successCount++
					mu.Unlock()
					return nil
				}

				// Check for uncommitted changes
				isDirty, err := opts.RepoClient.IsDirty(egCtx, fullRepoPath)
				if err != nil {
					spinner := multi.AddSpinner(fmt.Sprintf("Checking %s...", repoPath))
					spinner.Fail(fmt.Sprintf("Failed to check status of %s: %s", repoPath, err))
					mu.Lock()
					failedCount++
					failedRepos = append(failedRepos, repoPath)
					mu.Unlock()
					return nil
				}

				if isDirty {
					spinner := multi.AddSpinner(fmt.Sprintf("Skipping %s (has uncommitted changes)", repoPath))
					spinner.Warning()
					mu.Lock()
					dirtyCount++
					dirtyRepos = append(dirtyRepos, repoPath)
					mu.Unlock()
					return nil
				}

				spinner := multi.AddSpinner(fmt.Sprintf("Pulling %s...", repoPath))
				if err := opts.RepoClient.PullLatestCode(egCtx, fullRepoPath); err != nil {
					// Check if it's just "already up to date"
					if errors.Is(err, git.NoErrAlreadyUpToDate) {
						spinner.Info(fmt.Sprintf("%s is already up to date", repoPath))
						mu.Lock()
						successCount++
						mu.Unlock()
						return nil
					}
					spinner.Fail(fmt.Sprintf("Failed to pull %s: %s", repoPath, err))
					mu.Lock()
					failedCount++
					failedRepos = append(failedRepos, repoPath)
					mu.Unlock()
					return nil
				}

				spinner.Success(fmt.Sprintf("Pulled %s", repoPath))
				mu.Lock()
				successCount++
				mu.Unlock()
				return nil
			})
		}
	}

	_ = eg.Wait()
	multi.Stop() // explicitly stop to flush output before summary logs

	log.Info("")
	log.Info(
		"Pull complete: %d successful, %d skipped, %d dirty, %d failed",
		successCount,
		skippedCount,
		dirtyCount,
		failedCount,
	)

	if len(dirtyRepos) > 0 {
		log.Warn("Dirty repositories (skipped, require manual commit/stash):")
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
	if len(failedRepos) > 0 {
		log.Error("Failed repositories (require manual resolution):")
		for _, r := range failedRepos {
			log.Error("  - %s", r)
		}
	}
}
