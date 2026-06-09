package ui

import (
	"github.com/pterm/pterm"

	"github.com/eng618/eng/internal/log"
)

// DisableProgress can be set to true in tests to prevent pterm goroutines from spawning and causing race conditions.
var DisableProgress = false

// ProgressSpinner defines the interface for spinner operations.
type ProgressSpinner interface {
	UpdateText(text string)
	Success(text ...interface{})
	Fail(text ...interface{})
	Warning(text ...interface{})
	Info(text ...interface{})
}

// ptermSpinner wraps a pterm.SpinnerPrinter to implement ProgressSpinner.
type ptermSpinner struct {
	p *pterm.SpinnerPrinter
}

func (s *ptermSpinner) UpdateText(text string)      { s.p.UpdateText(text) }
func (s *ptermSpinner) Success(text ...interface{}) { s.p.Success(text...) }
func (s *ptermSpinner) Fail(text ...interface{})    { s.p.Fail(text...) }
func (s *ptermSpinner) Warning(text ...interface{}) { s.p.Warning(text...) }
func (s *ptermSpinner) Info(text ...interface{})    { s.p.Info(text...) }

// dummySpinner is a no-op implementation of ProgressSpinner.
type dummySpinner struct{}

func (s *dummySpinner) UpdateText(text string)      {}
func (s *dummySpinner) Success(text ...interface{}) {}
func (s *dummySpinner) Fail(text ...interface{})    {}
func (s *dummySpinner) Warning(text ...interface{}) {}
func (s *dummySpinner) Info(text ...interface{})    {}

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
func (m *MultiSpinner) AddSpinner(text string) ProgressSpinner {
	if DisableProgress || m.printer == nil {
		return &dummySpinner{}
	}
	spinner, _ := pterm.DefaultSpinner.WithWriter(m.printer.NewWriter()).Start(text)
	return &ptermSpinner{p: spinner}
}

// Stop stops the multi-printer.
func (m *MultiSpinner) Stop() {
	if DisableProgress || m.printer == nil {
		return
	}
	_, _ = m.printer.Stop()
}
