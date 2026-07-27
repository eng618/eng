package ui

import (
	"fmt"
	"io"
	"sync"

	"github.com/charmbracelet/lipgloss"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui/theme"
)

// DisableProgress can be set to true in tests to prevent progress goroutines from spawning.
var DisableProgress = false

// ProgressSpinner defines the interface for spinner operations.
type ProgressSpinner interface {
	UpdateText(text string)
	Success(text ...interface{})
	Fail(text ...interface{})
	Warning(text ...interface{})
	Info(text ...interface{})
}

type charmSpinner struct {
	mu     sync.Mutex
	text   string
	active bool
	out    io.Writer
}

func (s *charmSpinner) UpdateText(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.text = text
}

func (s *charmSpinner) Success(text ...interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active = false
	msg := fmt.Sprint(text...)
	if msg == "" {
		msg = s.text
	}
	banner := theme.SuccessBanner.Render("SUCCESS")
	txt := theme.BaseText.Render(msg)
	fmt.Fprintf(s.out, "%s %s\n", banner, txt)
}

func (s *charmSpinner) Fail(text ...interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active = false
	msg := fmt.Sprint(text...)
	if msg == "" {
		msg = s.text
	}
	banner := theme.ErrorBanner.Render("ERROR")
	txt := theme.ErrorText.Render(msg)
	fmt.Fprintf(s.out, "%s %s\n", banner, txt)
}

func (s *charmSpinner) Warning(text ...interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	msg := fmt.Sprint(text...)
	if msg == "" {
		msg = s.text
	}
	banner := theme.WarningBanner.Render("WARN")
	txt := theme.BaseText.Render(msg)
	fmt.Fprintf(s.out, "%s %s\n", banner, txt)
}

func (s *charmSpinner) Info(text ...interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	msg := fmt.Sprint(text...)
	if msg == "" {
		msg = s.text
	}
	banner := lipgloss.NewStyle().
		Background(theme.Primary).
		Foreground(theme.Background).
		Bold(true).
		Padding(0, 1).
		MarginRight(1).
		Render("INFO")
	txt := theme.BaseText.Render(msg)
	fmt.Fprintf(s.out, "%s %s\n", banner, txt)
}

// dummySpinner is a no-op implementation of ProgressSpinner.
type dummySpinner struct{}

func (s *dummySpinner) UpdateText(text string)      {}
func (s *dummySpinner) Success(text ...interface{}) {}
func (s *dummySpinner) Fail(text ...interface{})    {}
func (s *dummySpinner) Warning(text ...interface{}) {}
func (s *dummySpinner) Info(text ...interface{})    {}

// MultiSpinner manages multiple concurrent progress indicators.
type MultiSpinner struct {
	spinners []*charmSpinner
}

// NewMultiSpinner creates a new multi-spinner manager.
func NewMultiSpinner() (*MultiSpinner, error) {
	return &MultiSpinner{}, nil
}

// AddSpinner adds a new spinner to the display.
func (m *MultiSpinner) AddSpinner(text string) ProgressSpinner {
	if DisableProgress {
		return &dummySpinner{}
	}
	sp := &charmSpinner{
		text:   text,
		active: true,
		out:    log.Writer(),
	}
	m.spinners = append(m.spinners, sp)
	return sp
}

// Stop stops the multi-spinner.
func (m *MultiSpinner) Stop() {
	for _, sp := range m.spinners {
		sp.mu.Lock()
		sp.active = false
		sp.mu.Unlock()
	}
}
