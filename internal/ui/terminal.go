package ui

import (
	"io"
	"os"

	"golang.org/x/term"
)

// Out is the writer used for spinner output. Tests may replace it.
var Out io.Writer = os.Stderr

// forceTTY allows overriding TTY detection during unit testing.
var forceTTY *bool

// IsTerminal checks if the writer is a terminal.
func IsTerminal(w io.Writer) bool {
	if forceTTY != nil {
		return *forceTTY
	}
	if f, ok := w.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	if f, ok := w.(interface{ Fd() uintptr }); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}
