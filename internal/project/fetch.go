package project

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"

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
}

// Fetch ensures project repositories are fetched.
func Fetch(ctx context.Context, opts FetchOptions) {
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

	var successCount atomic.Int32
	var failedCount atomic.Int32
	var skippedCount atomic.Int32

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
					failedCount.Add(1)
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
					spinner := multi.AddSpinner(fmt.Sprintf("[DRY RUN] Would fetch: %s", repoPath))
					spinner.Success()
					successCount.Add(1)
					return nil
				}

				spinner := multi.AddSpinner(fmt.Sprintf("Fetching %s...", repoPath))
				if err := fetchRepo(egCtx, fullRepoPath); err != nil {
					spinner.Fail(fmt.Sprintf("Failed to fetch %s: %s", repoPath, err))
					failedCount.Add(1)
					return nil
				}

				spinner.Success(fmt.Sprintf("Fetched %s", repoPath))
				successCount.Add(1)
				return nil
			})
		}
	}

	_ = eg.Wait()
	multi.Stop() // explicitly stop to flush output before summary logs

	log.Info("")
	log.Info(
		"Fetch complete: %d successful, %d skipped, %d failed",
		successCount.Load(),
		skippedCount.Load(),
		failedCount.Load(),
	)
}

func isRepoCloned(repoPath string) bool {
	info, err := os.Stat(filepath.Join(repoPath, ".git"))
	return err == nil && info.IsDir()
}

func fetchRepo(ctx context.Context, repoPath string) error {
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "fetch", "--all", "--prune")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(out))
	}
	return nil
}
