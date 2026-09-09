// Package execx centralizes process execution behind mockable entry points.
//
// Historically each cmd package declared its own
// `var execCommand = exec.Command` / `lookPath = exec.LookPath`, scattering
// ~100 direct os/exec call sites. New code should use this package; existing
// call sites migrate incrementally.
package execx

import (
	"context"
	"os/exec"
)

// Mockable function variables. Tests may stub these.
var (
	Command        = exec.Command
	CommandContext = exec.CommandContext
	LookPath       = exec.LookPath
)

// Cmd aliases exec.Cmd so callers can drop the os/exec import entirely.
type Cmd = exec.Cmd

// ExitError aliases exec.ExitError for exit-code inspection without os/exec.
type ExitError = exec.ExitError

// Runner is a minimal interface for running external processes.
// DefaultRunner delegates to os/exec; tests provide fakes.
type Runner interface {
	CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd
	LookPath(file string) (string, error)
}

// DefaultRunner is the production Runner.
type DefaultRunner struct{}

// CommandContext creates an exec.Cmd bound to ctx.
func (DefaultRunner) CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	return CommandContext(ctx, name, args...)
}

// LookPath resolves a binary on PATH.
func (DefaultRunner) LookPath(file string) (string, error) {
	return LookPath(file)
}

// Default is the production Runner instance.
var Default Runner = DefaultRunner{}
