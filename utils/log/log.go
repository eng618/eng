/*
Package log is a wrapper around colorizing the log output. It has functions
that allow you to simply write output to the screen for various scenarios.

- Start     ==> Blue
- Success   ==> Green
- Info      ==> Cyan
- Debug     ==> Magenta
- Warn      ==> Yellow
- Error     ==> Red
- Fatal     ==> Red (and exits the program)
*/
package log

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
)

// Out and Err are writers used for normal and error output. Tests may replace them.
var Out io.Writer = os.Stdout
var Err io.Writer = os.Stderr

// Message prints a formatted message to the configured Out writer.
func Message(format string, a ...any) {
	s := fmt.Sprintf(format, a...)
	_, _ = fmt.Fprintln(Out, s)
}

// Writer returns an io.Writer that prints to the terminal using Message.
type logWriter struct{}

// Write implements io.Writer for logWriter, printing output as a message.
func (w *logWriter) Write(p []byte) (n int, err error) {
	// Write directly to Out to preserve raw output semantics
	return Out.Write(p)
}

// Writer returns a new logWriter for use as an io.Writer for standard output.
func Writer() *logWriter {
	return &logWriter{}
}

// ErrorWriter returns an io.Writer that prints to the terminal using Error.
type logErrorWriter struct{}

// Write implements io.Writer for logErrorWriter, printing output as an error message.
func (w *logErrorWriter) Write(p []byte) (n int, err error) {
	return Err.Write(p)
}

// ErrorWriter returns a new logErrorWriter for use as an io.Writer for error output.
func ErrorWriter() *logErrorWriter {
	return &logErrorWriter{}
}

// Start prints a message to the terminal in blue, indicating a starting action.
func Start(format string, a ...any) {
	_, _ = color.New(color.FgBlue).Fprintf(Out, "==> "+format+"\n", a...)
}

// Success prints a message to the terminal in green, indicating a successful action.
func Success(format string, a ...any) {
	_, _ = color.New(color.FgGreen).Fprintf(Out, "==> "+format+"\n", a...)
}

// Info prints a message to the terminal in cyan, indicating informational output.
func Info(format string, a ...any) {
	_, _ = color.New(color.FgCyan).Fprintf(Out, "==> "+format+"\n", a...)
}

// Debug prints a message to the terminal in magenta, for debugging output.
func Debug(format string, a ...any) {
	_, _ = color.New(color.FgMagenta).Fprintf(Out, "==> "+format+"\n", a...)
}

// Warn prints a message to the terminal in yellow, indicating a warning.
func Warn(format string, a ...any) {
	_, _ = color.New(color.FgYellow).Fprintf(Out, "==x "+format+"\n", a...)
}

// Error prints a message to the terminal in red, indicating an error.
func Error(format string, a ...any) {
	_, _ = color.New(color.FgRed).Fprintf(Err, "==x "+format+"\n", a...)
}

// Fatal prints a message to the terminal in red, then exits the program with an error code.
func Fatal(format string, a ...any) {
	Error(format, a...)
	os.Exit(1)
}

// Verbose prints a message to the terminal if v is true, prefixed with '---'.
func Verbose(v bool, format string, a ...any) {
	if v {
		Message("--- "+format, a...)
	}
}

// SetWriters allows tests to replace the output writers.
func SetWriters(out, errOut io.Writer) {
	if out != nil {
		Out = out
	}
	if errOut != nil {
		Err = errOut
	}
}

// ResetWriters restores writers to their default stdout/stderr.
func ResetWriters() {
	Out = os.Stdout
	Err = os.Stderr
}
