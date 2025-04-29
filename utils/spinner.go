package utils

import (
	"fmt"
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

func (sp *Spinner) SetProgress(progress float64, msg ...string) {
	if len(msg) > 0 {
		sp.s.Suffix = fmt.Sprintf(" %s", msg[0])
	} else {
		sp.s.Suffix = fmt.Sprintf(" (%.0f%%)", progress*100)
	}
}
