package utils

import (
	"fmt"
	"os"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

// Spinner wraps the mpb.Progress container and a single bar.
// It's designed to manage a single progress bar and allow logging
// messages above the bar without disrupting it.
type Spinner struct {
	p           *mpb.Progress
	bar         *mpb.Bar
	msgCh       chan string
	baseMessage string
}

// NewSpinner creates a new spinner with a default indeterminate style.
func NewSpinner(message string) *Spinner {
	p := mpb.New(mpb.WithOutput(os.Stderr))
	msgCh := make(chan string, 1)
	msgCh <- message

	bar := p.New(0, // Total is 0 for an indeterminate spinner
		mpb.SpinnerStyle(),
		mpb.PrependDecorators(
			decor.Any(func(_ decor.Statistics) string {
				select {
				case msg := <-msgCh:
					return msg
				default:
					return ""
				}
			}),
		),
		mpb.AppendDecorators(
			decor.Elapsed(decor.ET_STYLE_GO, decor.WC{W: 4}),
		),
	)

	return &Spinner{
		p:           p,
		bar:         bar,
		msgCh:       msgCh,
		baseMessage: message,
	}
}

// NewProgressSpinner creates a spinner that displays progress as a bar.
func NewProgressSpinner(message string) *Spinner {
	p := mpb.New(mpb.WithOutput(os.Stderr))
	msgCh := make(chan string, 1)
	msgCh <- message

	bar := p.New(100, // Total is 100 for percentage-based progress
		mpb.BarStyle().Lbound("[").Filler("=").Tip(">").Padding("-").Rbound("]"),
		mpb.PrependDecorators(
			decor.Any(func(_ decor.Statistics) string {
				select {
				case msg := <-msgCh:
					return msg
				default:
					return ""
				}
			}),
			decor.Percentage(decor.WCSyncSpace),
		),
		mpb.AppendDecorators(
			decor.Elapsed(decor.ET_STYLE_GO, decor.WC{W: 4}),
		),
	)

	return &Spinner{
		p:           p,
		bar:         bar,
		msgCh:       msgCh,
		baseMessage: message,
	}
}

// Start does nothing in this implementation, as the bar is visible on creation.
func (s *Spinner) Start() {
	// The bar is displayed automatically by the mpb.Progress container.
}

// Stop marks the bar as completed and waits for the progress container to finish.
func (s *Spinner) Stop() {
	if s.bar != nil {
		s.bar.SetTotal(0, true) // Mark as complete
	}
	if s.p != nil {
		s.p.Wait() // Wait for the container to finish rendering
	}
}

// UpdateMessage updates the message displayed next to the spinner/bar.
func (s *Spinner) UpdateMessage(msg string) {
	s.baseMessage = msg
	s.msgCh <- msg
}

// SetProgressBar sets the progress of the bar. Progress should be from 0.0 to 1.0.
func (s *Spinner) SetProgressBar(progress float64, msg ...string) {
	if s.bar != nil {
		currentMessage := s.baseMessage
		if len(msg) > 0 {
			currentMessage = msg[0]
		}
		s.UpdateMessage(currentMessage)
		s.bar.SetCurrent(int64(progress * 100))
	}
}

// Logf prints a formatted message above the progress bar.
func (s *Spinner) Logf(format string, a ...interface{}) {
	if s.p != nil {
		//nolint:errcheck
		fmt.Fprintf(s.p, format, a...)
	}
}
