package project

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/utils"
	"github.com/eng618/eng/internal/utils/config"
	"github.com/eng618/eng/internal/utils/log"
	"github.com/eng618/eng/internal/utils/repo"
)

// PullCmd defines the cobra command for pulling all project repositories.
// It runs git pull on all repositories in configured projects.
var PullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull updates for all project repositories",
	Long: `This command pulls the latest changes from remote for all repositories in configured projects.

Note: Repositories with uncommitted changes will be skipped.

Example:
  eng project pull                  # Pull all projects
  eng project pull -p MyProject     # Pull only the specified project
  eng project pull --dry-run        # Preview what would be pulled`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Pulling project repositories")

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
		dirtyCount := 0

		for _, project := range projects {
			log.Info("Pulling project: %s", project.Name)
			projectPath := filepath.Join(devPath, project.Name)

			for _, r := range project.Repos {
				repoPath, err := r.GetEffectivePath()
				if err != nil {
					log.Error("  Failed to determine path for %s: %s", r.URL, err)
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
					log.Info("  [DRY RUN] Would pull: %s", repoPath)
					successCount++
					continue
				}

				// Check for uncommitted changes
				isDirty, err := repo.IsDirty(fullRepoPath)
				if err != nil {
					log.Error("  Failed to check status of %s: %s", repoPath, err)
					failedCount++
					continue
				}

				if isDirty {
					log.Warn("  Skipping %s (has uncommitted changes)", repoPath)
					dirtyCount++
					continue
				}

				log.Info("  Pulling %s...", repoPath)
				if err := repo.PullLatestCode(fullRepoPath); err != nil {
					// Check if it's just "already up to date"
					if errors.Is(err, git.NoErrAlreadyUpToDate) {
						log.Info("  %s is already up to date", repoPath)
						successCount++
						continue
					}
					log.Error("  Failed to pull %s: %s", repoPath, err)
					failedCount++
					continue
				}

				log.Success("  Pulled %s", repoPath)
				successCount++
			}
		}

		log.Info("")
		log.Info(
			"Pull complete: %d successful, %d skipped, %d dirty, %d failed",
			successCount,
			skippedCount,
			dirtyCount,
			failedCount,
		)

		if dirtyCount > 0 {
			log.Warn("Some repositories were skipped due to uncommitted changes.")
			log.Info("Commit or stash your changes, then run again.")
		}
	},
}
