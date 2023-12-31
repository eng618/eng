/*
package log is a wrapper around colorizing the log output. It has functions
that allow you to simply write output to the screen for various scenarios.

- Successful ==> Green

- Info       ==> Cyan

- Warn       ==> Yellow

- Error      ==> Red

- Fatal      ==> Red
*/
package log

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

// Message simply prints a message to the terminal.
func Message(format string, a ...any) {
	s := fmt.Sprintf(format, a...)
	fmt.Println(s)
}

// Start prints a message to the terminal in the color green.
func Start(format string, a ...any) {
	color.Blue("==> "+format, a...)
}

// Success prints a message to the terminal in the color green.
func Success(format string, a ...any) {
	color.Green("==> "+format, a...)
}

// Info prints a message to the terminal in the color cyan.
func Info(format string, a ...any) {
	color.Cyan("==> "+format, a...)
}

// Warn prints a message to the terminal in the color yellow.
func Warn(format string, a ...any) {
	color.Yellow("==x "+format, a...)
}

// Error prints a message to the terminal in the color red.
func Error(format string, a ...any) {
	color.Red("==x "+format, a...)
}

// Fatal prints a message to the terminal in the color red, then exits with error.
func Fatal(format string, a ...any) {
	Error(fmt.Sprintf(format, a...))
	os.Exit(1)
}

// Verbose is a log wrapper.
func Verbose(v bool, format string, a ...any) {
	if v {
		Message(format, a...)
	}
}
