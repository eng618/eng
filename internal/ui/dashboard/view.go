package dashboard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/eng618/eng/internal/repo"
	"github.com/eng618/eng/internal/ui"
)

func (m Model) View() string {
	if !m.ready {
		return m.spinner.View() + " Initializing Dashboard…"
	}

	if m.isFallbackMode() {
		return m.renderFallbackScreen()
	}

	if m.showHelp {
		modalContent := limitLines(m.renderHelpModal(), m.windowHeight-4)
		modal := helpModalStyle.Render(modalContent)
		return overlayStyle.
			Width(m.windowWidth).
			Height(m.windowHeight).
			Render(lipgloss.Place(m.windowWidth, m.windowHeight, lipgloss.Center, lipgloss.Center, modal, lipgloss.WithWhitespaceChars(" ")))
	}

	var mainView string
	if m.isCompactMode() {
		mainView = m.renderCompactDashboard()
	} else {
		mainView = m.renderFullDashboard()
	}

	if m.actionState != "" {
		// Fixed-size Docker-like progress modal: box dimensions derive only
		// from the window size, never from repo-name or log-line lengths.
		return overlayStyle.
			Width(m.windowWidth).
			Height(m.windowHeight).
			Render(lipgloss.Place(m.windowWidth, m.windowHeight, lipgloss.Center, lipgloss.Center, m.renderActionModal(), lipgloss.WithWhitespaceChars(" ")))
	}

	return mainView
}

// actionModalSize returns fixed dimensions derived from the window only.
func (m Model) renderFullDashboard() string {
	leftStyle := inactivePaneStyle
	rightStyle := inactivePaneStyle

	if m.focusedPane == FocusLeft {
		leftStyle = activePaneStyle
		m.list.Title = "● Projects"
	} else {
		rightStyle = activePaneStyle
		m.list.Title = "Projects"
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
	return appStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, leftContent, rightContent))
}

func (m Model) renderCompactDashboard() string {
	var b strings.Builder

	// Render Tab Header
	var tab1, tab2 string
	item, ok := m.list.SelectedItem().(ProjectItem)
	projName := "Repos"
	if ok {
		projName = fmt.Sprintf("Repos (%s)", ui.Truncate(item.Project.Name, 15))
	}

	if m.focusedPane == FocusLeft {
		tab1 = tabActiveStyle.Render("[1: Projects]")
		tab2 = tabInactiveStyle.Render("2: " + projName)
	} else {
		tab1 = tabInactiveStyle.Render("1: Projects")
		tab2 = tabActiveStyle.Render("[2: " + projName + "]")
	}

	tabBar := lipgloss.JoinHorizontal(lipgloss.Left, tab1, "  ", tab2)
	b.WriteString(tabBar)
	b.WriteString("\n")

	contentHeight := m.windowHeight - 6
	if contentHeight < 1 {
		contentHeight = 1
	}

	if m.focusedPane == FocusLeft {
		b.WriteString(limitLines(m.list.View(), contentHeight))
		b.WriteString("\n")
		b.WriteString(ui.Truncate("[1/2] Switch tabs  [/] Filter  [a] Add  [?] Help", m.windowWidth-8))
	} else {
		b.WriteString(limitLines(m.renderRightPane(), contentHeight))
	}

	paneWidth := m.windowWidth - 4
	if paneWidth < 10 {
		paneWidth = 10
	}

	pane := compactPaneStyle.
		Width(paneWidth).
		Height(m.windowHeight - 2).
		Render(b.String())

	return appStyle.Render(pane)
}

