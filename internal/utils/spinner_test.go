package utils

import (
	"bytes"
	"testing"
)

func TestSpinner(t *testing.T) {
	var buf bytes.Buffer
	oldOut := Out
	Out = &buf
	defer func() { Out = oldOut }()

	t.Run("Indeterminate Spinner", func(t *testing.T) {
		s := NewSpinner("testing...")
		s.Start()
		s.UpdateMessage("updated")
		if s.currentMessage != "updated" {
			t.Errorf("expected currentMessage to be 'updated', got %q", s.currentMessage)
		}
		s.Logf("log message %d", 1)
		s.Stop()
	})

	t.Run("Progress Spinner", func(t *testing.T) {
		s := NewProgressSpinner("loading...")
		s.SetProgressBar(0.5, "halfway")
		if s.currentMessage != "halfway" {
			t.Errorf("expected currentMessage to be 'halfway', got %q", s.currentMessage)
		}
		s.Stop()
	})
}
