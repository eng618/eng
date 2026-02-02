// Package project provides cobra commands for managing project-based repository collections.
// A project is a logical grouping of related repositories (e.g., a product with multiple microservices).
package project

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/log"
)

// ProjectCmd serves as the base command for all project management operations.
// It groups subcommands like setup, list, add, remove, fetch, pull, and sync.
var ProjectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage project-based repository collections",
	Long: `This command facilitates the management of project-based repository collections.

A project is a logical grouping of related repositories. For example, you might have 
a project containing multiple microservices, or a shared infrastructure project.

Projects are stored in your development folder (configured via 'eng config git-dev-path'),
with each project having its own subdirectory containing all related repositories.

Example structure:
  ~/Development/
    MyProject/
      api/
      web/
      shared/
    Infrastructure/
      core/
      auth/`,
	Run: func(cmd *cobra.Command, args []string) {
		showInfo, _ := cmd.Flags().GetBool("info")
		isVerbose := utils.IsVerbose(cmd)

		if showInfo {
			log.Info("Current project management configuration:")
			devPath := viper.GetString("git.dev_path")

			if devPath == "" {
				log.Warn("  Development Path: Not Set")
				log.Info("  Use 'eng config git-dev-path' to set your development folder path")
			} else {
				log.Info("  Development Path: %s", devPath)
			}

			// Show configured projects
			var projects []map[string]interface{}
			if err := viper.UnmarshalKey("projects", &projects); err == nil && len(projects) > 0 {
				log.Info("  Configured Projects: %d", len(projects))
			} else {
				log.Info("  Configured Projects: 0")
				log.Info("  Use 'eng project add' to configure a project")
			}
			return
		}

		// If no subcommand is given, print the help information.
		if len(args) == 0 {
			log.Verbose(isVerbose, "No subcommand provided, showing help.")
			err := cmd.Help()
			cobra.CheckErr(err)
		} else {
			log.Verbose(isVerbose, "Subcommand '%s' provided.", args[0])
		}
	},
}

func init() {
	ProjectCmd.Flags().BoolP("info", "i", false, "Show current project management configuration")
	ProjectCmd.PersistentFlags().StringP("project", "p", "", "Filter operations to a specific project")
	ProjectCmd.PersistentFlags().Bool("dry-run", false, "Perform a dry run without making actual changes")

	// Register subcommands
	ProjectCmd.AddCommand(SetupCmd)
	ProjectCmd.AddCommand(ListCmd)
	ProjectCmd.AddCommand(AddCmd)
	ProjectCmd.AddCommand(RemoveCmd)
	ProjectCmd.AddCommand(FetchCmd)
	ProjectCmd.AddCommand(PullCmd)
	ProjectCmd.AddCommand(SyncCmd)
}
