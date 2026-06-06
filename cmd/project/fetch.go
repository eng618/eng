package project

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
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

		isVerbose := cmdutil.IsVerbose(cmd)
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

		var successCount atomic.Int32
		var failedCount atomic.Int32
		var skippedCount atomic.Int32

		multi, err := ui.NewMultiSpinner()
		if err != nil {
			log.Error("Failed to initialize UI: %s", err)
			return
		}
		defer multi.Stop()

		var eg errgroup.Group
		eg.SetLimit(10) // Concurrent fetch limit

		for _, project := range projects {
			projectPath := filepath.Join(devPath, project.Name)

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
						if isVerbose {
							spinner := multi.AddSpinner(fmt.Sprintf("Skipping %s (not cloned)", repoPath))
							spinner.Warning()
						}
						skippedCount.Add(1)
						return nil
					}

					if dryRun {
						spinner := multi.AddSpinner(fmt.Sprintf("[DRY RUN] Would fetch: %s", repoPath))
						spinner.Success()
						successCount.Add(1)
						return nil
					}

					spinner := multi.AddSpinner(fmt.Sprintf("Fetching %s...", repoPath))
					if err := fetchRepo(fullRepoPath); err != nil {
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
	},
}

// fetchRepo performs git fetch on a repository.
func fetchRepo(repoPath string) error {
	cmd := exec.Command("git", "-C", repoPath, "fetch", "--all", "--prune")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(out))
	}
	return nil
}
