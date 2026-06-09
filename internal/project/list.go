package project

import (
	"os"
	"path/filepath"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
)

// ListOptions holds the configuration for listing projects.
type ListOptions struct {
	IsVerbose     bool
	ProjectFilter string
	DevPath       string
	Projects      []config.Project
	RepoClient    RepoClient
}

// List displays project names, repository counts, and clone status.
func List(opts ListOptions) {
	if opts.RepoClient == nil {
		opts.RepoClient = &defaultRepoClient{}
	}
	devPath := opts.DevPath
	if devPath == "" {
		log.Warn("Development folder path is not set. Use 'eng config git-dev-path' to set it.")
		devPath = "(not configured)"
	} else {
		devPath = os.ExpandEnv(devPath)
	}

	projects := opts.Projects
	if len(projects) == 0 {
		log.Info("No projects configured.")
		log.Info("Use 'eng project add' to add a project.")
		return
	}

	// Filter by project if specified
	projects = filterProjects(projects, opts.ProjectFilter)
	if len(projects) == 0 {
		return
	}

	log.Info("Development path: %s", devPath)
	log.Info("")

	for _, project := range projects {
		projectPath := filepath.Join(devPath, project.Name)

		if opts.IsVerbose {
			log.Info("Project: %s", project.Name)
			log.Info("  Path: %s", projectPath)
			log.Info("  Repositories (%d):", len(project.Repos))

			for _, repo := range project.Repos {
				repoPath, err := repo.GetEffectivePath()
				if err != nil {
					log.Error("    ✗ %s (invalid path)", repo.URL)
					continue
				}

				fullRepoPath := filepath.Join(projectPath, repoPath)
				cloned := isRepoCloned(fullRepoPath)

				if cloned {
					log.Success("    ✓ %s", repoPath)
				} else {
					log.Warn("    ✗ %s (not cloned)", repoPath)
				}

				log.Info("      URL: %s", repo.URL)
				if repo.Path != "" {
					log.Info("      Custom path: %s", repo.Path)
				}
			}
		} else {
			clonedCount := 0
			for _, repo := range project.Repos {
				repoPath, err := repo.GetEffectivePath()
				if err != nil {
					continue
				}
				fullRepoPath := filepath.Join(projectPath, repoPath)
				if isRepoCloned(fullRepoPath) {
					clonedCount++
				}
			}

			statusIcon := "✓"
			if clonedCount < len(project.Repos) {
				statusIcon = "○"
			}
			log.Info("%s %s (%d/%d repos cloned)", statusIcon, project.Name, clonedCount, len(project.Repos))
		}
		log.Info("")
	}

	if !opts.IsVerbose {
		log.Info("Use -v for detailed repository information")
	}
}
