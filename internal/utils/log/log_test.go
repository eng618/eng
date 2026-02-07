package log

import (
	"bytes"
	"strings"
	"testing"
)

func TestLogFunctions(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	SetWriters(&outBuf, &errBuf)
	defer ResetWriters()

	tests := []struct {
		name     string
		logFunc  func(string, ...any)
		format   string
		args     []any
		waited   string
		isError  bool
	}{
		{
			name:    "Message",
			logFunc: Message,
			format:  "test message %s",
			args:    []any{"foo"},
			waited:  "test message foo",
		},
		{
			name:    "Start",
			logFunc: Start,
			format:  "starting %s",
			args:    []any{"task"},
			waited:  "==> starting task",
		},
		{
			name:    "Success",
			logFunc: Success,
			format:  "finished %s",
			args:    []any{"successfully"},
			waited:  "==> finished successfully",
		},
		{
			name:    "Info",
			logFunc: Info,
			format:  "info %d",
			args:    []any{123},
			waited:  "==> info 123",
		},
		{
			name:    "Debug",
			logFunc: Debug,
			format:  "debug %v",
			args:    []any{true},
			waited:  "==> debug true",
		},
		{
			name:    "Warn",
			logFunc: Warn,
			format:  "warning %s",
			args:    []any{"bit"},
			waited:  "==x warning bit",
		},
		{
			name:    "Error",
			logFunc: Error,
			format:  "error %s",
			args:    []any{"failed"},
			waited:  "==x error failed",
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outBuf.Reset()
			errBuf.Reset()
			tt.logFunc(tt.format, tt.args...)

			var got string
			if tt.isError {
				got = errBuf.String()
			} else {
				got = outBuf.String()
			}

			if !strings.Contains(got, tt.waited) {
				t.Errorf("%s() got = %q, want to contain %q", tt.name, got, tt.waited)
			}
		})
	}
}

func TestCMDWriter(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	SetWriters(&outBuf, &errBuf)
	defer ResetWriters()

	writer := Writer()
	msg := "hello world"
	n, err := writer.Write([]byte(msg))
	if err != nil {
		t.Fatalf("Writer.Write failed: %v", err)
	}
	if n != len(msg) {
		t.Errorf("Writer.Write binary length = %d, want %d", n, len(msg))
	}
	if outBuf.String() != msg {
		t.Errorf("Writer.Write got = %q, want %q", outBuf.String(), msg)
	}
}

func TestCMDErrorWriter(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	SetWriters(&outBuf, &errBuf)
	defer ResetWriters()

	writer := ErrorWriter()
	msg := "error occurred"
	n, err := writer.Write([]byte(msg))
	if err != nil {
		t.Fatalf("ErrorWriter.Write failed: %v", err)
	}
	if n != len(msg) {
		t.Errorf("ErrorWriter.Write binary length = %d, want %d", n, len(msg))
	}
	if errBuf.String() != msg {
		t.Errorf("ErrorWriter.Write got = %q, want %q", errBuf.String(), msg)
	}
}

func TestVerbose(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	SetWriters(&outBuf, &errBuf)
	defer ResetWriters()

	Verbose(false, "should not see this")
	if outBuf.Len() > 0 {
		t.Errorf("Verbose(false) printed: %q", outBuf.String())
	}

	outBuf.Reset()
	Verbose(true, "should see this %d", 42)
	if !strings.Contains(outBuf.String(), "--- should see this 42") {
		t.Errorf("Verbose(true) got = %q, want to contain '--- should see this 42'", outBuf.String())
	}
}

// TestFatal is tricky because it calls os.Exit. We could use a subprocess or just skip it for now.
// Given the scope, skipping is safer unless we want to implement the "Crashing Test" pattern.
