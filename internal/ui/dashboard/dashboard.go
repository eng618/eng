package dashboard

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/eng618/eng/internal/ui/theme"
)

// Run launches the interactive project and git dashboard.
//
// Projects are passed in as plain DTOs plus a provider for reloading them
// after mutations, so this package never reads configuration storage.
// Callers in cmd/ adapt config types (see FromConfigProjects) before calling.
func Run(projects []Project, devPath, editor string, provider ProjectProvider) error {
	if len(projects) == 0 {
		return theme.NewActionableError(
			fmt.Errorf("no projects configured"),
			"Use 'eng project add' to add a project before opening the dashboard.",
		)
	}

	m := NewModel(projects, devPath, editor, provider)

	// Alternate screen + mouse support so clicks can focus panes/select repos.
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run dashboard: %w", err)
	}

	return nil
}
