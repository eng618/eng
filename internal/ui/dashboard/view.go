package dashboard

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/ui/theme"
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
		projName = fmt.Sprintf("Repos (%s)", truncate(item.Project.Name, 15))
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
		b.WriteString(truncate("[1/2] Switch tabs  [/] Filter  [a] Add  [?] Help", m.windowWidth-8))
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
	projectName = truncate(projectName, innerRightWidth)
	if m.focusedPane == FocusRight {
		projectName = "● " + projectName
	}
	b.WriteString(projectNameStyle.Render(projectName))
	b.WriteString("\n\n")

	if len(p.Repos) == 0 {
		noReposStr := truncate("No repositories yet — press [a] to add one.", innerRightWidth)
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
		footerText = truncate(footerText, innerRightWidth)
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
	footerText = truncate(footerText, innerRightWidth)
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
			checkingStr := truncate("  ↻ Checking status…", innerRightWidth)
			repoLines = append(repoLines, statusMutedStyle.Render(checkingStr))
			repoLines = append(repoLines, "")
		} else if status.Error != nil {
			errStr := fmt.Sprintf("  ✗ Error: %s  [r] Retry", status.Error.Error())
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

			branchNameLine := truncate(branchText, innerRightWidth-10) // 10 chars for "  branch: "
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
			statusLine = truncate(statusLine, innerRightWidth)
			repoLines = append(repoLines, statusColor.Render(statusLine))
			repoLines = append(repoLines, "")
		}

		repoStarts[i] = len(allLines)
		allLines = append(allLines, repoLines...)
		repoEnds[i] = len(allLines) - 1
	}

	return allLines, repoStarts, repoEnds
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen < 3 {
		if maxLen < 0 {
			return ""
		}
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}

