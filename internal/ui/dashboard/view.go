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

	leftStyle := inactivePaneStyle
	rightStyle := inactivePaneStyle

	if m.focusedPane == FocusLeft {
		leftStyle = activePaneStyle
	} else {
		rightStyle = activePaneStyle
	}

	leftWidth := (m.windowWidth / 3) - 4
	rightWidth := m.windowWidth - leftWidth - 8

	leftStyle = leftStyle.Width(leftWidth).Height(m.windowHeight - 4)
	rightStyle = rightStyle.Width(rightWidth).Height(m.windowHeight - 4)

	// Render Left Pane
	leftContent := leftStyle.Render(m.list.View())

	// Render Right Pane
	rightContent := rightStyle.Render(m.renderRightPane())

	// Combine panes
	mainView := appStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, leftContent, rightContent))

	if m.actionState != "" {
		// Render Modal Overlay
		modalContent := lipgloss.JoinVertical(lipgloss.Center,
			m.spinner.View(),
			"",
			projectNameStyle.Render(m.actionState),
		)
		
		modal := modalStyle.Render(modalContent)

		// Place the modal in the center of the main view
		return overlayStyle.
			Width(m.windowWidth).
			Height(m.windowHeight).
			Render(lipgloss.Place(m.windowWidth, m.windowHeight, lipgloss.Center, lipgloss.Center, modal, lipgloss.WithWhitespaceChars(" ")))
	}

	return mainView
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

	for i, r := range p.Repos {
		// Highlight if focused on right pane and this is the selected repo
		repoTitle := fmt.Sprintf("repo: %s", r.URL)
		if m.focusedPane == FocusRight && i == m.selectedRepoIndex {
			b.WriteString(selectedRepoStyle.Render(repoTitle))
		} else {
			b.WriteString(repoNameStyle.Render(repoTitle))
		}
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

	if m.focusedPane == FocusRight {
		b.WriteString(statusMutedStyle.Render("\n[j/k] Navigate  [f] Fetch  [p] Pull  [s] Sync  [c] Clone  [o] Open  [Esc] Back"))
	} else {
		b.WriteString(statusMutedStyle.Render("\n[Enter/l] Focus Repositories  [f] Fetch All  [p] Pull All  [s] Sync All  [c] Setup All"))
	}

	return b.String()
}
