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

func TestCompletionOutputRequested(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"no args", []string{"eng"}, false},
		{"regular command", []string{"eng", "version"}, false},
		{"completion script", []string{"eng", "completion", "zsh"}, true},
		{"completion bare", []string{"eng", "completion"}, true},
		{"dynamic complete", []string{"eng", "__complete", "ver", ""}, true},
		{
			"completion as flag value still matches",
			[]string{"eng", "config", "--output", "completion"},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := completionOutputRequested(tt.args); got != tt.want {
				t.Errorf("completionOutputRequested(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
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
