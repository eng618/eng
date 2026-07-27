package ui

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"

	"github.com/eng618/eng/internal/ui/theme"
)

// Out is the writer used for spinner output. Tests may replace it.
var Out io.Writer = os.Stderr

// Spinner manages progress bars and status updates using Lip Gloss and Bubbles.
type Spinner struct {
	mu             sync.Mutex
	baseMessage    string
	currentMessage string
	prog           progress.Model
	isProgress     bool
	currentPercent float64
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

// Start displays initial spinner state.
func (s *Spinner) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if Out != nil {
		fmt.Fprintf(Out, "%s\n", lipgloss.NewStyle().Foreground(theme.Primary).Render("... "+s.currentMessage))
	}
}

// Stop completes progress output.
func (s *Spinner) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
}

// UpdateMessage updates the message displayed next to the progress bar.
func (s *Spinner) UpdateMessage(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.baseMessage = msg
	s.currentMessage = msg
}

// SetProgressBar sets the progress of the bar (0.0 to 1.0).
func (s *Spinner) SetProgressBar(percent float64, msg ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(msg) > 0 {
		s.currentMessage = msg[0]
	}
	s.currentPercent = percent
	if Out != nil && s.isProgress {
		barView := s.prog.ViewAs(percent)
		fmt.Fprintf(Out, "%s %s\n", s.currentMessage, barView)
	}
}

// Logf prints a formatted message above the progress bar.
func (s *Spinner) Logf(format string, a ...interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if Out != nil {
		fmt.Fprintf(Out, format+"\n", a...)
	}
}
