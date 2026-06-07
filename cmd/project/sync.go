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
		devPath := viper.GetString("git.dev_path")
		if devPath == "" {
			log.Error("Development folder path is not set. Use 'eng config git-dev-path' to set it.")
			return
		}

		opts := internalProject.SyncOptions{
			DryRun:        viper.GetBool("dry_run"),
			IsVerbose:     cmdutil.IsVerbose(cmd),
			ProjectFilter: viper.GetString("project_filter"),
			DevPath:       os.ExpandEnv(devPath),
			Projects:      config.GetProjects(),
		}

		internalProject.Sync(cmd.Context(), opts)
	},
}
