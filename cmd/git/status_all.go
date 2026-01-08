package git

import (
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/utils"
	"github.com/eng618/eng/internal/utils/log"
	"github.com/eng618/eng/internal/utils/repo"
)

// StatusAllCmd defines the cobra command for checking status of all git repositories.
// It shows the status of all repositories in the development folder.
var StatusAllCmd = &cobra.Command{
	Use:   "status-all",
	Short: "Check status of all git repositories in development folder",
	Long:  `This command checks the status of all git repositories found in your development folder.`,
	Run: func(cmd *cobra.Command, _args []string) {
		log.Start("Checking status of all git repositories")

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

		log.Info("Checking status of %d repositories:", len(repos))

		cleanCount := 0
		dirtyCount := 0

		for _, repoPath := range repos {
			repoName := filepath.Base(repoPath)

			// Check if repository is dirty
			isDirty, err := repo.IsDirty(repoPath)
			if err != nil {
				log.Error("  %s: Failed to check status - %s", repoName, err)
				continue
			}

			if isDirty {
				log.Warn("  %s: Has uncommitted changes", repoName)
				dirtyCount++
			} else {
				log.Success("  %s: Clean", repoName)
				cleanCount++
			}
		}

		log.Info("Status summary: %d clean, %d with uncommitted changes", cleanCount, dirtyCount)
	},
}
