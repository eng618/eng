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
	"os"

	"github.com/fatih/color"
)

// Message simply prints a formatted message to the terminal.
func Message(format string, a ...any) {
	s := fmt.Sprintf(format, a...)
	fmt.Println(s)
}

// Writer returns an io.Writer that prints to the terminal using Message.
type logWriter struct{}

// Write implements io.Writer for logWriter, printing output as a message.
func (w *logWriter) Write(p []byte) (n int, err error) {
	Message("%s", string(p))
	return len(p), nil
}

// Writer returns a new logWriter for use as an io.Writer for standard output.
func Writer() *logWriter {
	return &logWriter{}
}

// ErrorWriter returns an io.Writer that prints to the terminal using Error.
type logErrorWriter struct{}

// Write implements io.Writer for logErrorWriter, printing output as an error message.
func (w *logErrorWriter) Write(p []byte) (n int, err error) {
	Error("%s", string(p))
	return len(p), nil
}

// ErrorWriter returns a new logErrorWriter for use as an io.Writer for error output.
func ErrorWriter() *logErrorWriter {
	return &logErrorWriter{}
}

// Start prints a message to the terminal in blue, indicating a starting action.
func Start(format string, a ...any) {
	color.Blue("==> "+format, a...)
}

// Success prints a message to the terminal in green, indicating a successful action.
func Success(format string, a ...any) {
	color.Green("==> "+format, a...)
}

// Info prints a message to the terminal in cyan, indicating informational output.
func Info(format string, a ...any) {
	color.Cyan("==> "+format, a...)
}

// Debug prints a message to the terminal in magenta, for debugging output.
func Debug(format string, a ...any) {
	color.Magenta("==> "+format, a...)
}

// Warn prints a message to the terminal in yellow, indicating a warning.
func Warn(format string, a ...any) {
	color.Yellow("==x "+format, a...)
}

// Error prints a message to the terminal in red, indicating an error.
func Error(format string, a ...any) {
	color.Red("==x "+format, a...)
}

// Fatal prints a message to the terminal in red, then exits the program with an error code.
func Fatal(format string, a ...any) {
	Error(fmt.Sprintf(format, a...))
	os.Exit(1)
}

// Verbose prints a message to the terminal if v is true, prefixed with '---'.
func Verbose(v bool, format string, a ...any) {
	if v {
		Message("--- "+format, a...)
	}
}
