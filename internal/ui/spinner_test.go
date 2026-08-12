package ui

import (
	"bytes"
	"strings"
	"testing"
)

func TestSpinner(t *testing.T) {
	var buf bytes.Buffer
	oldOut := Out
	Out = &buf
	defer func() { Out = oldOut }()

	t.Run("Indeterminate Spinner Non-TTY", func(t *testing.T) {
		buf.Reset()
		s := NewSpinner("testing...")
		s.Start()
		s.UpdateMessage("updated")
		if s.currentMessage != "updated" {
			t.Errorf("expected currentMessage to be 'updated', got %q", s.currentMessage)
		}
		s.Logf("log message %d", 1)
		s.Stop()
	})

	t.Run("Progress Spinner Non-TTY", func(t *testing.T) {
		buf.Reset()
		s := NewProgressSpinner("loading...")
		s.SetProgressBar(0.5, "halfway")
		if s.currentMessage != "halfway" {
			t.Errorf("expected currentMessage to be 'halfway', got %q", s.currentMessage)
		}
		s.Stop()
	})

	t.Run("Progress Spinner TTY Mode", func(t *testing.T) {
		buf.Reset()
		trueVal := true
		forceTTY = &trueVal
		defer func() { forceTTY = nil }()

		s := NewProgressSpinner("step 1")
		s.SetProgressBar(0.2, "step 1")
		out1 := buf.String()
		if !strings.Contains(out1, "step 1") {
			t.Errorf("expected output to contain 'step 1', got %q", out1)
		}

		s.SetProgressBar(0.5, "step 2")
		out2 := buf.String()
		if !strings.Contains(out2, "\r\033[2K") {
			t.Errorf("expected carriage clear sequence '\\r\\033[2K', got %q", out2)
		}

		buf.Reset()
		s.Logf("completed item A")
		logOut := buf.String()
		if !strings.Contains(logOut, "completed item A") {
			t.Errorf("expected log output to contain 'completed item A', got %q", logOut)
		}
		if !strings.Contains(logOut, "step 2") {
			t.Errorf("expected re-rendered progress bar with 'step 2', got %q", logOut)
		}

		buf.Reset()
		s.Stop()
		stopOut := buf.String()
		if stopOut != "\r\033[2K" {
			t.Errorf("expected stop to clear line with '\\r\\033[2K', got %q", stopOut)
		}
	})
}
