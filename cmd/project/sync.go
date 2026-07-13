package project

import (
	"os"

	"github.com/spf13/cobra"

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
		gitCfg := config.GetGitConfig()
		devPath := gitCfg.DevPath
		if devPath == "" {
			log.Error("Development folder path is not set. Use 'eng config git-dev-path' to set it.")
			return
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		projectFilter, _ := cmd.Flags().GetString("project")

		opts := internalProject.SyncOptions{
			DryRun:        dryRun,
			IsVerbose:     cmdutil.IsVerbose(cmd),
			ProjectFilter: projectFilter,
			DevPath:       os.ExpandEnv(devPath),
			Projects:      config.GetProjects(),
		}

		ctx := cmd.Context()
		if ctx == nil {
			ctx = cmdutil.FallbackContext()
		}

		internalProject.Sync(ctx, opts)
	},
}
