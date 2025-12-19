package git

import (
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/log"
	"github.com/eng618/eng/utils/repo"
)

// BranchAllCmd defines the cobra command for showing current branch of all git repositories.
// It displays the current branch for all repositories in the development folder.
var BranchAllCmd = &cobra.Command{
	Use:   "branch-all",
	Short: "Show current branch of all git repositories in development folder",
	Long:  `This command shows the current branch for all git repositories found in your development folder.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Checking current branch of all git repositories")

		isVerbose := utils.IsVerbose(cmd)

		devPath, err := getWorkingPath(cmd)
		if err != nil {
			log.Error("%s", err)
			return
		}

		log.Verbose(isVerbose, "Development path: %s", devPath)

		repos, err := findGitRepositories(devPath)
		if err != nil {
			log.Error("Failed to find git repositories: %s", err)
			return
		}

		if len(repos) == 0 {
			log.Warn("No git repositories found in %s", devPath)
			return
		}

		log.Info("Checking branches of %d repositories:", len(repos))

		mainCount := 0
		otherBranchCount := 0

		for _, repoPath := range repos {
			repoName := filepath.Base(repoPath)

			// Get current branch
			branch, err := repo.GetCurrentBranch(repoPath)
			if err != nil {
				log.Error("  %s: Failed to get current branch - %s", repoName, err)
				continue
			}

			// Get main branch to compare
			mainBranch, err := repo.GetMainBranch(repoPath)
			if err != nil {
				log.Warn("  %s: Could not determine main branch - %s", repoName, err)
				mainBranch = "main" // fallback
			}

			if branch == mainBranch {
				log.Success("  %s: %s", repoName, branch)
				mainCount++
			} else {
				log.Warn("  %s: %s", repoName, branch)
				otherBranchCount++
			}
		}

		log.Info("Branch summary: %d on default branch, %d on other branches", mainCount, otherBranchCount)
	},
}
