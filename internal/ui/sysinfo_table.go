package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/eng618/eng/internal/ui/theme"
)

// KeyValueRow is a presentation DTO for one labeled value row. Callers map
// domain types (e.g. sysinfo.SystemInfo) to rows so this package never
// imports domain packages.
type KeyValueRow struct {
	Label string
	Value string
}

// RenderKeyValueTable renders a two-column label/value Lip Gloss table for
// diagnostics output such as system information.
func RenderKeyValueTable(title string, rows []KeyValueRow, termWidth int) string {
	if len(rows) == 0 {
		return theme.MutedText.Render("No system information available.")
	}
	if termWidth <= 0 {
		termWidth = GetTerminalWidth()
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Background).
		Background(theme.Primary).
		Padding(0, 1)

	labelStyle := lipgloss.NewStyle().Bold(true).Padding(0, 1)
	cellStyle := lipgloss.NewStyle().Padding(0, 1)
	borderStyle := lipgloss.NewStyle().Foreground(theme.Primary)

	// Overhead for table borders and inner paddings (2 columns = 3 borders + 4 padding spaces = 7).
	availWidth := termWidth - 7
	if availWidth < 40 {
		availWidth = 40
	}

	colLabel := clamp(availWidth*25/100, 12, 22)
	colValue := availWidth - colLabel
	if colValue < 20 {
		colValue = 20
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(borderStyle).
		Headers("FIELD", "VALUE").
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return headerStyle
			}
			if col == 0 {
				return labelStyle.MaxWidth(colLabel)
			}

			return cellStyle.MaxWidth(colValue)
		})

	for _, r := range rows {
		t.Row(
			Truncate(r.Label, colLabel),
			Truncate(r.Value, colValue),
		)
	}

	if title == "" {
		return t.Render()
	}

	heading := theme.PrimaryText.Bold(true).Render(title)

	return fmt.Sprintf("%s\n%s", heading, t.Render())
}
