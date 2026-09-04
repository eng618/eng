package cmd

import (
	"errors"
	"strings"
	"testing"

	"github.com/eng618/eng/internal/log"
)

func captureLog(t *testing.T) *strings.Builder {
	t.Helper()
	var buf strings.Builder
	log.SetWriters(&buf, &buf)
	t.Cleanup(log.ResetWriters)
	return &buf
}

func TestRunOnboardingPromptDeclined(t *testing.T) {
	buf := captureLog(t)
	editCalled := false

	runOnboardingPrompt(
		func(string, bool) (bool, error) { return false, nil },
		func() error { editCalled = true; return nil },
	)

	if editCalled {
		t.Error("expected editor not to run when declined")
	}
	if out := buf.String(); !strings.Contains(out, "eng config edit --interactive") {
		t.Errorf("expected skip hint, got %q", out)
	}
}

func TestRunOnboardingPromptAborted(t *testing.T) {
	buf := captureLog(t)
	editCalled := false

	runOnboardingPrompt(
		func(string, bool) (bool, error) { return false, errors.New("aborted") },
		func() error { editCalled = true; return nil },
	)

	if editCalled {
		t.Error("expected editor not to run when prompt aborts")
	}
	if out := buf.String(); !strings.Contains(out, "Skipped") {
		t.Errorf("expected skip message, got %q", out)
	}
}

func TestRunOnboardingPromptAccepted(t *testing.T) {
	buf := captureLog(t)
	editCalled := false

	runOnboardingPrompt(
		func(string, bool) (bool, error) { return true, nil },
		func() error { editCalled = true; return nil },
	)

	if !editCalled {
		t.Error("expected editor to run when confirmed")
	}
	if out := buf.String(); strings.Contains(out, "Skipped") {
		t.Errorf("expected no skip message, got %q", out)
	}
}

func TestRunOnboardingPromptEditFails(t *testing.T) {
	buf := captureLog(t)

	runOnboardingPrompt(
		func(string, bool) (bool, error) { return true, nil },
		func() error { return errors.New("boom") },
	)

	if out := buf.String(); !strings.Contains(out, "retry") {
		t.Errorf("expected retry hint, got %q", out)
	}
}
