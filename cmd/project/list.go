package project

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
	internalProject "github.com/eng618/eng/internal/project"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

// ListCmd defines the cobra command for listing configured projects.
// It displays project names, repository counts, and clone status.
var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured projects and their repositories",
	Long: `This command displays all configured projects and their repositories.

Use the --verbose flag to see detailed information including:
  - Repository URLs
  - Clone status (✓ cloned / ✗ missing)
  - Local paths

Example:
  eng project list               # Show projects summary
  eng project list -v            # Show detailed repository information
  eng project list -p MyProject  # Show only the specified project`,
	Run: func(cmd *cobra.Command, args []string) {
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary).
			MarginBottom(1)
		if !ui.DisableProgress {
			fmt.Fprintln(log.Out, headerStyle.Render("📌 Project Repository Collections"))
		}

		gitCfg := config.GetGitConfig()
		projectFilter, _ := cmd.Flags().GetString("project")

		opts := internalProject.ListOptions{
			IsVerbose:     cmdutil.IsVerbose(cmd),
			ProjectFilter: projectFilter,
			DevPath:       gitCfg.DevPath,
			Projects:      config.GetProjects(),
		}

		internalProject.List(opts)
	},
}
