package ui

import (
	"fmt"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"

	"github.com/eng618/eng/internal/ui/theme"
)

// Spinner manages progress bars and status updates using Lip Gloss and Bubbles.
type Spinner struct {
	mu             sync.Mutex
	baseMessage    string
	currentMessage string
	prog           progress.Model
	isProgress     bool
	currentPercent float64
	rendered       bool
}

// NewSpinner creates a new spinner with default theme styling.
func NewSpinner(message string) *Spinner {
	p := progress.New(
		progress.WithScaledGradient(
			string(theme.Primary.Dark),
			string(theme.Secondary.Dark),
		),
	)
	return &Spinner{
		baseMessage:    message,
		currentMessage: message,
		prog:           p,
		isProgress:     false,
	}
}

// NewProgressSpinner creates a spinner that displays progress as a bar.
func NewProgressSpinner(message string) *Spinner {
	p := progress.New(
		progress.WithScaledGradient(
			string(theme.Primary.Dark),
			string(theme.Secondary.Dark),
		),
	)
	return &Spinner{
		baseMessage:    message,
		currentMessage: message,
		prog:           p,
		isProgress:     true,
	}
}

func (s *Spinner) renderLocked() {
	if Out == nil {
		return
	}

	var line string
	if s.isProgress {
		barView := s.prog.ViewAs(s.currentPercent)
		line = fmt.Sprintf("%s %s", s.currentMessage, barView)
	} else {
		line = lipgloss.NewStyle().Foreground(theme.Primary).Render("... " + s.currentMessage)
	}

	if IsTerminal(Out) {
		if s.rendered {
			fmt.Fprint(Out, "\r\033[2K")
		}
		fmt.Fprint(Out, line)
		s.rendered = true
	} else {
		fmt.Fprintln(Out, line)
		s.rendered = false
	}
}

func (s *Spinner) clearLineLocked() {
	if Out == nil {
		return
	}
	if IsTerminal(Out) && s.rendered {
		fmt.Fprint(Out, "\r\033[2K")
		s.rendered = false
	}
}

// Start displays initial spinner state.
func (s *Spinner) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.renderLocked()
}

// Stop completes progress output and clears active indicator.
func (s *Spinner) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clearLineLocked()
}

// UpdateMessage updates the message displayed next to the progress bar.
func (s *Spinner) UpdateMessage(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.baseMessage = msg
	s.currentMessage = msg
	if s.rendered || !s.isProgress {
		s.renderLocked()
	}
}

// SetProgressBar sets the progress of the bar (0.0 to 1.0).
func (s *Spinner) SetProgressBar(percent float64, msg ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(msg) > 0 {
		s.currentMessage = msg[0]
	}
	s.currentPercent = percent
	if s.isProgress {
		s.renderLocked()
	}
}

// Logf prints a formatted message above the progress bar.
func (s *Spinner) Logf(format string, a ...interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if Out == nil {
		return
	}
	wasRendered := s.rendered
	s.clearLineLocked()

	msg := fmt.Sprintf(format, a...)
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	fmt.Fprint(Out, msg)

	if wasRendered {
		s.renderLocked()
	}
}
