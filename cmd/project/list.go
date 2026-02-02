package project

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/config"
	"github.com/eng618/eng/utils/log"
)

// ListCmd defines the cobra command for listing configured projects.
// It displays project names, repository counts, and clone status.
var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured projects and their repositories",
	Long: `This command displays all configured projects and their repositories.

Use the --verbose flag to see detailed information including:
  - Repository URLs
  - Clone status (✓ cloned / ✗ missing)
  - Local paths

Example:
  eng project list               # Show projects summary
  eng project list -v            # Show detailed repository information
  eng project list -p MyProject  # Show only the specified project`,
	Run: func(cmd *cobra.Command, args []string) {
		isVerbose := utils.IsVerbose(cmd)
		projectFilter, _ := cmd.Parent().PersistentFlags().GetString("project")

		devPath := viper.GetString("git.dev_path")
		if devPath == "" {
			log.Warn("Development folder path is not set. Use 'eng config git-dev-path' to set it.")
			devPath = "(not configured)"
		} else {
			devPath = os.ExpandEnv(devPath)
		}

		projects := config.GetProjects()
		if len(projects) == 0 {
			log.Info("No projects configured.")
			log.Info("Use 'eng project add' to add a project.")
			return
		}

		// Filter by project if specified
		if projectFilter != "" {
			filtered := make([]config.Project, 0)
			for _, p := range projects {
				if p.Name == projectFilter {
					filtered = append(filtered, p)
					break
				}
			}
			if len(filtered) == 0 {
				log.Error("Project '%s' not found in configuration", projectFilter)
				return
			}
			projects = filtered
		}

		log.Info("Development path: %s", devPath)
		log.Info("")

		for _, project := range projects {
			projectPath := filepath.Join(devPath, project.Name)

			if isVerbose {
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

		if !isVerbose {
			log.Info("Use -v for detailed repository information")
		}
	},
}

// isRepoCloned checks if a repository has been cloned at the given path.
func isRepoCloned(repoPath string) bool {
	gitDir := filepath.Join(repoPath, ".git")
	if info, err := os.Stat(gitDir); err == nil {
		return info.IsDir()
	}
	return false
}
