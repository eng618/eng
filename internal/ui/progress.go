package ui

import (
	"github.com/pterm/pterm"

	"github.com/eng618/eng/internal/log"
)

// DisableProgress can be set to true in tests to prevent pterm goroutines from spawning and causing race conditions.
var DisableProgress = false

// MultiSpinner manages multiple concurrent spinners.
type MultiSpinner struct {
	printer *pterm.MultiPrinter
}

// NewMultiSpinner creates a new multi-spinner manager.
func NewMultiSpinner() (*MultiSpinner, error) {
	if DisableProgress {
		return &MultiSpinner{printer: nil}, nil
	}

	pterm.SetDefaultOutput(log.Writer())
	p, err := pterm.DefaultMultiPrinter.WithWriter(log.Writer()).Start()
	if err != nil {
		return nil, err
	}
	return &MultiSpinner{printer: p}, nil
}

// AddSpinner adds a new spinner to the multi-printer display.
func (m *MultiSpinner) AddSpinner(text string) *pterm.SpinnerPrinter {
	if DisableProgress {
		return &pterm.SpinnerPrinter{}
	}
	spinner, _ := pterm.DefaultSpinner.WithWriter(m.printer.NewWriter()).Start(text)
	return spinner
}

// Stop stops the multi-printer.
func (m *MultiSpinner) Stop() {
	if DisableProgress || m.printer == nil {
		return
	}
	_, _ = m.printer.Stop()
}
