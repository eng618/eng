package theme

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"
)

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// stripANSI removes ANSI escape codes from styled terminal outputs.
func stripANSI(str string) string {
	return ansiRegex.ReplaceAllString(str, "")
}

// captureStderr redirects os.Stderr to a buffer during function execution.
func captureStderr(f func()) string {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	f()

	_ = w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

// captureStdout redirects os.Stdout to a buffer during function execution.
func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

// ============================================================================
// T - Tests (Unit and Table-driven Tests)
// ============================================================================

func TestNewActionableError(t *testing.T) {
	tests := []struct {
		name             string
		err              error
		suggestion       string
		expectNil        bool
		expectMessage    string
		expectSuggestion string
	}{
		{
			name:       "nil error returns nil",
			err:        nil,
			suggestion: "try again",
			expectNil:  true,
		},
		{
			name:             "valid error with suggestion",
			err:              errors.New("connection failed"),
			suggestion:       "check your internet connection",
			expectNil:        false,
			expectMessage:    "connection failed",
			expectSuggestion: "check your internet connection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actErr := NewActionableError(tt.err, tt.suggestion)
			if tt.expectNil {
				if actErr != nil {
					t.Fatalf("expected nil error, got: %v", actErr)
				}
				return
			}

			if actErr == nil {
				t.Fatal("expected actionable error, got nil")
			}

			if actErr.Error() != tt.expectMessage {
				t.Errorf("expected message %q, got %q", tt.expectMessage, actErr.Error())
			}

			var actionable ActionableError
			if !errors.As(actErr, &actionable) {
				t.Fatal("expected error to implement ActionableError interface")
			}

			if actionable.Suggestion() != tt.expectSuggestion {
				t.Errorf("expected suggestion %q, got %q", tt.expectSuggestion, actionable.Suggestion())
			}

			// Verify unwrap
			unwrapped := errors.Unwrap(actErr)
			if unwrapped == nil || unwrapped.Error() != tt.err.Error() {
				t.Errorf("expected unwrapped error to be equivalent to original error")
			}
		})
	}
}

func TestHandleError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectContains []string
	}{
		{
			name:           "nil error prints nothing",
			err:            nil,
			expectContains: []string{},
		},
		{
			name:           "standard error contains ERROR and message",
			err:            errors.New("db error"),
			expectContains: []string{"ERROR", "db error"},
		},
		{
			name:           "actionable error contains suggestion",
			err:            NewActionableError(errors.New("command missing"), "install python"),
			expectContains: []string{"ERROR", "command missing", "Suggestion", "install python"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStderr(func() {
				HandleError(tt.err)
			})
			cleanOut := stripANSI(output)

			if len(tt.expectContains) == 0 {
				if cleanOut != "" {
					t.Errorf("expected empty output, got: %q", cleanOut)
				}
				return
			}

			for _, sub := range tt.expectContains {
				if !strings.Contains(cleanOut, sub) {
					t.Errorf("expected output to contain %q, clean output: %q", sub, cleanOut)
				}
			}
		})
	}
}

func TestMessages(t *testing.T) {
	tests := []struct {
		name           string
		msgFunc        func(string)
		msg            string
		expectContains []string
	}{
		{
			name:           "SuccessMessage",
			msgFunc:        SuccessMessage,
			msg:            "successfully completed",
			expectContains: []string{"SUCCESS", "successfully completed"},
		},
		{
			name:           "InfoMessage",
			msgFunc:        InfoMessage,
			msg:            "some info",
			expectContains: []string{"INFO", "some info"},
		},
		{
			name:           "ErrorMessage",
			msgFunc:        ErrorMessage,
			msg:            "an error occurred",
			expectContains: []string{"ERROR", "an error occurred"},
		},
		{
			name:           "WarningMessage",
			msgFunc:        WarningMessage,
			msg:            "a warning",
			expectContains: []string{"WARN", "a warning"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStdout(func() {
				tt.msgFunc(tt.msg)
			})
			cleanOut := stripANSI(output)

			for _, sub := range tt.expectContains {
				if !strings.Contains(cleanOut, sub) {
					t.Errorf("expected output to contain %q, clean output: %q", sub, cleanOut)
				}
			}
		})
	}
}

func TestEngTheme(t *testing.T) {
	theme := EngTheme()
	if theme == nil {
		t.Fatal("expected EngTheme to return a non-nil Theme")
	}
	// Basic field checking
	if theme.Focused.Title.GetForeground() == nil {
		t.Error("expected Focused.Title style to have a custom foreground color")
	}
	if theme.Focused.SelectSelector.GetForeground() == nil {
		t.Error("expected Focused.SelectSelector style to have a custom foreground color")
	}
	if theme.Blurred.Title.GetForeground() == nil {
		t.Error("expected Blurred.Title style to have a custom foreground color")
	}
}

// ============================================================================
// E - Examples (Executable Documentation)
// ============================================================================

func ExampleNewActionableError() {
	err := NewActionableError(errors.New("auth failed"), "run: eng gitlab auth set")
	var actErr ActionableError
	if errors.As(err, &actErr) {
		fmt.Println(actErr.Error())
		fmt.Println(actErr.Suggestion())
	}
	// Output:
	// auth failed
	// run: eng gitlab auth set
}

// ============================================================================
// B - Benchmarks (Performance Profiling)
// ============================================================================

func BenchmarkNewActionableError(b *testing.B) {
	err := errors.New("underlying error")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewActionableError(err, "do something else")
	}
}

func BenchmarkSuccessMessage(b *testing.B) {
	// Discard stdout during benchmark to avoid console spam
	old := os.Stdout
	os.Stdout = os.NewFile(0, os.DevNull)
	defer func() { os.Stdout = old }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SuccessMessage("benchmark message")
	}
}

func BenchmarkHandleError(b *testing.B) {
	// Discard stderr during benchmark to avoid console spam
	old := os.Stderr
	os.Stderr = os.NewFile(0, os.DevNull)
	defer func() { os.Stderr = old }()

	err := NewActionableError(errors.New("benchmark error"), "check configurations")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HandleError(err)
	}
}

func BenchmarkEngTheme(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EngTheme()
	}
}
