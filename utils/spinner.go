package utils

import (
	"time"

	"github.com/briandowns/spinner"
)

type Spinner struct {
	s *spinner.Spinner
}

// NewSpinner creates a new spinner with the given message.
func NewSpinner(message string) *Spinner {
	s := spinner.New(spinner.CharSets[14], 120*time.Millisecond)
	s.Suffix = " " + message
	return &Spinner{s: s}
}

func (sp *Spinner) Start() {
	sp.s.Start()
}

func (sp *Spinner) Stop() {
	sp.s.Stop()
}

func (sp *Spinner) UpdateMessage(msg string) {
	sp.s.Suffix = " " + msg
}
