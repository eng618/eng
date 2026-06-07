package project

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
	internalProject "github.com/eng618/eng/internal/project"
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
		devPath := viper.GetString("git.dev_path")
		if devPath == "" {
			log.Error("Development folder path is not set. Use 'eng config git-dev-path' to set it.")
			return
		}

		opts := internalProject.FetchOptions{
			DryRun:        viper.GetBool("dry_run"),
			IsVerbose:     cmdutil.IsVerbose(cmd),
			ProjectFilter: viper.GetString("project_filter"),
			DevPath:       os.ExpandEnv(devPath),
			Projects:      config.GetProjects(),
		}

		internalProject.Fetch(cmd.Context(), opts)
	},
}
