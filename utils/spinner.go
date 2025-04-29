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

// NewProgressSpinner creates a new progress bar spinner with a message.
func NewProgressSpinner(message string) *Spinner {
	s := spinner.New(spinner.CharSets[36], 120*time.Millisecond)
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

// SetProgressBar sets the progress visually using the progress bar charset and appends the percentage.
func (sp *Spinner) SetProgressBar(progress float64, msg ...string) {
	frames := spinner.CharSets[36]
	idx := int(progress * float64(len(frames)-1))
	if idx < 0 {
		idx = 0
	}
	if idx >= len(frames) {
		idx = len(frames) - 1
	}
	sp.s.UpdateCharSet(frames)
	sp.s.Restart()
	sp.s.Prefix = frames[idx] + " "
	if len(msg) > 0 {
		sp.s.Suffix = fmt.Sprintf(" %s (%.0f%%)", msg[0], progress*100)
	} else {
		sp.s.Suffix = fmt.Sprintf(" (%.0f%%)", progress*100)
	}
}
