package ui

import (
	"github.com/charmbracelet/lipgloss"
	ltable "github.com/charmbracelet/lipgloss/table"

	"github.com/eng618/eng/internal/ui/theme"
)

// TableOpts configures RenderTable. Headers and Rows are plain strings;
// styling, borders, and truncation are applied centrally.
type TableOpts struct {
	Headers []string
	Rows    [][]string
	// StyleFunc optionally styles a cell (row, col zero-indexed, data rows only).
	StyleFunc func(row, col int) lipgloss.Style
	// MaxColWidth truncates cells wider than this (runes). 0 = no truncation.
	MaxColWidth int
	// EmptyMsg is returned when Rows is empty.
	EmptyMsg string
}

// RenderTable renders tabula data with the shared CLI theme: bold padded
// header, rounded Primary border, optional per-cell StyleFunc.
func RenderTable(opts TableOpts) string {
	if len(opts.Rows) == 0 {
		if opts.EmptyMsg != "" {
			return theme.MutedText.Render(opts.EmptyMsg)
		}
		return theme.MutedText.Render("No results found.")
	}
	rows := opts.Rows
	if opts.MaxColWidth > 0 {
		rows = make([][]string, len(opts.Rows))
		for i, r := range opts.Rows {
			nr := make([]string, len(r))
			for j, c := range r {
				nr[j] = Truncate(c, opts.MaxColWidth)
			}
			rows[i] = nr
		}
	}
	t := ltable.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(theme.Primary)).
		Headers(opts.Headers...).
		Rows(rows...)
	headerStyle := lipgloss.NewStyle().Bold(true).Padding(0, 1)
	t.StyleFunc(func(row, col int) lipgloss.Style {
		if row == 0 {
			return headerStyle
		}
		if opts.StyleFunc != nil {
			return opts.StyleFunc(row-1, col)
		}
		return lipgloss.NewStyle().Padding(0, 1)
	})
	return t.Render()
}
