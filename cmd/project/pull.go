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

// PullCmd defines the cobra command for pulling all project repositories.
// It runs git pull on all repositories in configured projects.
var PullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull updates for all project repositories",
	Long: `This command pulls the latest changes from remote for all repositories in configured projects.

Note: Repositories with uncommitted changes will be skipped.

Example:
  eng project pull                  # Pull all projects
  eng project pull -p MyProject     # Pull only the specified project
  eng project pull --dry-run        # Preview what would be pulled`,
	Run: func(cmd *cobra.Command, args []string) {
		devPath := viper.GetString("git.dev_path")
		if devPath == "" {
			log.Error("Development folder path is not set. Use 'eng config git-dev-path' to set it.")
			return
		}

		opts := internalProject.PullOptions{
			DryRun:        viper.GetBool("dry_run"),
			IsVerbose:     cmdutil.IsVerbose(cmd),
			ProjectFilter: viper.GetString("project_filter"),
			DevPath:       os.ExpandEnv(devPath),
			Projects:      config.GetProjects(),
		}

		internalProject.Pull(cmd.Context(), opts)
	},
}
