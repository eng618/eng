package ui

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// EngTheme returns a custom theme for huh forms based on the eng CLI brand colors.
// It detects the terminal's color profile (light vs dark) and applies the appropriate palette.
func EngTheme() *huh.Theme {
	t := huh.ThemeBase()

	// Create adaptive colors using lipgloss.AdaptiveColor to support both light and dark backgrounds.
	primary := lipgloss.AdaptiveColor{Light: "#4169e1", Dark: "#6e8cfc"}
	secondary := lipgloss.AdaptiveColor{Light: "#86aa68", Dark: "#92c76f"}
	bg := lipgloss.AdaptiveColor{Light: "#f4f4f4", Dark: "#161616"}
	fg := lipgloss.AdaptiveColor{Light: "#0e1629", Dark: "#ffffff"}
	muted := lipgloss.AdaptiveColor{Light: "#4a5568", Dark: "#a0aec0"} // Slightly muted foreground for help text
	errorColor := lipgloss.AdaptiveColor{Light: "#e53e3e", Dark: "#fc8181"}

	// Customize the base theme fields.
	f := &t.Focused
	f.Base = f.Base.Foreground(fg)
	f.Title = f.Title.Foreground(primary).Bold(true)
	f.NoteTitle = f.NoteTitle.Foreground(primary).Bold(true)
	f.Description = f.Description.Foreground(muted)
	f.Directory = f.Directory.Foreground(primary)
	f.File = f.File.Foreground(fg)
	f.ErrorIndicator = f.ErrorIndicator.Foreground(errorColor)
	f.ErrorMessage = f.ErrorMessage.Foreground(errorColor)
	f.SelectSelector = f.SelectSelector.Foreground(primary)
	f.NextIndicator = f.NextIndicator.Foreground(primary)
	f.PrevIndicator = f.PrevIndicator.Foreground(primary)
	f.Option = f.Option.Foreground(fg)
	f.MultiSelectSelector = f.MultiSelectSelector.Foreground(primary)
	f.SelectedOption = f.SelectedOption.Foreground(secondary)
	f.SelectedPrefix = f.SelectedPrefix.Foreground(secondary)
	f.UnselectedPrefix = f.UnselectedPrefix.Foreground(muted)
	f.UnselectedOption = f.UnselectedOption.Foreground(fg)
	f.FocusedButton = f.FocusedButton.Foreground(bg).Background(primary).Bold(true)
	f.BlurredButton = f.BlurredButton.Foreground(fg).Background(bg)
	f.TextInput.Cursor = f.TextInput.Cursor.Foreground(primary)
	f.TextInput.Placeholder = f.TextInput.Placeholder.Foreground(muted)
	f.TextInput.Prompt = f.TextInput.Prompt.Foreground(primary)

	b := &t.Blurred
	b.Base = b.Base.Foreground(fg)
	b.Title = b.Title.Foreground(muted)
	b.NoteTitle = b.NoteTitle.Foreground(muted)
	b.Description = b.Description.Foreground(muted)

	return t
}
