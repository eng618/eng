package project

import (
	"os"

	"github.com/spf13/cobra"

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
		gitCfg := config.GetGitConfig()
		devPath := gitCfg.DevPath
		if devPath == "" {
			log.Error("Development folder path is not set. Use 'eng config git-dev-path' to set it.")
			return
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		projectFilter, _ := cmd.Flags().GetString("project")

		opts := internalProject.FetchOptions{
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

		internalProject.Fetch(ctx, opts)
	},
}
