package project

import (
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/project"
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
		dryRun, _ := cmd.Parent().PersistentFlags().GetBool("dry-run")
		projectFilter, _ := cmd.Parent().PersistentFlags().GetString("project")

		opts := project.SetupOptions{
			DryRun:        dryRun,
			IsVerbose:     cmdutil.IsVerbose(cmd),
			ProjectFilter: projectFilter,
		}

		project.Setup(opts)
	},
}