func (m Model) renderRightPane() string {
	item, ok := m.list.SelectedItem().(ProjectItem)
	if !ok {
		return "No project selected.\n\nPress / to filter, Esc to clear filter."
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
	projectName = ui.Truncate(projectName, innerRightWidth)
	if m.focusedPane == FocusRight {
		projectName = "● " + projectName
	}
	b.WriteString(projectNameStyle.Render(projectName))
	b.WriteString("\n\n")

	if len(p.Repos) == 0 {
		noReposStr := ui.Truncate("No repositories yet — press [a] to add one.", innerRightWidth)
		b.WriteString(statusMutedStyle.Render(noReposStr))
		b.WriteString("\n")
		b.WriteString(m.renderFooter(innerRightWidth))
		return b.String()
	}

	innerRightHeight := m.windowHeight - 6
	if innerRightHeight < 5 {
		return "Terminal too small"
	}
	H_repos := innerRightHeight - 4

	if innerRightWidth >= 75 {
		b.WriteString(m.renderRepoTable(p, innerRightWidth, H_repos))
	} else {
		allLines, _, _ := m.getRepoLines()

		// Slice allLines based on m.repoScrollOffset and H_repos
		start := m.repoScrollOffset
		end := start + H_repos
		if end > len(allLines) {
			end = len(allLines)
		}

		var lines []string
		for i := start; i < end; i++ {
			lines = append(lines, allLines[i])
		}

		// Pad with empty lines if needed to keep the help/footer sticky at the bottom
		renderedCount := end - start
		if renderedCount < H_repos {
			for i := 0; i < H_repos-renderedCount; i++ {
				lines = append(lines, "")
			}
		}

		b.WriteString(strings.Join(lines, "\n"))
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
		if m.notificationType == NotifyError {
			footerText += "  [r] Retry"
		}
		footerText = ui.Truncate(footerText, innerRightWidth)
		b.WriteString("\n")
		b.WriteString(m.notificationStyle.Render(footerText))
	} else {
		b.WriteString("\n")
		b.WriteString(m.renderFooter(innerRightWidth))
	}

	return b.String()
}

// renderFooter builds the context-aware hint bar. It always mentions filter
// (/) and help (?) so key discovery doesn't depend on terminal width.
func (m Model) renderFooter(innerRightWidth int) string {
	var footerText string
	if m.focusedPane == FocusRight {
		footerText = "[j/k] Navigate  [f] Fetch  [p] Pull  [s] Sync  [c] Clone  [o] Open  [e/E] Edit  [t] Term  [r] Refresh  [a] Add  [/] Filter  [?] Help  [Esc] Back"
	} else {
		footerText = "[Enter/l] Focus  [f] Fetch All  [p] Pull All  [s] Sync All  [e/E] Edit  [t] Term  [r] Refresh  [a] Add  [/] Filter  [?] Help"
	}
	if scrollHint := m.scrollIndicator(); scrollHint != "" {
		footerText += "  " + scrollHint
	}
	footerText = ui.Truncate(footerText, innerRightWidth)
	return statusMutedStyle.Render(footerText)
}

// scrollIndicator reports position like "3/12" plus ▲▼ when content overflows.
func (m Model) scrollIndicator() string {
	item, ok := m.list.SelectedItem().(ProjectItem)
	if !ok || len(item.Project.Repos) <= 1 {
		return ""
	}
	total := len(item.Project.Repos)
	current := m.selectedRepoIndex + 1
	if current < 1 {
		current = 1
	}
	if current > total {
		current = total
	}
	if m.repoScrollOffset > 0 || total > 5 {
		return fmt.Sprintf("%d/%d ▲▼", current, total)
	}
	return fmt.Sprintf("%d/%d", current, total)
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

		repoName, err := repo.EffectivePath(r.URL, r.Path)
		if err != nil {
			repoName = r.URL
		}
		repoTitle := fmt.Sprintf("repo: %s", repoName)
		repoTitle = ui.Truncate(repoTitle, innerRightWidth)

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
			checkingStr := ui.Truncate("  ↻ Checking status…", innerRightWidth)
			repoLines = append(repoLines, statusMutedStyle.Render(checkingStr))
			repoLines = append(repoLines, "")
		} else if status.Error != nil {
			errStr := fmt.Sprintf("  ✗ Error: %s  [r] Retry", status.Error.Error())
			errStr = ui.Truncate(errStr, innerRightWidth)
			repoLines = append(repoLines, statusErrorStyle.Render(errStr))
			repoLines = append(repoLines, "")
		} else if !status.IsCloned {
			missingStr := ui.Truncate("  ✗ Missing (Not Cloned)", innerRightWidth)
			repoLines = append(repoLines, statusErrorStyle.Render(missingStr))
			repoLines = append(repoLines, "")
		} else {
			clonedStr := ui.Truncate("  ✓ Cloned", innerRightWidth)
			repoLines = append(repoLines, statusSuccessStyle.Render(clonedStr))

			var branchText string
			branchColor := statusMutedStyle

			if status.IsDetached {
				branchText = status.Branch
				branchColor = statusWarningStyle
			} else {
				branchText = status.Branch
				if status.HasUpstream {
					var parts []string
					if status.AheadCount > 0 {
						parts = append(parts, fmt.Sprintf("↑%d", status.AheadCount))
					}
					if status.BehindCount > 0 {
						parts = append(parts, fmt.Sprintf("↓%d", status.BehindCount))
					}
					if len(parts) > 0 {
						branchText = fmt.Sprintf("%s %s", status.Branch, strings.Join(parts, " "))
						if status.AheadCount > 0 && status.BehindCount > 0 {
							branchColor = statusWarningStyle
						} else if status.BehindCount > 0 {
							branchColor = statusErrorStyle
						} else {
							branchColor = statusSuccessStyle
						}
					} else {
						if status.Branch == "main" || status.Branch == "master" {
							branchColor = statusSuccessStyle
						}
					}
				} else {
					branchText = fmt.Sprintf("%s (unpublished)", status.Branch)
					branchColor = statusMutedStyle
				}
			}

			branchNameLine := ui.Truncate(branchText, innerRightWidth-10) // 10 chars for "  branch: "
			repoLines = append(repoLines, fmt.Sprintf("  branch: %s", branchColor.Render(branchNameLine)))

			var statusText string
			statusColor := statusSuccessStyle

			switch {
			case status.OngoingOp != "":
				statusText = fmt.Sprintf("Ongoing %s!", status.OngoingOp)
				statusColor = statusWarningStyle
			case status.ConflictCount > 0:
				statusText = fmt.Sprintf("Merge conflicts! (%d files)", status.ConflictCount)
				statusColor = statusErrorStyle
			case status.UnstagedCount > 0 || status.StagedCount > 0 || status.UntrackedCount > 0:
				var parts []string
				if status.UnstagedCount > 0 {
					parts = append(parts, fmt.Sprintf("%d modified", status.UnstagedCount))
					statusColor = statusWarningStyle
				}
				if status.StagedCount > 0 {
					parts = append(parts, fmt.Sprintf("%d staged", status.StagedCount))
					if status.UnstagedCount == 0 {
						statusColor = statusSuccessStyle
					}
				}
				if status.UntrackedCount > 0 {
					parts = append(parts, fmt.Sprintf("%d untracked", status.UntrackedCount))
					if status.UnstagedCount == 0 && status.StagedCount == 0 {
						statusColor = statusMutedStyle
					}
				}
				statusText = strings.Join(parts, ", ")
			default:
				statusText = "Clean"
				statusColor = statusSuccessStyle
			}

			statusLine := fmt.Sprintf("  status: %s", statusText)
			statusLine = ui.Truncate(statusLine, innerRightWidth)
			repoLines = append(repoLines, statusColor.Render(statusLine))
			repoLines = append(repoLines, "")
		}

		repoStarts[i] = len(allLines)
		allLines = append(allLines, repoLines...)
		repoEnds[i] = len(allLines) - 1
	}

	return allLines, repoStarts, repoEnds
}

// relativeTime renders "just now", "2m ago", "3h ago", falling back to date.
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
