package ui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"

	"github.com/eng618/eng/internal/ui/theme"
)

// spinnerFrames are the braille animation frames cycled by Start on terminals.
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// spinnerTickInterval and spinnerElapsedGrace are vars (not consts) so tests
// can speed up the animation without real-time waits.
var (
	spinnerTickInterval = 100 * time.Millisecond
	spinnerElapsedGrace = 2 * time.Second
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

	wg        sync.WaitGroup
	stopCh    chan struct{}
	startedAt time.Time
	frame     int
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
	} else if IsTerminal(Out) {
		frame := spinnerFrames[s.frame%len(spinnerFrames)]
		msg := s.currentMessage
		if !s.startedAt.IsZero() {
			if elapsed := time.Since(s.startedAt); elapsed >= spinnerElapsedGrace {
				msg = fmt.Sprintf("%s (%ds)", msg, int(elapsed.Seconds()))
			}
		}
		line = lipgloss.NewStyle().Foreground(theme.Primary).Render(frame + " " + msg)
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

// Start displays initial spinner state. On terminals it also launches an
// animation ticker (braille frames + elapsed timer) until Stop/Success/Fail.
// Non-terminal output and DisableProgress keep the legacy single-print.
func (s *Spinner) Start() {
	s.mu.Lock()
	s.startedAt = time.Now()
	if Out != nil && !s.isProgress && !DisableProgress && IsTerminal(Out) && s.stopCh == nil {
		s.stopCh = make(chan struct{})
		s.wg.Add(1)
		go s.tickLoop(s.stopCh)
	}
	s.renderLocked()
	s.mu.Unlock()
}

// tickLoop re-renders the spinner frame until stopCh closes.
func (s *Spinner) tickLoop(stopCh chan struct{}) {
	defer s.wg.Done()
	ticker := time.NewTicker(spinnerTickInterval)
	defer ticker.Stop()
	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			s.mu.Lock()
			select {
			case <-stopCh:
				s.mu.Unlock()
				return
			default:
			}
			s.frame++
			s.renderLocked()
			s.mu.Unlock()
		}
	}
}

// stopAnimation halts the ticker goroutine. Call without holding s.mu.
func (s *Spinner) stopAnimation() {
	s.mu.Lock()
	ch := s.stopCh
	s.stopCh = nil
	s.mu.Unlock()
	if ch != nil {
		close(ch)
		s.wg.Wait()
	}
}

// Stop completes progress output and clears active indicator.
func (s *Spinner) Stop() {
	s.stopAnimation()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clearLineLocked()
}

// Success clears the spinner and leaves a ✓ completion trace.
func (s *Spinner) Success(msg string) {
	s.stopAnimation()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clearLineLocked()
	if Out == nil {
		return
	}
	if msg == "" {
		msg = s.currentMessage
	}
	line := lipgloss.NewStyle().Foreground(theme.Secondary).Render("✓ " + msg)
	if IsTerminal(Out) {
		fmt.Fprintln(Out, line)
		s.rendered = false
	} else {
		fmt.Fprintln(Out, line)
		s.rendered = false
	}
}

// Fail clears the spinner and leaves a ✗ completion trace.
func (s *Spinner) Fail(msg string) {
	s.stopAnimation()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clearLineLocked()
	if Out == nil {
		return
	}
	if msg == "" {
		msg = s.currentMessage
	}
	line := lipgloss.NewStyle().Foreground(theme.Destructive).Render("✗ " + msg)
	fmt.Fprintln(Out, line)
	s.rendered = false
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
