package ui

import (
	"github.com/pterm/pterm"

	"github.com/eng618/eng/internal/log"
)

// MultiSpinner manages multiple concurrent spinners.
type MultiSpinner struct {
	printer *pterm.MultiPrinter
}

// NewMultiSpinner creates a new multi-spinner manager.
func NewMultiSpinner() (*MultiSpinner, error) {
	pterm.SetDefaultOutput(log.Writer())
	p, err := pterm.DefaultMultiPrinter.WithWriter(log.Writer()).Start()
	if err != nil {
		return nil, err
	}
	return &MultiSpinner{printer: p}, nil
}

// AddSpinner adds a new spinner to the multi-printer display.
func (m *MultiSpinner) AddSpinner(text string) *pterm.SpinnerPrinter {
	spinner, _ := pterm.DefaultSpinner.WithWriter(m.printer.NewWriter()).Start(text)
	return spinner
}

// Stop stops the multi-printer.
func (m *MultiSpinner) Stop() {
	_, _ = m.printer.Stop()
}
