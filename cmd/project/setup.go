package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/utils"
	"github.com/eng618/eng/internal/utils/config"
	"github.com/eng618/eng/internal/utils/log"
)

// SetupCmd defines the cobra command for setting up project repositories.
// It ensures project directories exist and clones any missing repositories.
var SetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup project directories and clone missing repositories",
	Long: `This command ensures all configured projects have their directory structure 
set up and all repositories are cloned.

It is safe to run multiple times - existing repositories will be skipped.
Use this command when:
  - Setting up a new development machine
  - A new repository has been added to a project's configuration
  - You want to verify all project repos are present

Example:
  eng project setup                  # Setup all projects
  eng project setup -p MyProject     # Setup only the specified project
  eng project setup --dry-run        # Preview what would be done`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Setting up project repositories")

		isVerbose := utils.IsVerbose(cmd)
		dryRun, _ := cmd.Parent().PersistentFlags().GetBool("dry-run")
		projectFilter, _ := cmd.Parent().PersistentFlags().GetString("project")

		devPath := viper.GetString("git.dev_path")
		if devPath == "" {
			log.Error("Development folder path is not set. Use 'eng config git-dev-path' to set it.")
			return
		}
		devPath = os.ExpandEnv(devPath)

		log.Verbose(isVerbose, "Development path: %s", devPath)

		if dryRun {
			log.Info("Dry run mode - no actual changes will be made")
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

		totalRepos := 0
		clonedCount := 0
		skippedCount := 0
		failedCount := 0

		for _, project := range projects {
			log.Info("Processing project: %s", project.Name)

			projectPath := filepath.Join(devPath, project.Name)

			// Ensure project directory exists
			if dryRun {
				log.Info("  [DRY RUN] Would ensure directory exists: %s", projectPath)
			} else {
				if err := os.MkdirAll(projectPath, 0o755); err != nil {
					log.Error("  Failed to create project directory: %s", err)
					continue
				}
				log.Verbose(isVerbose, "  Project directory ready: %s", projectPath)
			}

			for _, repo := range project.Repos {
				totalRepos++

				repoPath, err := repo.GetEffectivePath()
				if err != nil {
					log.Error("  Failed to determine path for %s: %s", repo.URL, err)
					failedCount++
					continue
				}

				fullRepoPath := filepath.Join(projectPath, repoPath)

				// Check if repo already exists
				if _, err := os.Stat(filepath.Join(fullRepoPath, ".git")); err == nil {
					log.Verbose(isVerbose, "  Repository already exists: %s", repoPath)
					skippedCount++
					continue
				}

				if dryRun {
					log.Info("  [DRY RUN] Would clone %s to %s", repo.URL, fullRepoPath)
					clonedCount++
					continue
				}

				log.Info("  Cloning %s...", repoPath)

				// Clone the repository
				if err := cloneRepository(repo.URL, fullRepoPath); err != nil {
					log.Error("  Failed to clone %s: %s", repo.URL, err)
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
		} else if !dryRun && clonedCount > 0 {
			log.Success("All project repositories set up successfully!")
		}
	},
}

// cloneRepository clones a git repository to the specified path.
// It uses go-git for the clone operation and provides informative error messages.
func cloneRepository(url, destPath string) error {
	_, err := git.PlainClone(destPath, false, &git.CloneOptions{
		URL:      url,
		Progress: log.Writer(),
	})
	if err != nil {
		errStr := strings.ToLower(err.Error())
		// Provide more helpful error messages for common issues
		switch {
		case errors.Is(err, git.ErrRepositoryAlreadyExists):
			return fmt.Errorf("repository already exists at %s", destPath)
		case strings.Contains(errStr, "authentication"):
			return fmt.Errorf(
				"authentication failed - ensure your SSH keys are configured or use HTTPS with credentials: %w",
				err,
			)
		case strings.Contains(errStr, "could not read username"):
			return fmt.Errorf(
				"credentials required - for HTTPS URLs, configure git credential helper or use SSH: %w",
				err,
			)
		case strings.Contains(errStr, "ssh:"):
			return fmt.Errorf(
				"SSH error - ensure your SSH keys are loaded (ssh-add) and have access to the repository: %w",
				err,
			)
		default:
			return err
		}
	}

	return nil
}
