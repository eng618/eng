package dashboard

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/eng618/eng/internal/repo"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

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

func (m Model) renderRepoTable(p Project, innerRightWidth, H_repos int) string {
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

		repoName, err := repo.EffectivePath(r.URL, r.Path)
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
		repoText = fmt.Sprintf("▸ %-*s", wRepo-2, ui.Truncate(repoName, wRepo-2))
		return selectedTableCellStyle.Render(repoText)
	}
	repoText = fmt.Sprintf("  %-*s", wRepo-2, ui.Truncate(repoName, wRepo-2))
	return repoNameStyle.Render(repoText)
}

func renderBranchCell(status RepoStatus, isSelected bool, wBranch int) string {
	var branchText string
	branchColor := statusMutedStyle

	if !status.IsCloned {
		branchText = "—"
	} else if status.IsDetached {
		branchText = ui.Truncate(status.Branch, wBranch-1)
		branchColor = statusWarningStyle
	} else {
		branchText = ui.Truncate(status.Branch, wBranch-1)
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

	statusTextFormatted := fmt.Sprintf("%-*s", wStatus, ui.Truncate(statusText, wStatus-1))
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

	upstreamTextFormatted := fmt.Sprintf("%-*s", wUpstream, ui.Truncate(upstreamText, wUpstream-1))
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

	updatedTextFormatted := fmt.Sprintf("%*s", wUpdated, ui.Truncate(updatedText, wUpdated))
	if isSelected {
		return selectedTableCellStyle.Render(updatedTextFormatted)
	}
	return statusMutedStyle.Render(updatedTextFormatted)
}
