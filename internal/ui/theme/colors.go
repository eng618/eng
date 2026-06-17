package theme

import "github.com/charmbracelet/lipgloss"

// Color tokens based on gv-tech design system.
var (
	// Primary brand color
	Primary = lipgloss.AdaptiveColor{Light: "#4169e1", Dark: "#6e8dfc"}
	// Secondary accent color
	Secondary = lipgloss.AdaptiveColor{Light: "#86ab69", Dark: "#93c770"}
	// General Background
	Background = lipgloss.AdaptiveColor{Light: "#f5f5f5", Dark: "#171717"}
	// General Foreground text
	Foreground = lipgloss.AdaptiveColor{Light: "#0f1729", Dark: "#ffffff"}
	// Muted background/borders
	Muted = lipgloss.AdaptiveColor{Light: "#ebebeb", Dark: "#0f0f0f"}
	// Muted foreground text (e.g. descriptions, secondary info)
	MutedForeground = lipgloss.AdaptiveColor{Light: "#65758b", Dark: "#b3b3b3"}
	// Destructive/Error color
	Destructive = lipgloss.AdaptiveColor{Light: "#ef4444", Dark: "#7f1d1d"}
)
