package ts

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
)

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

func TestTailscaleCmd(t *testing.T) {
	cmd := TailscaleCmd

	output := captureStdout(func() {
		cmd.Run(cmd, []string{})
	})

	if !strings.Contains(output, "tailscale called") {
		t.Errorf("expected output to contain 'tailscale called', got %q", output)
	}
}

func TestUpCmd_Success(t *testing.T) {
	original := execCommand
	defer func() { execCommand = original }()

	var invokedCmd string
	var invokedArgs []string

	execCommand = func(name string, arg ...string) *exec.Cmd {
		invokedCmd = name
		invokedArgs = arg
		return exec.Command("true")
	}

	cmd := UpCmd
	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if invokedCmd != "sudo" {
		t.Errorf("expected invoked command 'sudo', got %q", invokedCmd)
	}

	expectedArgs := []string{"tailscale", "up"}
	if len(invokedArgs) != len(expectedArgs) {
		t.Fatalf("expected arguments length %d, got %d", len(expectedArgs), len(invokedArgs))
	}
	for i, arg := range invokedArgs {
		if arg != expectedArgs[i] {
			t.Errorf("expected arg[%d] to be %q, got %q", i, expectedArgs[i], arg)
		}
	}
}

func TestUpCmd_Failure(t *testing.T) {
	original := execCommand
	defer func() { execCommand = original }()

	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("false")
	}

	cmd := UpCmd
	err := cmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("expected command execution to fail, but it succeeded")
	}
}

func TestDownCmd_Success(t *testing.T) {
	original := execCommand
	defer func() { execCommand = original }()

	var invokedCmd string
	var invokedArgs []string

	execCommand = func(name string, arg ...string) *exec.Cmd {
		invokedCmd = name
		invokedArgs = arg
		return exec.Command("true")
	}

	cmd := DownCmd
	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if invokedCmd != "sudo" {
		t.Errorf("expected invoked command 'sudo', got %q", invokedCmd)
	}

	expectedArgs := []string{"tailscale", "down"}
	if len(invokedArgs) != len(expectedArgs) {
		t.Fatalf("expected arguments length %d, got %d", len(expectedArgs), len(invokedArgs))
	}
	for i, arg := range invokedArgs {
		if arg != expectedArgs[i] {
			t.Errorf("expected arg[%d] to be %q, got %q", i, expectedArgs[i], arg)
		}
	}
}

func TestDownCmd_Failure(t *testing.T) {
	original := execCommand
	defer func() { execCommand = original }()

	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("false")
	}

	cmd := DownCmd
	err := cmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("expected command execution to fail, but it succeeded")
	}
}

// ============================================================================
// E - Examples (Executable Documentation)
// ============================================================================

func ExampleTailscaleCmd() {
	fmt.Println("Command Use:", TailscaleCmd.Use)
	fmt.Println("Aliases:", TailscaleCmd.Aliases)
	// Output:
	// Command Use: tailscale
	// Aliases: [ts]
}

// ============================================================================
// B - Benchmarks (Performance Profiling)
// ============================================================================

func BenchmarkTailscaleCmd_Init(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TailscaleCmd.Commands()
	}
}
