package project

import (
	"strconv"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/utils/config"
	"github.com/eng618/eng/internal/utils/log"
)

// RemoveCmd defines the cobra command for removing projects or repositories.
// It provides interactive prompts for safe removal.
var RemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a project or repository from configuration",
	Long: `This command removes a project or a repository from a project's configuration.

Note: This only removes the entry from your configuration. 
It does NOT delete any files from disk.

Example:
  eng project remove                  # Interactive removal
  eng project remove -p MyProject     # Remove from the specified project`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Removing project configuration")

		projectFilter, _ := cmd.Flags().GetString("project")

		projects := config.GetProjects()
		if len(projects) == 0 {
			log.Info("No projects configured.")
			return
		}

		existingNames := config.GetProjectNames()

		var projectName string

		if projectFilter != "" {
			// Use the specified project
			projectName = projectFilter
			found := false
			for _, name := range existingNames {
				if name == projectFilter {
					found = true
					break
				}
			}
			if !found {
				log.Error("Project '%s' not found", projectFilter)
				return
			}
		} else {
			// Interactive project selection
			prompt := &survey.Select{
				Message: "Select a project:",
				Options: existingNames,
			}
			if err := survey.AskOne(prompt, &projectName); err != nil {
				log.Error("Prompt failed: %s", err)
				return
			}
		}

		// Get the project
		project := config.GetProjectByName(projectName)
		if project == nil {
			log.Error("Project '%s' not found", projectName)
			return
		}

		// Ask what to remove
		options := []string{"[Remove entire project]"}
		for _, repo := range project.Repos {
			path, err := repo.GetEffectivePath()
			if err != nil {
				path = repo.URL
			}
			options = append(options, path)
		}

		var selection string
		selectPrompt := &survey.Select{
			Message: "What would you like to remove?",
			Options: options,
		}
		if err := survey.AskOne(selectPrompt, &selection); err != nil {
			log.Error("Prompt failed: %s", err)
			return
		}

		if selection == "[Remove entire project]" {
			// Confirm project removal
			var confirm bool
			confirmPrompt := &survey.Confirm{
				Message: "Remove project '" + projectName + "' with " + strconv.Itoa(
					len(project.Repos),
				) + " repositories?",
				Default: false,
			}
			if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
				log.Error("Prompt failed: %s", err)
				return
			}

			if !confirm {
				log.Info("Canceled.")
				return
			}

			if err := config.RemoveProject(projectName); err != nil {
				log.Error("Failed to remove project: %s", err)
				return
			}

			log.Success("Removed project '%s' from configuration", projectName)
			log.Info("Note: Files on disk were not deleted.")
		} else {
			// Find the repo URL for the selected path
			var repoURL string
			for _, repo := range project.Repos {
				path, _ := repo.GetEffectivePath()
				if path == selection {
					repoURL = repo.URL
					break
				}
			}

			if repoURL == "" {
				log.Error("Repository not found")
				return
			}

			// Confirm repo removal
			var confirm bool
			confirmPrompt := &survey.Confirm{
				Message: "Remove repository '" + selection + "' from project '" + projectName + "'?",
				Default: false,
			}
			if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
				log.Error("Prompt failed: %s", err)
				return
			}

			if !confirm {
				log.Info("Canceled.")
				return
			}

			if err := config.RemoveRepoFromProject(projectName, repoURL); err != nil {
				log.Error("Failed to remove repository: %s", err)
				return
			}

			log.Success("Removed repository '%s' from project '%s'", selection, projectName)
			log.Info("Note: Files on disk were not deleted.")
		}
	},
}
