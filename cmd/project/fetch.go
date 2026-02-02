package project

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/config"
	"github.com/eng618/eng/utils/log"
)

// FetchCmd defines the cobra command for fetching all project repositories.
// It runs git fetch on all repositories in configured projects.
var FetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch updates for all project repositories",
	Long: `This command fetches updates from remote for all repositories in configured projects.

Example:
  eng project fetch                  # Fetch all projects
  eng project fetch -p MyProject     # Fetch only the specified project
  eng project fetch --dry-run        # Preview what would be fetched`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Fetching project repositories")

		isVerbose := utils.IsVerbose(cmd)
		dryRun, _ := cmd.Parent().PersistentFlags().GetBool("dry-run")
		projectFilter, _ := cmd.Parent().PersistentFlags().GetString("project")

		devPath := viper.GetString("git.dev_path")
		if devPath == "" {
			log.Error("Development folder path is not set. Use 'eng config git-dev-path' to set it.")
			return
		}
		devPath = os.ExpandEnv(devPath)

		if dryRun {
			log.Info("Dry run mode - no actual git operations will be performed")
		}

		projects := config.GetProjects()
		if len(projects) == 0 {
			log.Warn("No projects configured. Use 'eng project add' to add a project.")
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

		successCount := 0
		failedCount := 0
		skippedCount := 0

		for _, project := range projects {
			log.Info("Fetching project: %s", project.Name)
			projectPath := filepath.Join(devPath, project.Name)

			for _, repo := range project.Repos {
				repoPath, err := repo.GetEffectivePath()
				if err != nil {
					log.Error("  Failed to determine path for %s: %s", repo.URL, err)
					failedCount++
					continue
				}

				fullRepoPath := filepath.Join(projectPath, repoPath)

				// Check if repo exists
				if !isRepoCloned(fullRepoPath) {
					log.Verbose(isVerbose, "  Skipping %s (not cloned)", repoPath)
					skippedCount++
					continue
				}

				if dryRun {
					log.Info("  [DRY RUN] Would fetch: %s", repoPath)
					successCount++
					continue
				}

				log.Info("  Fetching %s...", repoPath)
				if err := fetchRepo(fullRepoPath); err != nil {
					log.Error("  Failed to fetch %s: %s", repoPath, err)
					failedCount++
					continue
				}

				log.Success("  Fetched %s", repoPath)
				successCount++
			}
		}

		log.Info("")
		log.Info("Fetch complete: %d successful, %d skipped, %d failed", successCount, skippedCount, failedCount)
	},
}

// fetchRepo performs git fetch on a repository.
func fetchRepo(repoPath string) error {
	cmd := exec.Command("git", "-C", repoPath, "fetch", "--all", "--prune")
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()
	return cmd.Run()
}
