package dashboard

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/eng618/eng/internal/ui/theme"
)

var (
	// Ensure we align with the global GV-Tech design system.
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	listTitleStyle = lipgloss.NewStyle().
		Foreground(theme.Foreground).
		Background(theme.Primary).
		Padding(0, 1)

	paneBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "╰",
		BottomRight: "╯",
	}

	leftPaneStyle = lipgloss.NewStyle().
		Border(paneBorder).
		BorderForeground(theme.Muted).
		Padding(1)

	rightPaneStyle = lipgloss.NewStyle().
		Border(paneBorder).
		BorderForeground(theme.Muted).
		Padding(1)

	projectNameStyle = lipgloss.NewStyle().
		Foreground(theme.Primary).
		Bold(true)

	repoNameStyle = lipgloss.NewStyle().
		Foreground(theme.Foreground).
		Bold(true)

	statusSuccessStyle = lipgloss.NewStyle().Foreground(theme.Secondary)
	statusErrorStyle   = lipgloss.NewStyle().Foreground(theme.Destructive)
	statusWarningStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#f59e0b", Dark: "#d97706"})
	statusMutedStyle   = lipgloss.NewStyle().Foreground(theme.MutedForeground)
)
