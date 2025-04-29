package utils

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
)

type Spinner struct {
	s             *spinner.Spinner
	charSet       []string // Store the character set for progress calculation
	isProgressBar bool     // Flag to indicate if this spinner uses a progress charset
	baseMessage   string   // Store the core message
}

// NewSpinner creates a new standard spinner with the given message.
func NewSpinner(message string) *Spinner {
	s := spinner.New(spinner.CharSets[14], 120*time.Millisecond) // Default spinner
	s.Suffix = " " + message
	return &Spinner{
		s:             s,
		charSet:       spinner.CharSets[14], // Store the actual charset
		isProgressBar: false,
		baseMessage:   message, // Store the base message
	}
}

// NewProgressSpinner creates a new spinner styled as a progress bar.
// Use SetProgressBar() on the returned spinner to update its progress.
func NewProgressSpinner(message string) *Spinner {
	// Using CharSet 36 as the default progress bar style
	progressCharSet := spinner.CharSets[36]
	s := spinner.New(progressCharSet, 120*time.Millisecond)
	s.Suffix = " " + message // Initial suffix includes the message
	// Initialize prefix to the first frame (0%)
	s.Prefix = progressCharSet[0] + " "
	return &Spinner{
		s:             s,
		charSet:       progressCharSet, // Store the progress charset
		isProgressBar: true,
		baseMessage:   message, // Store the base message
	}
}

func (sp *Spinner) Start() {
	sp.s.Start()
}

func (sp *Spinner) Stop() {
	sp.s.Stop()
}

// UpdateMessage updates the base message of the spinner.
// For non-progress spinners, it updates the suffix immediately.
// For progress spinners, the change will be reflected on the next SetProgressBar call
// if no specific message is provided to SetProgressBar.
func (sp *Spinner) UpdateMessage(msg string) {
	sp.baseMessage = msg // Update the stored base message
	// Only update the immediate suffix if it's not a progress bar,
	// as SetProgressBar handles suffix formatting for progress bars.
	if !sp.isProgressBar {
		sp.s.Suffix = " " + msg
	}
	// If it IS a progress bar, the next call to SetProgressBar
	// will pick up the new baseMessage if no specific message is passed to it.
}

// SetProgressBar updates the visual progress of the spinner.
// It assumes the spinner was created using NewProgressSpinner or configured
// with a character set suitable for displaying progress steps (like CharSets[36]).
// The progress value should be between 0.0 and 1.0.
// If msg is provided, it overrides the base message for this update.
// If msg is not provided, the stored base message is used.
func (sp *Spinner) SetProgressBar(progress float64, msg ...string) {
	// Determine the message to display for this update
	currentMessage := sp.baseMessage // Default to stored base message
	if len(msg) > 0 {
		currentMessage = msg[0] // Override with provided message for this update
	}

	// Handle non-progress bar spinners (just update suffix with percentage)
	if !sp.isProgressBar || len(sp.charSet) == 0 {
		// Format suffix based on whether there's a message
		if currentMessage != "" {
			sp.s.Suffix = fmt.Sprintf(" %s (%.0f%%)", currentMessage, progress*100)
		} else {
			sp.s.Suffix = fmt.Sprintf(" (%.0f%%)", progress*100)
		}
		// Don't try to set prefix if not a progress bar type
		// Note: We don't return here anymore, allowing suffix update even if isProgressBar is true but charSet is empty
		if !sp.isProgressBar {
			return
		}
	}

	// --- Progress Bar Specific Logic ---

	// Calculate frame index based on progress
	frames := sp.charSet
	idx := int(progress * float64(len(frames)-1))

	// Clamp index
	if idx < 0 {
		idx = 0
	}
	if idx >= len(frames) {
		idx = len(frames) - 1
	}

	// Update Prefix to show the progress character
	sp.s.Prefix = frames[idx] + " "

	// Update Suffix to show the message (current or base) and percentage
	if currentMessage != "" {
		sp.s.Suffix = fmt.Sprintf(" %s (%.0f%%)", currentMessage, progress*100)
	} else {
		// If no message provided and base message was empty, just show percentage
		sp.s.Suffix = fmt.Sprintf(" (%.0f%%)", progress*100)
	}
}
