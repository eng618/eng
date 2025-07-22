package git

import (
	"os/exec"
	"path/filepath"

	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/log"
	"github.com/spf13/cobra"
)

// FetchAllCmd defines the cobra command for fetching all git repositories.
// It fetches updates from remote for all repositories in the development folder.
var FetchAllCmd = &cobra.Command{
	Use:   "fetch-all",
	Short: "Fetch all git repositories in development folder",
	Long:  `This command fetches updates from remote for all git repositories found in your development folder.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Fetching all git repositories")

		isVerbose := utils.IsVerbose(cmd)
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		devPath, err := getWorkingPath(cmd)
		if err != nil {
			log.Error("%s", err)
			return
		}

		log.Verbose(isVerbose, "Development path: %s", devPath)

		if dryRun {
			log.Info("Dry run mode - no actual git operations will be performed")
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

		for _, repoPath := range repos {
			repoName := filepath.Base(repoPath)
			log.Info("Fetching repository: %s", repoName)

			if dryRun {
				log.Info("  [DRY RUN] Would fetch repository at: %s", repoPath)
				successCount++
				continue
			}

			// Perform git fetch
			if err := fetchRepository(repoPath); err != nil {
				log.Error("  Failed to fetch %s: %s", repoName, err)
				failureCount++
				continue
			}

			log.Success("  Successfully fetched %s", repoName)
			successCount++
		}

		log.Info("Fetch completed: %d successful, %d failed", successCount, failureCount)

		if failureCount > 0 {
			log.Warn("Some repositories failed to fetch. Check the output above for details.")
		} else {
			log.Success("All git repositories fetched successfully")
		}
	},
}

func init() {
	FetchAllCmd.Flags().Bool("dry-run", false, "Perform a dry run without making actual changes")
}

// fetchRepository performs a git fetch operation on the given repository path.
func fetchRepository(repoPath string) error {
	cmd := exec.Command("git", "-C", repoPath, "fetch", "--all", "--prune")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error("Git fetch output: %s", string(output))
		return err
	}
	log.Info("Git fetch output: %s", string(output))
	return nil
}
