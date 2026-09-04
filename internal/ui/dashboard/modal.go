package dashboard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

func (m Model) actionModalSize() (width, height int) {
	width = m.windowWidth - 8
	if width > 76 {
		width = 76
	}
	if width < 40 {
		width = 40
	}
	// Header (2) + progress (2) + rows + tail (1) + hint (1), clamped.
	want := 6 + len(m.actionRows)
	if want < 8 {
		want = 8
	}
	height = m.windowHeight - 6
	if height > want {
		height = want
	}
	if height < 8 {
		height = 8
	}
	return width, height
}

func (m Model) renderActionModal() string {
	modalWidth, _ := m.actionModalSize()
	inner := modalWidth - 4 // modal padding (1,2) on each side
	if inner < 20 {
		inner = 20
	}

	title := m.actionTitle
	if title == "" {
		title = m.actionState
	}
	pct := 0.0
	if m.totalActions > 0 {
		pct = float64(m.completedActions) / float64(m.totalActions)
		if pct > 1.0 {
			pct = 1.0
		}
	}

	var b strings.Builder
	header := fmt.Sprintf("%s  %s", m.spinner.View(), ui.Truncate(title, inner-4))
	b.WriteString(projectNameStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(renderProgressBar(inner-12, pct))
	b.WriteString(" ")
	b.WriteString(progressInfoStyle.Render(fmt.Sprintf("%d/%d", m.completedActions, m.totalActions)))
	b.WriteString("\n")

	// Visible row window around the running row; box height never changes.
	maxRows := m.windowHeight - 12
	if maxRows > len(m.actionRows) {
		maxRows = len(m.actionRows)
	}
	if maxRows < 1 {
		maxRows = 1
	}
	start := 0
	if len(m.actionRows) > maxRows {
		center := m.actionCurrent - maxRows/2
		if center < 0 {
			center = 0
		}
		if center > len(m.actionRows)-maxRows {
			center = len(m.actionRows) - maxRows
		}
		start = center
	}
	end := start + maxRows
	if end > len(m.actionRows) {
		end = len(m.actionRows)
	}

	detailWidth := 12
	nameWidth := inner - detailWidth - 5
	if nameWidth < 12 {
		nameWidth = 12
	}

	for i := start; i < end; i++ {
		b.WriteString("\n")
		b.WriteString(renderActionRow(m.actionRows[i], m.spinner.View(), nameWidth, detailWidth))
	}
	if len(m.actionRows) > maxRows {
		b.WriteString("\n")
		b.WriteString(statusMutedStyle.Render(fmt.Sprintf("▲▼ showing %d/%d", end-start, len(m.actionRows))))
	}

	tail := ui.Truncate(m.actionTail, inner)
	b.WriteString("\n")
	if tail != "" {
		b.WriteString(statusMutedStyle.Render("› " + tail))
	} else {
		b.WriteString(statusMutedStyle.Render("›"))
	}

	return modalStyle.Width(modalWidth).Render(b.String())
}

func renderActionRow(row ActionRow, spinnerView string, nameWidth, detailWidth int) string {
	var glyph, detail string
	var style lipgloss.Style
	switch row.Status {
	case ActionRunning:
		glyph = spinnerView
		detail = "Working…"
		if row.Detail != "" && row.Detail != "Working…" {
			detail = row.Detail
		}
		style = actionRowRunningStyle
	case ActionDone:
		glyph = "✓"
		detail = row.Detail
		if detail == "" {
			detail = "Done"
		}
		style = actionRowSuccessStyle
	case ActionFailed:
		glyph = "✗"
		detail = row.Detail
		if detail == "" {
			detail = "Error"
		}
		style = actionRowErrorStyle
	case ActionSkipped:
		glyph = "−"
		detail = row.Detail
		if detail == "" {
			detail = "Skipped"
		}
		style = actionRowSkippedStyle
	default:
		glyph = "○"
		detail = "Queued"
		style = actionRowPendingStyle
	}
	name := ui.Truncate(row.PrettyName, nameWidth)
	detail = ui.Truncate(detail, detailWidth)
	// Fixed columns: glyph (2) + name (nameWidth) + detail (right-aligned).
	line := fmt.Sprintf("%s %-*s %*s", glyph, nameWidth, name, detailWidth, detail)
	return style.Render(line)
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
