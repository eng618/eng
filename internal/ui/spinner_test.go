package ui

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"
)

// lockedBuffer is a goroutine-safe bytes.Buffer for asserting animated output.
type lockedBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (l *lockedBuffer) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.Write(p)
}

func (l *lockedBuffer) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.String()
}

func (l *lockedBuffer) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.Len()
}

func (l *lockedBuffer) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.b.Reset()
}

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

func TestSpinnerAnimationTTY(t *testing.T) {
	buf := &lockedBuffer{}
	oldOut := Out
	Out = buf
	defer func() { Out = oldOut }()

	trueVal := true
	forceTTY = &trueVal
	defer func() { forceTTY = nil }()

	oldTick, oldGrace := spinnerTickInterval, spinnerElapsedGrace
	spinnerTickInterval = 10 * time.Millisecond
	spinnerElapsedGrace = 0
	defer func() { spinnerTickInterval, spinnerElapsedGrace = oldTick, oldGrace }()

	t.Run("Animates frames and stops cleanly", func(t *testing.T) {
		buf.Reset()
		s := NewSpinner("working...")
		s.Start()
		time.Sleep(100 * time.Millisecond)
		animated := buf.String()
		foundFrame := false
		for _, f := range spinnerFrames {
			if strings.Contains(animated, f) {
				foundFrame = true
				break
			}
		}
		if !foundFrame {
			t.Errorf("expected animated braille frame in output, got %q", animated)
		}
		if !strings.Contains(animated, "working...") {
			t.Errorf("expected message in animated output, got %q", animated)
		}

		s.Stop()
		afterStop := buf.Len()
		time.Sleep(50 * time.Millisecond)
		if buf.Len() != afterStop {
			t.Errorf("expected no output after Stop, grew from %d to %d bytes", afterStop, buf.Len())
		}
	})

	t.Run("Stop Success Fail are idempotent", func(t *testing.T) {
		buf.Reset()
		s := NewSpinner("idempotent...")
		s.Start()
		time.Sleep(30 * time.Millisecond)
		s.Stop()
		s.Stop()
		s.Success("done")
		s.Fail("failed")
		if !strings.Contains(buf.String(), "done") || !strings.Contains(buf.String(), "failed") {
			t.Errorf("expected completion traces, got %q", buf.String())
		}
	})

	t.Run("Elapsed timer appears", func(t *testing.T) {
		buf.Reset()
		s := NewSpinner("slow...")
		s.Start()
		time.Sleep(50 * time.Millisecond)
		s.Stop()
		if !strings.Contains(buf.String(), "s)") {
			t.Errorf("expected elapsed timer in output, got %q", buf.String())
		}
	})
}

func TestSpinnerNoAnimationWhenDisabled(t *testing.T) {
	buf := &lockedBuffer{}
	oldOut, oldDisable := Out, DisableProgress
	Out = buf
	DisableProgress = true
	defer func() { Out, DisableProgress = oldOut, oldDisable }()

	trueVal := true
	forceTTY = &trueVal
	defer func() { forceTTY = nil }()

	s := NewSpinner("quiet...")
	s.Start()
	afterStart := buf.Len()
	time.Sleep(50 * time.Millisecond)
	s.Stop()
	// Only the single static initial render is allowed; no ticker output.
	if buf.Len() != afterStart && buf.Len() != afterStart+len("\r\033[2K") {
		t.Errorf("expected no animated output when disabled, got %q", buf.String())
	}
}
