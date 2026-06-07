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
		devPath := viper.GetString("git.dev_path")
		if devPath == "" {
			log.Error("Development folder path is not set. Use 'eng config git-dev-path' to set it.")
			return
		}

		opts := internalProject.SetupOptions{
			DryRun:        viper.GetBool("dry_run"),
			IsVerbose:     cmdutil.IsVerbose(cmd),
			ProjectFilter: viper.GetString("project_filter"),
			DevPath:       os.ExpandEnv(devPath),
			Projects:      config.GetProjects(),
		}

		internalProject.Setup(cmd.Context(), opts)
	},
}
