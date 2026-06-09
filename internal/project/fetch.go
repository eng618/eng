package project

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
)

// FetchOptions holds the configuration for fetching projects.
type FetchOptions struct {
	DryRun        bool
	IsVerbose     bool
	ProjectFilter string
	DevPath       string
	Projects      []config.Project
	RepoClient    RepoClient
}

// Fetch ensures project repositories are fetched.
func Fetch(ctx context.Context, opts FetchOptions) {
	if opts.RepoClient == nil {
		opts.RepoClient = &defaultRepoClient{}
	}
	log.Start("Fetching project repositories")

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
	var successCount, failedCount, skippedCount int
	var failedRepos, skippedRepos []string

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
	eg.SetLimit(10) // Concurrent fetch limit

	for _, project := range projects {
		projectPath := filepath.Join(opts.DevPath, project.Name)

		for _, repo := range project.Repos {
			r := repo // explicitly capture loop variable for closure

			eg.Go(func() error {
				repoPath, err := r.GetEffectivePath()
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
					spinner := multi.AddSpinner(fmt.Sprintf("[DRY RUN] Would fetch: %s", repoPath))
					spinner.Success()
					mu.Lock()
					successCount++
					mu.Unlock()
					return nil
				}

				spinner := multi.AddSpinner(fmt.Sprintf("Fetching %s...", repoPath))
				if err := opts.RepoClient.FetchAllPrune(egCtx, fullRepoPath); err != nil {
					spinner.Fail(fmt.Sprintf("Failed to fetch %s: %s", repoPath, err))
					mu.Lock()
					failedCount++
					failedRepos = append(failedRepos, repoPath)
					mu.Unlock()
					return nil
				}

				spinner.Success(fmt.Sprintf("Fetched %s", repoPath))
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
	log.Info("Fetch complete: %d successful, %d skipped, %d failed", successCount, skippedCount, failedCount)

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

func isRepoCloned(repoPath string) bool {
	info, err := os.Stat(filepath.Join(repoPath, ".git"))
	return err == nil && info.IsDir()
}
