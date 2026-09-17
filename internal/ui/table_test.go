package ui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/require"
)

func TestRenderTable_HeadersAndRows(t *testing.T) {
	out := RenderTable(TableOpts{
		Headers: []string{"NAME", "STATUS"},
		Rows:    [][]string{{"web", "running"}, {"db", "stopped"}},
	})
	require.Contains(t, out, "NAME")
	require.Contains(t, out, "web")
	require.Contains(t, out, "db")
	// Rounded border characters from shared theme.
	require.Contains(t, out, "╭")
	require.Contains(t, out, "╰")
}

func TestRenderTable_Empty(t *testing.T) {
	out := RenderTable(TableOpts{Headers: []string{"A"}, EmptyMsg: "No containers found."})
	require.Contains(t, out, "No containers found.")

	def := RenderTable(TableOpts{Headers: []string{"A"}})
	require.Contains(t, def, "No results found.")
}

func TestRenderTable_TruncatesWideCells(t *testing.T) {
	out := RenderTable(TableOpts{
		Headers:     []string{"LABEL"},
		Rows:        [][]string{{"abcdefghijklmnopqrstuvwxyz"}},
		MaxColWidth: 10,
	})
	require.Contains(t, out, "abcdefg...")
	require.NotContains(t, out, "abcdefghijklmnopqrstuvwxyz")
}

func TestRenderTable_StyleFuncApplies(t *testing.T) {
	called := false
	out := RenderTable(TableOpts{
		Headers: []string{"S"},
		Rows:    [][]string{{"x"}},
		StyleFunc: func(row, col int) lipgloss.Style {
			called = true
			return lipgloss.NewStyle().Padding(0, 1)
		},
	})
	require.True(t, called)
	require.Contains(t, out, "x")
}
