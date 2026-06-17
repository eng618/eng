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

	inactivePaneStyle = lipgloss.NewStyle().
				Border(paneBorder).
				BorderForeground(theme.Muted).
				Padding(1)

	activePaneStyle = lipgloss.NewStyle().
			Border(paneBorder).
			BorderForeground(theme.Primary).
			Padding(1)

	modalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Primary).
			Padding(1, 4).
			Align(lipgloss.Center)

	overlayStyle = lipgloss.NewStyle().
			Align(lipgloss.Center)

	spinnerStyle = lipgloss.NewStyle().Foreground(theme.Primary)

	projectNameStyle = lipgloss.NewStyle().
				Foreground(theme.Primary).
				Bold(true)

	repoNameStyle = lipgloss.NewStyle().
			Foreground(theme.Foreground).
			Bold(true)

	selectedRepoStyle = lipgloss.NewStyle().
				Background(theme.Muted).
				Foreground(theme.Foreground).
				Bold(true).
				Padding(0, 1)

	statusSuccessStyle = lipgloss.NewStyle().Foreground(theme.Secondary)
	statusErrorStyle   = lipgloss.NewStyle().Foreground(theme.Destructive)
	statusWarningStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#f59e0b", Dark: "#d97706"})
	statusMutedStyle   = lipgloss.NewStyle().Foreground(theme.MutedForeground)

	notificationSuccessStyle = lipgloss.NewStyle().Foreground(theme.Secondary).Bold(true)
	notificationErrorStyle   = lipgloss.NewStyle().Foreground(theme.Destructive).Bold(true)
	notificationWarnStyle    = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#f59e0b", Dark: "#d97706"}).Bold(true)
)
