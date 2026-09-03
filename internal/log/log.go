/*
Package log is a wrapper around colorizing the log output. It has functions
that allow you to simply write output to the screen for various scenarios.

- Start     ==> Blue/Primary (action starting)
- Success   ✓ Green/Secondary (action succeeded)
- Info      → Cyan/Primary (informational)
- Debug     ··· Magenta (debugging, verbose only in spirit)
- Warn      ⚠ Yellow (to stderr, non-fatal)
- Error     ✗ Red/Destructive (to stderr, fatal-ish)
*/
package log

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/charmbracelet/lipgloss"

	"github.com/eng618/eng/internal/ui/theme"
)

// Out and Err are writers used for normal and error output. Tests may replace them.
var (
	Out io.Writer = os.Stdout
	Err io.Writer = os.Stderr
	mu  sync.RWMutex

	startStyle   = lipgloss.NewStyle().Foreground(theme.Primary)
	successStyle = lipgloss.NewStyle().Foreground(theme.Secondary)
	infoStyle    = lipgloss.NewStyle().Foreground(theme.Primary)
	debugStyle   = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#c084fc", Dark: "#e879f9"})
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#d97706", Dark: "#f59e0b"})
	errorStyle   = lipgloss.NewStyle().Foreground(theme.Destructive)
)

// Message prints a formatted message to the configured Out writer.
func Message(format string, a ...any) {
	mu.Lock()
	defer mu.Unlock()
	s := fmt.Sprintf(format, a...)
	_, _ = fmt.Fprintln(Out, s)
}

// CMDWriter is an io.Writer that writes to the terminal using Info.
type CMDWriter struct{}

// Write implements io.Writer for LogWriter, printing output as an info message.
func (w *CMDWriter) Write(p []byte) (n int, err error) {
	mu.Lock()
	defer mu.Unlock()
	// Write directly to Out to preserve raw output semantics
	return Out.Write(p)
}

// Writer returns a new LogWriter for use as an io.Writer for standard output.
func Writer() *CMDWriter {
	return &CMDWriter{}
}

// CMDErrorWriter is an io.Writer that prints to the terminal using Error.
type CMDErrorWriter struct{}

// Write implements io.Writer for LogErrorWriter, printing output as an error message.
func (w *CMDErrorWriter) Write(p []byte) (n int, err error) {
	mu.Lock()
	defer mu.Unlock()
	return Err.Write(p)
}

// ErrorWriter returns a new LogErrorWriter for use as an io.Writer for error output.
func ErrorWriter() *CMDErrorWriter {
	return &CMDErrorWriter{}
}

// Start prints a message to the terminal in blue, indicating a starting action.
func Start(format string, a ...any) {
	mu.Lock()
	defer mu.Unlock()
	msg := fmt.Sprintf("==> "+format, a...)
	_, _ = fmt.Fprintln(Out, startStyle.Render(msg))
}

// Success prints a message to the terminal in green, indicating a successful action.
func Success(format string, a ...any) {
	mu.Lock()
	defer mu.Unlock()
	msg := fmt.Sprintf("✓ "+format, a...)
	_, _ = fmt.Fprintln(Out, successStyle.Render(msg))
}

// Info prints a message to the terminal in cyan, indicating informational output.
func Info(format string, a ...any) {
	mu.Lock()
	defer mu.Unlock()
	msg := fmt.Sprintf("→ "+format, a...)
	_, _ = fmt.Fprintln(Out, infoStyle.Render(msg))
}

// Debug prints a message to the terminal in magenta, for debugging output.
func Debug(format string, a ...any) {
	mu.Lock()
	defer mu.Unlock()
	msg := fmt.Sprintf("··· "+format, a...)
	_, _ = fmt.Fprintln(Out, debugStyle.Render(msg))
}

// Warn prints a message to stderr in yellow, indicating a warning.
func Warn(format string, a ...any) {
	mu.Lock()
	defer mu.Unlock()
	msg := fmt.Sprintf("⚠ "+format, a...)
	_, _ = fmt.Fprintln(Err, warnStyle.Render(msg))
}

// Error prints a message to stderr in red, indicating an error.
func Error(format string, a ...any) {
	mu.Lock()
	defer mu.Unlock()
	msg := fmt.Sprintf("✗ "+format, a...)
	_, _ = fmt.Fprintln(Err, errorStyle.Render(msg))
}

// Verbose prints a message to the terminal if v is true, prefixed with '---'.
func Verbose(v bool, format string, a ...any) {
	if v {
		Message("--- "+format, a...)
	}
}

// SetWriters allows tests to replace the output writers.
// It also propagates to theme writers so banner messages stay unified.
func SetWriters(out, errOut io.Writer) {
	mu.Lock()
	defer mu.Unlock()
	if out != nil {
		Out = out
		theme.SetWriters(out, nil)
	}
	if errOut != nil {
		Err = errOut
		theme.SetWriters(nil, errOut)
	}
}

// ResetWriters restores writers to their default stdout/stderr.
func ResetWriters() {
	mu.Lock()
	defer mu.Unlock()
	Out = os.Stdout
	Err = os.Stderr
	theme.SetWriters(os.Stdout, os.Stderr)
}
