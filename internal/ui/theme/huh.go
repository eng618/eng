package theme

import (
	"github.com/charmbracelet/huh"
)

// EngTheme returns a custom theme for huh forms based on the eng CLI brand colors.
// It uses the gv-tech token palette for light and dark modes.
func EngTheme() *huh.Theme {
	t := huh.ThemeBase()

	// Customize the base theme fields using our global tokens.
	f := &t.Focused
	f.Base = f.Base.Foreground(Foreground)
	f.Title = f.Title.Foreground(Primary).Bold(true)
	f.NoteTitle = f.NoteTitle.Foreground(Primary).Bold(true)
	f.Description = f.Description.Foreground(MutedForeground)
	f.Directory = f.Directory.Foreground(Primary)
	f.File = f.File.Foreground(Foreground)
	f.ErrorIndicator = f.ErrorIndicator.Foreground(Destructive)
	f.ErrorMessage = f.ErrorMessage.Foreground(Destructive)
	f.SelectSelector = f.SelectSelector.Foreground(Primary)
	f.NextIndicator = f.NextIndicator.Foreground(Primary)
	f.PrevIndicator = f.PrevIndicator.Foreground(Primary)
	f.Option = f.Option.Foreground(Foreground)
	f.MultiSelectSelector = f.MultiSelectSelector.Foreground(Primary)
	f.SelectedOption = f.SelectedOption.Foreground(Secondary)
	f.SelectedPrefix = f.SelectedPrefix.Foreground(Secondary)
	f.UnselectedPrefix = f.UnselectedPrefix.Foreground(MutedForeground)
	f.UnselectedOption = f.UnselectedOption.Foreground(Foreground)
	f.FocusedButton = f.FocusedButton.Foreground(Background).Background(Primary).Bold(true)
	f.BlurredButton = f.BlurredButton.Foreground(Foreground).Background(Muted)
	f.TextInput.Cursor = f.TextInput.Cursor.Foreground(Primary)
	f.TextInput.Placeholder = f.TextInput.Placeholder.Foreground(MutedForeground)
	f.TextInput.Prompt = f.TextInput.Prompt.Foreground(Primary)

	b := &t.Blurred
	b.Base = b.Base.Foreground(Foreground)
	b.Title = b.Title.Foreground(MutedForeground)
	b.NoteTitle = b.NoteTitle.Foreground(MutedForeground)
	b.Description = b.Description.Foreground(MutedForeground)

	return t
}
