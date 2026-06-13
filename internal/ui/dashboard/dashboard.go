package dashboard

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/ui/theme"
)

// Run launches the interactive project and git dashboard.
func Run() error {
	projects := config.GetProjects()
	if len(projects) == 0 {
		return theme.NewActionableError(
			fmt.Errorf("no projects configured"),
			"Use 'eng project add' to add a project before opening the dashboard.",
		)
	}

	gitDevPath := config.GetGitConfig().DevPath
	if gitDevPath == "" {
		home, _ := os.UserHomeDir()
		gitDevPath = home + "/Development"
	}
	gitDevPath = os.ExpandEnv(gitDevPath)

	m := NewModel(projects, gitDevPath)

	// Use alternate screen buffer
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run dashboard: %w", err)
	}

	return nil
}
