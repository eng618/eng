package dashboard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/eng618/eng/internal/ui/theme"
)

func (m Model) View() string {
	if !m.ready {
		return "Initializing Dashboard..."
	}

	if m.windowWidth < 60 || m.windowHeight < 12 {
		return m.renderFallbackScreen()
	}

	if m.showHelp {
		modalContent := m.renderHelpModal()
		modal := helpModalStyle.Render(modalContent)
		return overlayStyle.
			Width(m.windowWidth).
			Height(m.windowHeight).
			Render(lipgloss.Place(m.windowWidth, m.windowHeight, lipgloss.Center, lipgloss.Center, modal, lipgloss.WithWhitespaceChars(" ")))
	}

	leftStyle := inactivePaneStyle
	rightStyle := inactivePaneStyle

	if m.focusedPane == FocusLeft {
		leftStyle = activePaneStyle
	} else {
		rightStyle = activePaneStyle
	}

	totalPanesWidth := m.windowWidth - 4
	leftPaneOuterWidth := totalPanesWidth / 4
	if leftPaneOuterWidth < 20 {
		leftPaneOuterWidth = 20
	}
	if leftPaneOuterWidth > 30 {
		leftPaneOuterWidth = 30
	}
	rightPaneOuterWidth := totalPanesWidth - leftPaneOuterWidth

	leftStyleWidth := leftPaneOuterWidth - 2
	rightStyleWidth := rightPaneOuterWidth - 2

	leftStyle = leftStyle.Width(leftStyleWidth).Height(m.windowHeight - 4)
	rightStyle = rightStyle.Width(rightStyleWidth).Height(m.windowHeight - 4)

	// Render Left Pane
	leftContent := leftStyle.Render(limitLines(m.list.View(), m.windowHeight-6))

	// Render Right Pane
	rightContent := rightStyle.Render(limitLines(m.renderRightPane(), m.windowHeight-6))

	// Combine panes
	mainView := appStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, leftContent, rightContent))

	if m.actionState != "" {
		// Render Modal Overlay
		var logLines string
		if len(m.actionLogs) > 0 {
			// Show up to the last 8 lines
			startIdx := 0
			if len(m.actionLogs) > 8 {
				startIdx = len(m.actionLogs) - 8
			}
			logLines = strings.Join(m.actionLogs[startIdx:], "\n")
		}

		var progressLine string
		if m.totalActions > 0 {
			pct := float64(m.completedActions) / float64(m.totalActions)
			if pct > 1.0 {
				pct = 1.0
			}
			progressBar := renderProgressBar(30, pct)
			progressInfo := fmt.Sprintf(
				"%d of %d repositories processed (%d%%)",
				m.completedActions,
				m.totalActions,
				int(pct*100),
			)
			progressLine = lipgloss.JoinVertical(lipgloss.Center,
				progressBar,
				progressInfoStyle.Render(progressInfo),
				"",
			)
		}

		modalContent := lipgloss.JoinVertical(lipgloss.Center,
			m.spinner.View(),
			"",
			projectNameStyle.Render(m.actionState),
			"",
			progressLine,
			logLines,
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

	totalPanesWidth := m.windowWidth - 4
	leftPaneOuterWidth := totalPanesWidth / 4
	if leftPaneOuterWidth < 20 {
		leftPaneOuterWidth = 20
	}
	if leftPaneOuterWidth > 30 {
		leftPaneOuterWidth = 30
	}
	rightPaneOuterWidth := totalPanesWidth - leftPaneOuterWidth
	rightStyleWidth := rightPaneOuterWidth - 2
	innerRightWidth := rightStyleWidth - 2

	projectName := fmt.Sprintf("Project: %s", p.Name)
	projectName = truncate(projectName, innerRightWidth)
	b.WriteString(projectNameStyle.Render(projectName))
	b.WriteString("\n\n")

	if len(p.Repos) == 0 {
		noReposStr := truncate("No repositories configured for this project.", innerRightWidth)
		b.WriteString(statusMutedStyle.Render(noReposStr))
		return b.String()
	}

	innerRightHeight := m.windowHeight - 6
	if innerRightHeight < 5 {
		return "Terminal too small"
	}
	H_repos := innerRightHeight - 4

	allLines, _, _ := m.getRepoLines()

	// Slice allLines based on m.repoScrollOffset and H_repos
	start := m.repoScrollOffset
	end := start + H_repos
	if end > len(allLines) {
		end = len(allLines)
	}

	for i := start; i < end; i++ {
		b.WriteString(allLines[i])
		b.WriteString("\n")
	}

	// Pad with empty lines if needed to keep the help/footer sticky at the bottom
	renderedCount := end - start
	if renderedCount < H_repos {
		for i := 0; i < H_repos-renderedCount; i++ {
			b.WriteString("\n")
		}
	}

	var footerText string
	if m.notification != "" {
		var prefix string
		switch m.notificationType {
		case NotifySuccess:
			prefix = "✓ "
		case NotifyError:
			prefix = "✗ "
		case NotifyWarn:
			prefix = "⚠ "
		}
		footerText = prefix + m.notification
		footerText = truncate(footerText, innerRightWidth)
		b.WriteString("\n")
		b.WriteString(m.notificationStyle.Render(footerText))
	} else {
		if m.focusedPane == FocusRight {
			footerText = "[j/k] Navigate  [f] Fetch  [p] Pull  [s] Sync  [c] Clone  [o] Open  [e] Editor  [?] Help  [Esc] Back"
		} else {
			footerText = "[Enter/l] Focus Repositories  [f] Fetch All  [p] Pull All  [s] Sync All  [c] Setup All  [?] Help"
		}
		footerText = truncate(footerText, innerRightWidth)
		b.WriteString(statusMutedStyle.Render("\n" + footerText))
	}

	return b.String()
}

func (m Model) getRepoLines() (allLines []string, repoStarts, repoEnds []int) {
	item, ok := m.list.SelectedItem().(ProjectItem)
	if !ok {
		return nil, nil, nil
	}
	p := item.Project

	totalPanesWidth := m.windowWidth - 4
	leftPaneOuterWidth := totalPanesWidth / 4
	if leftPaneOuterWidth < 20 {
		leftPaneOuterWidth = 20
	}
	if leftPaneOuterWidth > 30 {
		leftPaneOuterWidth = 30
	}
	rightPaneOuterWidth := totalPanesWidth - leftPaneOuterWidth
	rightStyleWidth := rightPaneOuterWidth - 2
	innerRightWidth := rightStyleWidth - 2

	repoStarts = make([]int, len(p.Repos))
	repoEnds = make([]int, len(p.Repos))

	for i, r := range p.Repos {
		var repoLines []string

		repoName, err := r.GetEffectivePath()
		if err != nil {
			repoName = r.URL
		}
		repoTitle := fmt.Sprintf("repo: %s", repoName)
		repoTitle = truncate(repoTitle, innerRightWidth)

		var titleLine string
		if m.focusedPane == FocusRight && i == m.selectedRepoIndex {
			titleLine = selectedRepoStyle.Render(repoTitle)
		} else {
			titleLine = repoNameStyle.Render(repoTitle)
		}
		repoLines = append(repoLines, titleLine)

		key := p.Name + r.URL
		status, exists := m.repoStatuses[key]

		if !exists || status.Loading {
			checkingStr := truncate("  [ Checking status... ]", innerRightWidth)
			repoLines = append(repoLines, statusMutedStyle.Render(checkingStr))
			repoLines = append(repoLines, "")
		} else if status.Error != nil {
			errStr := fmt.Sprintf("  ✗ Error: %s", status.Error.Error())
			errStr = truncate(errStr, innerRightWidth)
			repoLines = append(repoLines, statusErrorStyle.Render(errStr))
			repoLines = append(repoLines, "")
		} else if !status.IsCloned {
			missingStr := truncate("  ✗ Missing (Not Cloned)", innerRightWidth)
			repoLines = append(repoLines, statusErrorStyle.Render(missingStr))
			repoLines = append(repoLines, "")
		} else {
			clonedStr := truncate("  ✓ Cloned", innerRightWidth)
			repoLines = append(repoLines, statusSuccessStyle.Render(clonedStr))

			branchColor := statusMutedStyle
			if status.Branch == "main" || status.Branch == "master" {
				branchColor = statusSuccessStyle
			}
			branchName := truncate(status.Branch, innerRightWidth-10) // 10 chars for "  branch: "
			repoLines = append(repoLines, fmt.Sprintf("  branch: %s", branchColor.Render(branchName)))

			statusText := "  status: Clean"
			if status.IsDirty {
				statusText = "  status: Uncommitted changes!"
			}
			statusText = truncate(statusText, innerRightWidth)
			if status.IsDirty {
				repoLines = append(repoLines, statusWarningStyle.Render(statusText))
			} else {
				repoLines = append(repoLines, statusSuccessStyle.Render(statusText))
			}
			repoLines = append(repoLines, "")
		}

		repoStarts[i] = len(allLines)
		allLines = append(allLines, repoLines...)
		repoEnds[i] = len(allLines) - 1
	}

	return allLines, repoStarts, repoEnds
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen < 3 {
		if maxLen < 0 {
			return ""
		}
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

func limitLines(s string, maxLines int) string {
	if maxLines <= 0 {
		return ""
	}
	lines := strings.Split(s, "\n")
	if len(lines) <= maxLines {
		return s
	}
	return strings.Join(lines[:maxLines], "\n")
}

func (m Model) renderFallbackScreen() string {
	msg := fmt.Sprintf(
		"Terminal Too Small\n\nWidth: %d/60, Height: %d/12\n\nPlease resize your window or\npress [q] or [Ctrl+C] to quit.",
		m.windowWidth,
		m.windowHeight,
	)

	return lipgloss.Place(
		m.windowWidth,
		m.windowHeight,
		lipgloss.Center,
		lipgloss.Center,
		modalStyle.Render(msg),
		lipgloss.WithWhitespaceChars(" "),
	)
}

func (m Model) renderHelpModal() string {
	var b strings.Builder
	b.WriteString(projectNameStyle.Render("Keyboard Shortcuts"))
	b.WriteString("\n\n")

	keyStyle := lipgloss.NewStyle().Foreground(theme.Primary).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(theme.Foreground)

	fmt.Fprintf(&b, "  %s   %s\n", keyStyle.Render("?     "), descStyle.Render("Toggle Help Menu"))
	fmt.Fprintf(&b, "  %s   %s\n\n", keyStyle.Render("q/Ctrl+C"), descStyle.Render("Quit Application"))

	b.WriteString(statusMutedStyle.Render("Navigation:"))
	b.WriteString("\n\n")
	fmt.Fprintf(&b, "  %s   %s\n", keyStyle.Render("h/Left"), descStyle.Render("Focus Projects Pane (Left)"))
	fmt.Fprintf(&b, "  %s   %s\n", keyStyle.Render("l/Right"), descStyle.Render("Focus Repositories Pane (Right)"))
	fmt.Fprintf(&b, "  %s   %s\n\n", keyStyle.Render("j/k/Up/Down"), descStyle.Render("Navigate Lists"))

	b.WriteString(statusMutedStyle.Render("Actions (Context-aware):"))
	b.WriteString("\n\n")
	fmt.Fprintf(&b, "  %s   %s\n", keyStyle.Render("f     "), descStyle.Render("Fetch repository (or all)"))
	fmt.Fprintf(&b, "  %s   %s\n", keyStyle.Render("p     "), descStyle.Render("Pull repository (or all)"))
	fmt.Fprintf(&b, "  %s   %s\n", keyStyle.Render("s     "), descStyle.Render("Sync repository (or all)"))
	fmt.Fprintf(&b, "  %s   %s\n", keyStyle.Render("c     "), descStyle.Render("Clone/Setup repository (or all)"))
	fmt.Fprintf(&b, "  %s   %s\n", keyStyle.Render("o     "), descStyle.Render("Open in Finder / File Explorer"))
	fmt.Fprintf(&b, "  %s   %s\n\n", keyStyle.Render("e     "), descStyle.Render("Open in Configured Editor"))

	b.WriteString(statusMutedStyle.Render("Press any key to close"))

	return b.String()
}

func renderProgressBar(width int, percentage float64) string {
	if width <= 0 {
		return ""
	}
	filledLength := int(percentage * float64(width))
	if filledLength > width {
		filledLength = width
	}
	if filledLength < 0 {
		filledLength = 0
	}

	filled := strings.Repeat("█", filledLength)
	empty := strings.Repeat("░", width-filledLength)

	return progressBarFilledStyle.Render(filled) + progressBarTrackStyle.Render(empty)
}
