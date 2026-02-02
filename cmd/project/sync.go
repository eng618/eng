package project

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/utils"
	"github.com/eng618/eng/internal/utils/config"
	"github.com/eng618/eng/internal/utils/log"
	"github.com/eng618/eng/internal/utils/repo"
)

// SyncCmd defines the cobra command for syncing all project repositories.
// It fetches and pulls all repositories in configured projects.
var SyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync all project repositories (fetch + pull)",
	Long: `This command synchronizes all repositories in configured projects by:
  1. Fetching updates from remote (git fetch --all --prune)
  2. Pulling changes for the current branch (git pull)

Repositories with uncommitted changes will have fetch performed but pull will be skipped.

Example:
  eng project sync                  # Sync all projects
  eng project sync -p MyProject     # Sync only the specified project
  eng project sync --dry-run        # Preview what would be synced`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Syncing project repositories")

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

		fetchSuccess := 0
		fetchFailed := 0
		pullSuccess := 0
		pullFailed := 0
		skippedCount := 0
		dirtyCount := 0

		for _, project := range projects {
			log.Info("Syncing project: %s", project.Name)
			projectPath := filepath.Join(devPath, project.Name)

			for _, r := range project.Repos {
				repoPath, err := r.GetEffectivePath()
				if err != nil {
					log.Error("  Failed to determine path for %s: %s", r.URL, err)
					fetchFailed++
					pullFailed++
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
					log.Info("  [DRY RUN] Would sync: %s", repoPath)
					fetchSuccess++
					pullSuccess++
					continue
				}

				log.Info("  Syncing %s...", repoPath)

				// Fetch
				if err := fetchRepo(fullRepoPath); err != nil {
					log.Error("    Fetch failed: %s", err)
					fetchFailed++
				} else {
					log.Verbose(isVerbose, "    Fetched successfully")
					fetchSuccess++
				}

				// Check for uncommitted changes before pull
				isDirty, err := repo.IsDirty(fullRepoPath)
				if err != nil {
					log.Error("    Failed to check status: %s", err)
					pullFailed++
					continue
				}

				if isDirty {
					log.Warn("    Skipping pull (has uncommitted changes)")
					dirtyCount++
					continue
				}

				// Pull
				if err := repo.PullLatestCode(fullRepoPath); err != nil {
					if err.Error() == "already up-to-date" {
						log.Verbose(isVerbose, "    Already up to date")
						pullSuccess++
						continue
					}
					log.Error("    Pull failed: %s", err)
					pullFailed++
					continue
				}

				log.Success("    Synced %s", repoPath)
				pullSuccess++
			}
		}

		log.Info("")
		log.Info("Sync complete:")
		log.Info("  Fetch: %d successful, %d failed", fetchSuccess, fetchFailed)
		log.Info(
			"  Pull:  %d successful, %d failed, %d dirty, %d skipped",
			pullSuccess,
			pullFailed,
			dirtyCount,
			skippedCount,
		)

		if dirtyCount > 0 {
			log.Warn("Some repositories were not pulled due to uncommitted changes.")
		}
	},
}
