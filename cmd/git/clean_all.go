package git

import (
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/log"
)

// CleanAllCmd defines the cobra command for cleaning untracked files in all git repositories.
// It removes untracked files and directories for all repositories in the development folder.
var CleanAllCmd = &cobra.Command{
	Use:   "clean-all",
	Short: "Clean untracked files in all git repositories in development folder",
	Long:  `This command removes untracked files and directories for all git repositories found in your development folder.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Cleaning untracked files in all git repositories")

		isVerbose := utils.IsVerbose(cmd)
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		force, _ := cmd.Flags().GetBool("force")
		directories, _ := cmd.Flags().GetBool("directories")

		devPath, err := getWorkingPath(cmd)
		if err != nil {
			log.Error("%s", err)
			return
		}

		log.Verbose(isVerbose, "Development path: %s", devPath)

		if dryRun {
			log.Info("Dry run mode - no actual git operations will be performed")
		}

		if !force && !dryRun {
			log.Warn("This will permanently delete untracked files. Use --force to confirm or --dry-run to preview.")
			return
		}

		repos, err := findGitRepositories(devPath)
		if err != nil {
			log.Error("Failed to find git repositories: %s", err)
			return
		}

		if len(repos) == 0 {
			log.Warn("No git repositories found in %s", devPath)
			return
		}

		log.Info("Found %d git repositories", len(repos))

		successCount := 0
		failureCount := 0
		skippedCount := 0

		for _, repoPath := range repos {
			repoName := filepath.Base(repoPath)
			log.Info("Checking repository: %s", repoName)

			if dryRun {
				hasUntracked, err := hasUntrackedFiles(repoPath)
				if err != nil {
					log.Error("  [DRY RUN] Failed to check for untracked files: %s", err)
					failureCount++
					continue
				}
				if hasUntracked {
					log.Info("  [DRY RUN] Would clean untracked files in: %s", repoPath)
					// Show what would be cleaned
					if err := showUntrackedFiles(repoPath); err != nil {
						log.Error("  [DRY RUN] Failed to show untracked files: %s", err)
					}
					successCount++
				} else {
					log.Info("  [DRY RUN] No untracked files to clean, skipping: %s", repoPath)
					skippedCount++
				}
				continue
			}

			// Check if repository has untracked files
			hasUntracked, err := hasUntrackedFiles(repoPath)
			if err != nil {
				log.Error("  Failed to check for untracked files: %s", err)
				failureCount++
				continue
			}

			if !hasUntracked {
				log.Info("  No untracked files to clean, skipping...")
				skippedCount++
				continue
			}

			// Clean untracked files
			if err := cleanRepository(repoPath, directories); err != nil {
				log.Error("  Failed to clean %s: %s", repoName, err)
				failureCount++
				continue
			}

			log.Success("  Successfully cleaned untracked files in %s", repoName)
			successCount++
		}

		log.Info("Clean completed: %d successful, %d failed, %d skipped", successCount, failureCount, skippedCount)

		if failureCount > 0 {
			log.Warn("Some repositories failed to clean. Check the output above for details.")
		} else {
			log.Success("All git repositories with untracked files cleaned successfully")
		}
	},
}

func init() {
	CleanAllCmd.Flags().Bool("dry-run", false, "Perform a dry run without making actual changes")
	CleanAllCmd.Flags().Bool("force", false, "Force clean untracked files (required for actual cleaning)")
	CleanAllCmd.Flags().BoolP("directories", "d", false, "Also remove untracked directories")
}

// cleanRepository performs a git clean operation on the given repository path.
func cleanRepository(repoPath string, directories bool) error {
	args := []string{"-C", repoPath, "clean", "-f"}
	if directories {
		args = append(args, "-d")
	}

	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error("Git clean output: %s", string(output))
		return err
	}
	log.Info("Git clean output: %s", string(output))
	return nil
}

// hasUntrackedFiles checks if the repository has untracked files.
func hasUntrackedFiles(repoPath string) (bool, error) {
	cmd := exec.Command("git", "-C", repoPath, "ls-files", "--others", "--exclude-standard")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	return len(output) > 0, nil
}

// showUntrackedFiles shows what untracked files would be cleaned (for dry-run).
func showUntrackedFiles(repoPath string) error {
	cmd := exec.Command("git", "-C", repoPath, "clean", "-n")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	if len(output) > 0 {
		log.Info("    Files to be cleaned: %s", string(output))
	}
	return nil
}
