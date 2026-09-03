package ui

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/progress"
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

// MultiProgressBar manages multiple progress bars for concurrent operations.
type MultiProgressBar struct {
	mu         sync.Mutex
	bars       []*ProgressBar
	out        io.Writer
	rendered   bool
	terminal   bool
	overallMsg string
	startTime  time.Time
	doneCount  int
	totalCount int
}

// ProgressBar represents a single progress bar in the multi-progress display.
type ProgressBar struct {
	ID          int
	Label       string
	Detail      string
	Percent     float64
	CurrentPass int
	TotalPasses int
	Done        bool
	Error       error
	StartTime   time.Time
	EndTime     time.Time
	prog        progress.Model
}

// NewMultiProgressBar creates a new multi-progress-bar display.
func NewMultiProgressBar(out io.Writer) *MultiProgressBar {
	if out == nil {
		out = log.Writer()
	}
	return &MultiProgressBar{
		bars:      make([]*ProgressBar, 0),
		out:       out,
		terminal:  IsTerminal(out),
		startTime: time.Now(),
	}
}

// AddBar adds a new progress bar to the display.
func (m *MultiProgressBar) AddBar(label string, totalPasses int) *ProgressBar {
	m.mu.Lock()
	defer m.mu.Unlock()

	p := progress.New(
		progress.WithScaledGradient(
			string(theme.Primary.Dark),
			string(theme.Secondary.Dark),
		),
		progress.WithWidth(40),
	)

	bar := &ProgressBar{
		ID:          len(m.bars),
		Label:       label,
		TotalPasses: totalPasses,
		prog:        p,
		StartTime:   time.Now(),
	}
	m.bars = append(m.bars, bar)
	m.totalCount = len(m.bars)
	return bar
}

// UpdateBar updates a specific progress bar.
func (m *MultiProgressBar) UpdateBar(bar *ProgressBar, percent float64, detail string, currentPass int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	bar.Percent = percent
	bar.Detail = detail
	bar.CurrentPass = currentPass
	if percent >= 1.0 && !bar.Done {
		bar.Done = true
		bar.EndTime = time.Now()
		m.doneCount++
	}
	m.renderLocked()
}

// CompleteBar marks a bar as complete with optional error.
func (m *MultiProgressBar) CompleteBar(bar *ProgressBar, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	bar.Done = true
	bar.Percent = 1.0
	bar.EndTime = time.Now()
	bar.Error = err
	m.doneCount++
	m.renderLocked()
}

// SetOverallMessage sets the overall progress message.
func (m *MultiProgressBar) SetOverallMessage(msg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.overallMsg = msg
	m.renderLocked()
}

// renderLocked renders all progress bars (must hold mutex).
func (m *MultiProgressBar) renderLocked() {
	if m.out == nil || !m.terminal {
		return
	}

	if m.rendered {
		// Move cursor up to first bar line
		fmt.Fprintf(m.out, "\033[%dA", len(m.bars)+2) // +2 for overall message and blank line
	}

	// Print overall message
	if m.overallMsg != "" {
		elapsed := time.Since(m.startTime).Round(time.Second)
		overallLine := fmt.Sprintf("%s  [%d/%d]  %s", m.overallMsg, m.doneCount, m.totalCount, elapsed)
		fmt.Fprintf(m.out, "\r\033[2K%s\n", overallLine)
	} else {
		fmt.Fprintf(m.out, "\r\033[2K\n")
	}

	// Print each bar
	for _, bar := range m.bars {
		m.renderBarLocked(bar)
	}

	m.rendered = true
}

func (m *MultiProgressBar) renderBarLocked(bar *ProgressBar) {
	var statusIcon string
	var labelStyle lipgloss.Style

	if bar.Error != nil {
		statusIcon = theme.ErrorText.Render("✗")
		labelStyle = theme.ErrorText
	} else if bar.Done {
		statusIcon = theme.SuccessText.Render("✓")
		labelStyle = theme.SuccessText
	} else {
		statusIcon = lipgloss.NewStyle().Foreground(theme.Primary).Render("⟳")
		labelStyle = theme.BaseText
	}

	// Truncate label if too long
	maxLabelWidth := 40
	label := bar.Label
	if len(label) > maxLabelWidth {
		label = label[:maxLabelWidth-3] + "..."
	}

	barView := bar.prog.ViewAs(bar.Percent)

	var detail string
	if bar.TotalPasses > 1 && bar.CurrentPass > 0 {
		detail = fmt.Sprintf("  Pass %d/%d", bar.CurrentPass, bar.TotalPasses)
	}
	if bar.Detail != "" {
		detail += "  " + bar.Detail
	}

	line := fmt.Sprintf("  %s  %s  %s%s", statusIcon, labelStyle.Render(label), barView, detail)
	fmt.Fprintf(m.out, "\r\033[2K%s\n", line)
}

// Stop completes progress output and clears active indicators.
func (m *MultiProgressBar) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.terminal || !m.rendered {
		return
	}

	// Clear all bar lines
	fmt.Fprintf(m.out, "\033[%dB", len(m.bars)) // Move down past bars
	for i := 0; i < len(m.bars)+1; i++ {
		fmt.Fprintf(m.out, "\r\033[2K") // Clear line
		if i < len(m.bars) {
			fmt.Fprintf(m.out, "\033[1A") // Move up
		}
	}
	fmt.Fprintf(m.out, "\033[1A")   // Move up to overall line
	fmt.Fprintf(m.out, "\r\033[2K") // Clear overall line
	m.rendered = false
}

// Logf prints a formatted message above the progress bars.
func (m *MultiProgressBar) Logf(format string, a ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.out == nil {
		return
	}

	wasRendered := m.rendered
	if wasRendered && m.terminal {
		// Move down past bars
		fmt.Fprintf(m.out, "\033[%dB", len(m.bars))
	}

	msg := fmt.Sprintf(format, a...)
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	fmt.Fprint(m.out, msg)

	if wasRendered && m.terminal {
		// Move back up
		fmt.Fprintf(m.out, "\033[%dA", len(m.bars)+1)
		m.renderLocked()
	}
}

// Summary returns a summary of all completed bars.
func (m *MultiProgressBar) Summary() (success, failed int, totalBytes int64, elapsed time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, bar := range m.bars {
		if bar.Error != nil {
			failed++
		} else if bar.Done {
			success++
		}
	}
	elapsed = time.Since(m.startTime).Round(time.Second)
	return success, failed, totalBytes, elapsed
}

// ProgressBar returns the progress bar at index i.
func (m *MultiProgressBar) ProgressBar(i int) *ProgressBar {
	m.mu.Lock()
	defer m.mu.Unlock()
	if i >= 0 && i < len(m.bars) {
		return m.bars[i]
	}
	return nil
}
