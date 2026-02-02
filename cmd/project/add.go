package project

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/utils/config"
	"github.com/eng618/eng/internal/utils/log"
)

// AddCmd defines the cobra command for adding projects or repositories.
// It provides interactive prompts for configuration.
var AddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new project or repository to configuration",
	Long: `This command interactively adds a new project or adds a repository to an existing project.

You will be prompted for:
  - Project name (new or existing)
  - Repository URL (SSH or HTTPS)
  - Optional custom directory name

Example:
  eng project add                  # Interactive add
  eng project add -p MyProject     # Add a repo to the specified project`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Adding project configuration")

		projectFilter, _ := cmd.Flags().GetString("project")

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
				// Project doesn't exist, confirm creation
				var confirmCreate bool
				prompt := &survey.Confirm{
					Message: "Project '" + projectFilter + "' doesn't exist. Create it?",
					Default: true,
				}
				if err := survey.AskOne(prompt, &confirmCreate); err != nil {
					log.Error("Prompt failed: %s", err)
					return
				}
				if !confirmCreate {
					log.Info("Canceled.")
					return
				}
			}
		} else {
			// Interactive project selection/creation
			options := append([]string{"[Create new project]"}, existingNames...)

			var selection string
			prompt := &survey.Select{
				Message: "Select a project or create a new one:",
				Options: options,
			}
			if err := survey.AskOne(prompt, &selection); err != nil {
				log.Error("Prompt failed: %s", err)
				return
			}

			if selection == "[Create new project]" {
				namePrompt := &survey.Input{
					Message: "Enter new project name:",
				}
				if err := survey.AskOne(namePrompt, &projectName, survey.WithValidator(survey.Required)); err != nil {
					log.Error("Prompt failed: %s", err)
					return
				}
			} else {
				projectName = selection
			}
		}

		// Get repository URL
		var repoURL string
		urlPrompt := &survey.Input{
			Message: "Enter repository URL (SSH or HTTPS):",
			Help:    "Examples: git@github.com:org/repo.git or https://github.com/org/repo.git",
		}
		if err := survey.AskOne(urlPrompt, &repoURL, survey.WithValidator(survey.Required)); err != nil {
			log.Error("Prompt failed: %s", err)
			return
		}

		// Derive default path from URL
		defaultPath, err := config.RepoNameFromURL(repoURL)
		if err != nil {
			log.Error("Could not parse repository URL: %s", err)
			return
		}

		// Ask for optional custom path
		var customPath string
		pathPrompt := &survey.Input{
			Message: "Custom directory name (leave empty for default):",
			Default: "",
			Help:    "Default: " + defaultPath,
		}
		if err := survey.AskOne(pathPrompt, &customPath); err != nil {
			log.Error("Prompt failed: %s", err)
			return
		}

		// Create the repo entry
		repo := config.ProjectRepo{
			URL:  repoURL,
			Path: customPath, // Empty string will use default
		}

		// Check if this is a new project
		existingProject := config.GetProjectByName(projectName)
		if existingProject == nil {
			// Create new project
			newProject := config.Project{
				Name:  projectName,
				Repos: []config.ProjectRepo{repo},
			}
			if err := config.AddProject(newProject); err != nil {
				log.Error("Failed to add project: %s", err)
				return
			}
			log.Success("Created new project '%s' with repository", projectName)
		} else {
			// Add to existing project
			if err := config.AddRepoToProject(projectName, repo); err != nil {
				log.Error("Failed to add repository: %s", err)
				return
			}
			log.Success("Added repository to project '%s'", projectName)
		}

		// Ask if they want to add more
		var addMore bool
		morePrompt := &survey.Confirm{
			Message: "Add another repository?",
			Default: false,
		}
		if err := survey.AskOne(morePrompt, &addMore); err != nil {
			log.Error("Prompt failed: %s", err)
			return
		}

		if addMore {
			// Re-run the add command for the same project
			_ = cmd.Flags().Set("project", projectName)
			cmd.Run(cmd, args)
			return
		}

		log.Info("")
		log.Info("Run 'eng project setup' to clone the new repositories.")

		// Show current project state
		updatedProjects := config.GetProjects()
		for _, p := range updatedProjects {
			if p.Name == projectName {
				log.Info("")
				log.Info("Project '%s' now has %d repository(ies):", p.Name, len(p.Repos))
				for _, r := range p.Repos {
					path, _ := r.GetEffectivePath()
					log.Info("  - %s", path)
				}
				break
			}
		}
	},
}
