package config

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/viper"

	"github.com/eng618/eng/utils/log"
)

// Project represents a collection of related repositories.
type Project struct {
	Name  string        `mapstructure:"name"`
	Repos []ProjectRepo `mapstructure:"repos"`
}

// ProjectRepo represents a single repository within a project.
type ProjectRepo struct {
	URL  string `mapstructure:"url"`
	Path string `mapstructure:"path,omitempty"` // Optional, defaults to repo name from URL
}

// GetProjects retrieves the list of configured projects from the config file.
func GetProjects() []Project {
	var projects []Project

	if !viper.IsSet("projects") {
		// Initialize with empty array if not set
		viper.Set("projects", []Project{})
		if err := viper.WriteConfig(); err != nil {
			log.Warn("Error initializing projects config: %v", err)
		}
		return projects
	}

	if err := viper.UnmarshalKey("projects", &projects); err != nil {
		log.Error("Failed to unmarshal projects configuration: %v", err)
		return []Project{}
	}

	return projects
}

// GetProjectByName retrieves a specific project by name.
// Returns nil if not found.
func GetProjectByName(name string) *Project {
	projects := GetProjects()
	for i := range projects {
		if strings.EqualFold(projects[i].Name, name) {
			return &projects[i]
		}
	}
	return nil
}

// SaveProjects persists the list of projects to the config file.
func SaveProjects(projects []Project) error {
	viper.Set("projects", projects)
	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("failed to save projects configuration: %w", err)
	}
	return nil
}

// AddProject adds a new project or updates an existing one.
// If a project with the same name exists, it updates it; otherwise, it appends.
func AddProject(project Project) error {
	projects := GetProjects()

	// Check if project already exists
	for i := range projects {
		if strings.EqualFold(projects[i].Name, project.Name) {
			projects[i] = project
			return SaveProjects(projects)
		}
	}

	// Append new project
	projects = append(projects, project)
	return SaveProjects(projects)
}

// AddRepoToProject adds a repository to an existing project.
// Returns an error if the project doesn't exist.
func AddRepoToProject(projectName string, repo ProjectRepo) error {
	projects := GetProjects()

	for i := range projects {
		if strings.EqualFold(projects[i].Name, projectName) {
			// Check if repo already exists
			for _, existingRepo := range projects[i].Repos {
				if existingRepo.URL == repo.URL {
					return fmt.Errorf("repository %s already exists in project %s", repo.URL, projectName)
				}
			}
			projects[i].Repos = append(projects[i].Repos, repo)
			return SaveProjects(projects)
		}
	}

	return fmt.Errorf("project %s not found", projectName)
}

// RemoveProject removes a project from the configuration.
func RemoveProject(projectName string) error {
	projects := GetProjects()
	newProjects := make([]Project, 0, len(projects))

	found := false
	for _, p := range projects {
		if strings.EqualFold(p.Name, projectName) {
			found = true
			continue
		}
		newProjects = append(newProjects, p)
	}

	if !found {
		return fmt.Errorf("project %s not found", projectName)
	}

	return SaveProjects(newProjects)
}

// RemoveRepoFromProject removes a repository from a project.
func RemoveRepoFromProject(projectName, repoURL string) error {
	projects := GetProjects()

	for i := range projects {
		if strings.EqualFold(projects[i].Name, projectName) {
			newRepos := make([]ProjectRepo, 0, len(projects[i].Repos))
			found := false

			for _, repo := range projects[i].Repos {
				if repo.URL == repoURL {
					found = true
					continue
				}
				newRepos = append(newRepos, repo)
			}

			if !found {
				return fmt.Errorf("repository %s not found in project %s", repoURL, projectName)
			}

			projects[i].Repos = newRepos
			return SaveProjects(projects)
		}
	}

	return fmt.Errorf("project %s not found", projectName)
}

// RepoNameFromURL extracts the repository name from a git URL.
// Supports both SSH (git@host:path/repo.git) and HTTPS (https://host/path/repo.git) formats.
func RepoNameFromURL(repoURL string) (string, error) {
	// SSH format: git@host:path/repo.git or ssh://git@host/path/repo.git
	sshPattern := regexp.MustCompile(`^(?:git|ssh)@[^:]+:(.+?)(?:\.git)?$`)
	if matches := sshPattern.FindStringSubmatch(repoURL); len(matches) == 2 {
		path := strings.TrimSuffix(matches[1], ".git")
		return filepath.Base(path), nil
	}

	// SSH with protocol: ssh://git@host/path/repo.git
	sshProtoPattern := regexp.MustCompile(`^ssh://[^/]+/(.+?)(?:\.git)?$`)
	if matches := sshProtoPattern.FindStringSubmatch(repoURL); len(matches) == 2 {
		path := strings.TrimSuffix(matches[1], ".git")
		return filepath.Base(path), nil
	}

	// HTTPS format: https://host/path/repo.git
	if strings.HasPrefix(repoURL, "http://") || strings.HasPrefix(repoURL, "https://") {
		u, err := url.Parse(repoURL)
		if err != nil {
			return "", fmt.Errorf("failed to parse URL: %w", err)
		}
		path := strings.TrimPrefix(u.Path, "/")
		path = strings.TrimSuffix(path, ".git")
		return filepath.Base(path), nil
	}

	return "", fmt.Errorf("unsupported URL format: %s", repoURL)
}

// GetEffectivePath returns the effective path for a repo.
// If Path is set, it uses that; otherwise, it derives from the URL.
func (r *ProjectRepo) GetEffectivePath() (string, error) {
	if r.Path != "" {
		return r.Path, nil
	}
	return RepoNameFromURL(r.URL)
}

// GetProjectNames returns a list of all configured project names.
func GetProjectNames() []string {
	projects := GetProjects()
	names := make([]string, len(projects))
	for i, p := range projects {
		names[i] = p.Name
	}
	return names
}
