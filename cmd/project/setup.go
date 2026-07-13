package project

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
	internalProject "github.com/eng618/eng/internal/project"
)

// SetupCmd defines the cobra command for setting up project repositories.
var SetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup project directories and clone missing repositories",
	Long: `This command ensures all configured projects have their directory structure 
set up and all repositories are cloned.

It is safe to run multiple times - existing repositories will be skipped.
Use this command when:
  - Setting up a new development machine
  - A new repository has been added to a project's configuration
  - You want to verify all project repos are present

Example:
  eng project setup                  # Setup all projects
  eng project setup -p MyProject     # Setup only the specified project
  eng project setup --dry-run        # Preview what would be done`,
	Run: func(cmd *cobra.Command, args []string) {
		gitCfg := config.GetGitConfig()
		devPath := gitCfg.DevPath
		if devPath == "" {
			log.Error("Development folder path is not set. Use 'eng config git-dev-path' to set it.")
			return
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		projectFilter, _ := cmd.Flags().GetString("project")

		opts := internalProject.SetupOptions{
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

		internalProject.Setup(ctx, opts)
	},
}
