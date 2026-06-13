package dashboard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if !m.ready {
		return "Initializing Dashboard..."
	}

	// Render Left Pane
	leftContent := leftPaneStyle.Render(m.list.View())

	// Render Right Pane
	rightContent := rightPaneStyle.Render(m.renderRightPane())

	return appStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, leftContent, rightContent))
}

func (m Model) renderRightPane() string {
	item, ok := m.list.SelectedItem().(ProjectItem)
	if !ok {
		return "No project selected."
	}
	p := item.Project

	var b strings.Builder

	b.WriteString(projectNameStyle.Render(fmt.Sprintf("Project: %s", p.Name)))
	b.WriteString("\n\n")

	if len(p.Repos) == 0 {
		b.WriteString(statusMutedStyle.Render("No repositories configured for this project."))
		return b.String()
	}

	for _, r := range p.Repos {
		b.WriteString(repoNameStyle.Render(fmt.Sprintf("repo: %s", r.URL)))
		b.WriteString("\n")

		key := p.Name + r.URL
		status, exists := m.repoStatuses[key]

		if !exists || status.Loading {
			b.WriteString(statusMutedStyle.Render("  [ Checking status... ]\n\n"))
			continue
		}

		if status.Error != nil {
			b.WriteString(statusErrorStyle.Render(fmt.Sprintf("  ✗ Error: %s\n\n", status.Error.Error())))
			continue
		}

		if !status.IsCloned {
			b.WriteString(statusErrorStyle.Render("  ✗ Missing (Not Cloned)\n\n"))
			continue
		}

		b.WriteString(statusSuccessStyle.Render("  ✓ Cloned"))
		b.WriteString("\n")

		branchColor := statusMutedStyle
		if status.Branch == "main" || status.Branch == "master" {
			branchColor = statusSuccessStyle
		}
		b.WriteString(fmt.Sprintf("  branch: %s\n", branchColor.Render(status.Branch)))

		if status.IsDirty {
			b.WriteString(statusWarningStyle.Render("  status: Uncommitted changes!\n"))
		} else {
			b.WriteString(statusSuccessStyle.Render("  status: Clean\n"))
		}

		b.WriteString("\n")
	}

	return b.String()
}
