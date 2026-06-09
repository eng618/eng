package project

import (
	"context"
	"os"
	"path/filepath"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
)

// SetupOptions holds the configuration for setting up projects.
type SetupOptions struct {
	DryRun        bool
	IsVerbose     bool
	ProjectFilter string
	DevPath       string
	Projects      []config.Project
	RepoClient    RepoClient
}

// SetupStats tracks the results of the setup operation.
type SetupStats struct {
	TotalRepos   int
	ClonedCount  int
	SkippedCount int
	FailedCount  int
	FailedRepos  []string
}

// Setup ensures project directories exist and clones missing repositories.
func Setup(ctx context.Context, opts SetupOptions) {
	if opts.RepoClient == nil {
		opts.RepoClient = &defaultRepoClient{}
	}
	log.Start("Setting up project repositories")

	if opts.DryRun {
		log.Info("Dry run mode - no actual changes will be made")
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

	stats := &SetupStats{}
	for _, p := range projects {
		setupProject(ctx, p, opts.DevPath, opts, stats)
	}
	printSummary(stats, opts.DryRun)
}

func filterProjects(projects []config.Project, filter string) []config.Project {
	if filter == "" {
		return projects
	}

	for _, p := range projects {
		if p.Name == filter {
			return []config.Project{p}
		}
	}

	log.Error("Project '%s' not found in configuration", filter)
	return []config.Project{}
}

func setupProject(ctx context.Context, p config.Project, devPath string, opts SetupOptions, stats *SetupStats) {
	log.Info("Processing project: %s", p.Name)

	projectPath := filepath.Join(devPath, p.Name)

	if opts.DryRun {
		log.Info("  [DRY RUN] Would ensure directory exists: %s", projectPath)
	} else {
		if err := os.MkdirAll(projectPath, 0o755); err != nil {
			log.Error("  Failed to create project directory: %s", err)
			return
		}
		log.Verbose(opts.IsVerbose, "  Project directory ready: %s", projectPath)
	}

	for _, projectRepo := range p.Repos {
		setupRepo(ctx, projectRepo, projectPath, opts, stats)
	}
}

func setupRepo(
	ctx context.Context,
	projectRepo config.ProjectRepo,
	projectPath string,
	opts SetupOptions,
	stats *SetupStats,
) {
	stats.TotalRepos++

	repoPath, err := projectRepo.GetEffectivePath()
	if err != nil {
		log.Error("  Failed to determine path for %s: %s", projectRepo.URL, err)
		stats.FailedCount++
		stats.FailedRepos = append(stats.FailedRepos, repoPath)
		return
	}

	fullRepoPath := filepath.Join(projectPath, repoPath)

	if _, err := os.Stat(filepath.Join(fullRepoPath, ".git")); err == nil {
		log.Verbose(opts.IsVerbose, "  Repository already exists: %s", repoPath)
		stats.SkippedCount++
		return
	}

	if opts.DryRun {
		log.Info("  [DRY RUN] Would clone %s to %s", projectRepo.URL, fullRepoPath)
		stats.ClonedCount++
		return
	}

	log.Info("  Cloning %s...", repoPath)

	if err := opts.RepoClient.Clone(ctx, projectRepo.URL, fullRepoPath); err != nil {
		log.Error("  Failed to clone %s: %s", projectRepo.URL, err)
		stats.FailedCount++
		stats.FailedRepos = append(stats.FailedRepos, repoPath)
		return
	}

	log.Success("  Cloned %s", repoPath)
	stats.ClonedCount++
}

func printSummary(stats *SetupStats, dryRun bool) {
	log.Info("")
	log.Info("Setup complete:")
	log.Info("  Total repositories: %d", stats.TotalRepos)
	log.Info("  Cloned: %d", stats.ClonedCount)
	log.Info("  Already present: %d", stats.SkippedCount)
	if stats.FailedCount > 0 {
		log.Error("  Failed: %d", stats.FailedCount)
		log.Error("Some repositories failed to clone (require manual resolution):")
		for _, r := range stats.FailedRepos {
			log.Error("  - %s", r)
		}
		log.Info("Common issues:")
		log.Info("  - SSH key not configured for the repository host")
		log.Info("  - Repository URL is incorrect")
		log.Info("  - Network connectivity issues")
	} else if !dryRun && stats.ClonedCount > 0 {
		log.Success("All project repositories set up successfully!")
	}
}
