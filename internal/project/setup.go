package project

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/repo"
)

// SetupOptions holds the configuration for setting up projects.
type SetupOptions struct {
	DryRun        bool
	IsVerbose     bool
	ProjectFilter string
}

// Setup ensures project directories exist and clones missing repositories.
func Setup(opts SetupOptions) {
	log.Start("Setting up project repositories")

	devPath := viper.GetString("git.dev_path")
	if devPath == "" {
		log.Error("Development folder path is not set. Use 'eng config git-dev-path' to set it.")
		return
	}
	devPath = os.ExpandEnv(devPath)

	log.Verbose(opts.IsVerbose, "Development path: %s", devPath)

	if opts.DryRun {
		log.Info("Dry run mode - no actual changes will be made")
	}

	projects := config.GetProjects()
	if len(projects) == 0 {
		log.Warn("No projects configured. Use 'eng project add' to add a project.")
		return
	}

	// Filter by project if specified
	if opts.ProjectFilter != "" {
		filtered := make([]config.Project, 0)
		for _, p := range projects {
			if p.Name == opts.ProjectFilter {
				filtered = append(filtered, p)
				break
			}
		}
		if len(filtered) == 0 {
			log.Error("Project '%s' not found in configuration", opts.ProjectFilter)
			return
		}
		projects = filtered
	}

	totalRepos := 0
	clonedCount := 0
	skippedCount := 0
	failedCount := 0

	for _, p := range projects {
		log.Info("Processing project: %s", p.Name)

		projectPath := filepath.Join(devPath, p.Name)

		// Ensure project directory exists
		if opts.DryRun {
			log.Info("  [DRY RUN] Would ensure directory exists: %s", projectPath)
		} else {
			if err := os.MkdirAll(projectPath, 0o755); err != nil {
				log.Error("  Failed to create project directory: %s", err)
				continue
			}
			log.Verbose(opts.IsVerbose, "  Project directory ready: %s", projectPath)
		}

		for _, projectRepo := range p.Repos {
			totalRepos++

			repoPath, err := projectRepo.GetEffectivePath()
			if err != nil {
				log.Error("  Failed to determine path for %s: %s", projectRepo.URL, err)
				failedCount++
				continue
			}

			fullRepoPath := filepath.Join(projectPath, repoPath)

			// Check if repo already exists
			if _, err := os.Stat(filepath.Join(fullRepoPath, ".git")); err == nil {
				log.Verbose(opts.IsVerbose, "  Repository already exists: %s", repoPath)
				skippedCount++
				continue
			}

			if opts.DryRun {
				log.Info("  [DRY RUN] Would clone %s to %s", projectRepo.URL, fullRepoPath)
				clonedCount++
				continue
			}

			log.Info("  Cloning %s...", repoPath)

			// Clone the repository
			if err := repo.Clone(projectRepo.URL, fullRepoPath); err != nil {
				log.Error("  Failed to clone %s: %s", projectRepo.URL, err)
				failedCount++
				continue
			}

			log.Success("  Cloned %s", repoPath)
			clonedCount++
		}
	}

	log.Info("")
	log.Info("Setup complete:")
	log.Info("  Total repositories: %d", totalRepos)
	log.Info("  Cloned: %d", clonedCount)
	log.Info("  Already present: %d", skippedCount)
	if failedCount > 0 {
		log.Warn("  Failed: %d", failedCount)
	}

	if failedCount > 0 {
		log.Warn("Some repositories failed to clone. Check the output above for details.")
		log.Info("Common issues:")
		log.Info("  - SSH key not configured for the repository host")
		log.Info("  - Repository URL is incorrect")
		log.Info("  - Network connectivity issues")
	} else if !opts.DryRun && clonedCount > 0 {
		log.Success("All project repositories set up successfully!")
	}
}
