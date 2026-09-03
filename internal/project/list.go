package project

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/repo"
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

	// Sort projects alphabetically
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})

	log.Info("Development path: %s", devPath)
	log.Info("")

	for _, project := range projects {
		projectPath := filepath.Join(devPath, project.Name)

		if opts.IsVerbose {
			log.Info("Project: %s", project.Name)
			log.Info("  Path: %s", projectPath)
			log.Info("  Repositories (%d):", len(project.Repos))

			sortedRepos := make([]config.ProjectRepo, len(project.Repos))
			copy(sortedRepos, project.Repos)
			sort.Slice(sortedRepos, func(i, j int) bool {
				return sortedRepos[i].URL < sortedRepos[j].URL
			})

			for _, repoItem := range sortedRepos {
				repoPath, err := repoItem.GetEffectivePath()
				if err != nil {
					log.Error("    %s (invalid path)", repoItem.URL)
					continue
				}

				fullRepoPath := filepath.Join(projectPath, repoPath)
				cloned := repo.IsCloned(fullRepoPath)

				if cloned {
					log.Success("    %s (cloned)", repoPath)
				} else {
					log.Warn("    %s (not cloned)", repoPath)
				}

				log.Info("      URL: %s", repoItem.URL)
				if repoItem.Path != "" {
					log.Info("      Custom path: %s", repoItem.Path)
				}
			}
		} else {
			clonedCount := 0
			for _, r := range project.Repos {
				repoPath, err := r.GetEffectivePath()
				if err != nil {
					continue
				}
				fullRepoPath := filepath.Join(projectPath, repoPath)
				if repo.IsCloned(fullRepoPath) {
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