// relativeTime renders "just now", "2m ago", "3h ago", falling back to date.
func relativeTime(t time.Time) string {
	d := time.Since(t)
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	if d < 7*24*time.Hour {
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
	return t.Format("2006-01-02")
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
		"Terminal Too Small\n\nExpand your window to view the dashboard.\n\nWidth: %d/50, Height: %d/10\n\nPlease resize your window or\npress [q] or [Ctrl+C] to quit.",
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

	fmt.Fprintf(&b, "  %s   %s\n", keyStyle.Render("?/Esc "), descStyle.Render("Open / close this help"))
	fmt.Fprintf(&b, "  %s   %s\n", keyStyle.Render("q     "), descStyle.Render("Quit (Ctrl+C always quits)"))
	fmt.Fprintf(&b, "  %s   %s\n\n", keyStyle.Render("1/2   "), descStyle.Render("Switch Projects / Repos tabs"))

	b.WriteString(statusMutedStyle.Render("Navigation:"))
	b.WriteString("\n\n")
	fmt.Fprintf(&b, "  %s   %s\n", keyStyle.Render("Tab   "), descStyle.Render("Switch focused pane"))
	fmt.Fprintf(&b, "  %s   %s\n", keyStyle.Render("Enter/l"), descStyle.Render("Focus Repositories pane"))
	fmt.Fprintf(&b, "  %s   %s\n", keyStyle.Render("Esc/h "), descStyle.Render("Back to Projects pane"))
	fmt.Fprintf(&b, "  %s   %s\n", keyStyle.Render("/     "), descStyle.Render("Filter projects (Esc clears)"))
	fmt.Fprintf(&b, "  %s   %s\n\n", keyStyle.Render("j/k   "), descStyle.Render("Navigate repositories"))

	b.WriteString(statusMutedStyle.Render("Actions (context-aware: single repo vs whole project):"))
	b.WriteString("\n\n")
	fmt.Fprintf(&b, "  %s   %s\n", keyStyle.Render("f     "), descStyle.Render("Fetch (git fetch --all --prune)"))
	fmt.Fprintf(&b, "  %s   %s\n", keyStyle.Render("p     "), descStyle.Render("Pull (git pull)"))
	fmt.Fprintf(&b, "  %s   %s\n", keyStyle.Render("s     "), descStyle.Render("Sync (stash, pull --rebase, pop)"))
	fmt.Fprintf(&b, "  %s   %s\n", keyStyle.Render("c     "), descStyle.Render("Clone missing repository"))
	fmt.Fprintf(&b, "  %s   %s\n", keyStyle.Render("o     "), descStyle.Render("Open in Finder / File Explorer"))
	fmt.Fprintf(&b, "  %s   %s\n", keyStyle.Render("e     "), descStyle.Render("Open in configured editor"))
	fmt.Fprintf(&b, "  %s   %s\n", keyStyle.Render("E     "), descStyle.Render("Choose editor to open in…"))
	fmt.Fprintf(&b, "  %s   %s\n", keyStyle.Render("t     "), descStyle.Render("Open new terminal here"))
	fmt.Fprintf(&b, "  %s   %s\n", keyStyle.Render("r     "), descStyle.Render("Refresh statuses ([r] retries errors)"))
	fmt.Fprintf(&b, "  %s   %s\n\n", keyStyle.Render("a     "), descStyle.Render("Add project or repository"))

	b.WriteString(statusMutedStyle.Render("Press Esc, ?, or q to close"))

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

func (m Model) renderRepoTable(p config.Project, innerRightWidth, H_repos int) string {
	gap := "  "
	gapWidth := len(gap)
	totalGaps := 4
	usableWidth := innerRightWidth - (gapWidth * totalGaps)
	if usableWidth < 40 {
		usableWidth = 40
	}

	wRepo := int(float64(usableWidth) * 0.28)
	wBranch := int(float64(usableWidth) * 0.20)
	wStatus := int(float64(usableWidth) * 0.27)
	wUpstream := int(float64(usableWidth) * 0.13)
	wUpdated := usableWidth - wRepo - wBranch - wStatus - wUpstream

	H_body := H_repos - 2
	if H_body < 1 {
		H_body = 1
	}

	headerRepo := fmt.Sprintf("  %-*s", wRepo-2, "REPOSITORY")
	headerBranch := fmt.Sprintf("%-*s", wBranch, "BRANCH")
	headerStatus := fmt.Sprintf("%-*s", wStatus, "STATUS")
	headerUpstream := fmt.Sprintf("%-*s", wUpstream, "UPSTREAM")
	headerUpdated := fmt.Sprintf("%*s", wUpdated, "UPDATED")

	headerLine := headerRepo + gap + headerBranch + gap + headerStatus + gap + headerUpstream + gap + headerUpdated
	tableHeaderStyle := lipgloss.NewStyle().Foreground(theme.Primary).Bold(true)

	underlineLine := strings.Repeat("─", wRepo) + gap +
		strings.Repeat("─", wBranch) + gap +
		strings.Repeat("─", wStatus) + gap +
		strings.Repeat("─", wUpstream) + gap +
		strings.Repeat("─", wUpdated)

	var lines []string
	lines = append(lines, tableHeaderStyle.Render(headerLine))
	lines = append(lines, statusMutedStyle.Render(underlineLine))

	start := m.repoScrollOffset
	end := start + H_body
	if end > len(p.Repos) {
		end = len(p.Repos)
	}

	for idx := start; idx < end; idx++ {
		r := p.Repos[idx]
		key := p.Name + r.URL
		status := m.repoStatuses[key]

		repoName, err := r.GetEffectivePath()
		if err != nil {
			repoName = r.URL
		}

		isSelected := m.focusedPane == FocusRight && idx == m.selectedRepoIndex

		repoCell := renderRepoCell(repoName, isSelected, wRepo)
		branchCell := renderBranchCell(status, isSelected, wBranch)
		statusCell := renderStatusCell(status, isSelected, wStatus)
		upstreamCell := renderUpstreamCell(status, isSelected, wUpstream)
		updatedCell := renderUpdatedCell(status, isSelected, wUpdated)

		rowLine := repoCell + gap + branchCell + gap + statusCell + gap + upstreamCell + gap + updatedCell
		lines = append(lines, rowLine)
	}

	renderedCount := end - start
	if renderedCount < H_body {
		for i := 0; i < H_body-renderedCount; i++ {
			lines = append(lines, "")
		}
	}

	return strings.Join(lines, "\n")
}

func renderRepoCell(repoName string, isSelected bool, wRepo int) string {
	var repoText string
	if isSelected {
		repoText = fmt.Sprintf("▸ %-*s", wRepo-2, truncate(repoName, wRepo-2))
		return selectedTableCellStyle.Render(repoText)
	}
	repoText = fmt.Sprintf("  %-*s", wRepo-2, truncate(repoName, wRepo-2))
	return repoNameStyle.Render(repoText)
}

func renderBranchCell(status RepoStatus, isSelected bool, wBranch int) string {
	var branchText string
	branchColor := statusMutedStyle

	if !status.IsCloned {
		branchText = "—"
	} else if status.IsDetached {
		branchText = truncate(status.Branch, wBranch-1)
		branchColor = statusWarningStyle
	} else {
		branchText = truncate(status.Branch, wBranch-1)
		if status.HasUpstream {
			if status.AheadCount > 0 && status.BehindCount > 0 {
				branchColor = statusWarningStyle
			} else if status.BehindCount > 0 {
				branchColor = statusErrorStyle
			} else {
				branchColor = statusSuccessStyle
			}
		} else {
			branchColor = statusMutedStyle
		}
	}

	branchTextFormatted := fmt.Sprintf("%-*s", wBranch, branchText)
	if isSelected {
		return selectedTableCellStyle.Render(branchTextFormatted)
	}
	return branchColor.Render(branchTextFormatted)
}

func renderStatusCell(status RepoStatus, isSelected bool, wStatus int) string {
	var statusText string
	statusColor := statusSuccessStyle

	if !status.IsCloned {
		statusText = "Missing"
		statusColor = statusErrorStyle
	} else if status.Loading {
		statusText = "↻ Checking…"
		statusColor = statusMutedStyle
	} else if status.Error != nil {
		statusText = "Error"
		statusColor = statusErrorStyle
	} else if status.OngoingOp != "" {
		statusText = fmt.Sprintf("Ongoing %s!", status.OngoingOp)
		statusColor = statusWarningStyle
	} else if status.ConflictCount > 0 {
		statusText = "Conflict!"
		statusColor = statusErrorStyle
	} else if status.UnstagedCount > 0 || status.StagedCount > 0 || status.UntrackedCount > 0 {
		var parts []string
		if status.UnstagedCount > 0 {
			parts = append(parts, fmt.Sprintf("%d mod", status.UnstagedCount))
			statusColor = statusWarningStyle
		}
		if status.StagedCount > 0 {
			parts = append(parts, fmt.Sprintf("%d stg", status.StagedCount))
			if status.UnstagedCount == 0 {
				statusColor = statusSuccessStyle
			}
		}
		if status.UntrackedCount > 0 {
			parts = append(parts, fmt.Sprintf("%d unt", status.UntrackedCount))
			if status.UnstagedCount == 0 && status.StagedCount == 0 {
				statusColor = statusMutedStyle
			}
		}
		statusText = strings.Join(parts, ", ")
	} else {
		statusText = "Clean"
		statusColor = statusSuccessStyle
	}

	statusTextFormatted := fmt.Sprintf("%-*s", wStatus, truncate(statusText, wStatus-1))
	if isSelected {
		return selectedTableCellStyle.Render(statusTextFormatted)
	}
	return statusColor.Render(statusTextFormatted)
}

func renderUpstreamCell(status RepoStatus, isSelected bool, wUpstream int) string {
	var upstreamText string
	upstreamColor := statusMutedStyle

	if !status.IsCloned || !status.HasUpstream || status.IsDetached {
		upstreamText = "—"
	} else {
		var parts []string
		if status.AheadCount > 0 {
			parts = append(parts, fmt.Sprintf("↑%d", status.AheadCount))
		}
		if status.BehindCount > 0 {
			parts = append(parts, fmt.Sprintf("↓%d", status.BehindCount))
		}

		if len(parts) > 0 {
			upstreamText = strings.Join(parts, " ")
			if status.AheadCount > 0 && status.BehindCount > 0 {
				upstreamColor = statusWarningStyle
			} else if status.BehindCount > 0 {
				upstreamColor = statusErrorStyle
			} else {
				upstreamColor = statusSuccessStyle
			}
		} else {
			upstreamText = "in sync"
			upstreamColor = statusSuccessStyle
		}
	}

	upstreamTextFormatted := fmt.Sprintf("%-*s", wUpstream, truncate(upstreamText, wUpstream-1))
	if isSelected {
		return selectedTableCellStyle.Render(upstreamTextFormatted)
	}
	return upstreamColor.Render(upstreamTextFormatted)
}

func renderUpdatedCell(status RepoStatus, isSelected bool, wUpdated int) string {
	var updatedText string
	if !status.IsCloned || status.LastUpdated.IsZero() {
		updatedText = "—"
	} else {
		updatedText = relativeTime(status.LastUpdated)
	}

	updatedTextFormatted := fmt.Sprintf("%*s", wUpdated, truncate(updatedText, wUpdated))
	if isSelected {
		return selectedTableCellStyle.Render(updatedTextFormatted)
	}
	return statusMutedStyle.Render(updatedTextFormatted)
}
