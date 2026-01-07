package git

import (
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/utils"
	"github.com/eng618/eng/internal/utils/log"
)

// ListCmd defines the cobra command for listing all git repositories.
// It lists all git repositories found in the development folder.
var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all git repositories in development folder",
	Long:  `This command lists all git repositories found in your development folder.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Listing git repositories")

		isVerbose := utils.IsVerbose(cmd)
		showPaths, _ := cmd.Flags().GetBool("paths")

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

		log.Info("Found %d git repositories:", len(repos))

		for i, repoPath := range repos {
			repoName := filepath.Base(repoPath)
			if showPaths {
				log.Info("  %d. %s (%s)", i+1, repoName, repoPath)
			} else {
				log.Info("  %d. %s", i+1, repoName)
			}
		}

		log.Success("Listed %d git repositories", len(repos))
	},
}

func init() {
	ListCmd.Flags().BoolP("paths", "p", false, "Show full paths for each repository")
}
