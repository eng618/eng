package setup

import (
	"context"
	"os"
	"os/exec"

	"github.com/eng618/eng/internal/sysinfo"
)

// Mockable system interaction points, migrated from cmd/system/system.go.
// Prefer internal/paths and internal/execx for new code.
var (
	execCommand  = exec.Command
	lookPath     = exec.LookPath
	userHomeDir  = os.UserHomeDir
	stat         = os.Stat
	detectDistro = sysinfo.Detect
)

// Cross-domain setup steps, wired by cmd/root.go (the only package allowed
// to aggregate sibling commands). Defaults fail loudly so missing wiring
// surfaces in tests rather than silently skipping.
//
// TODO: move GPG/IDE implementations to internal/ services so setup
// calls services directly instead of via root-wired function variables.
var (
	// GPGSetup runs the GPG setup flow (wired to cmd/gpg.SetupGPG).
	GPGSetup = func(_ bool) error {
		return errUnwired("GPGSetup (cmd/gpg)")
	}
	// IDEUpdate runs the Antigravity IDE installer (wired to cmd/update.RunIdeUpdate).
	IDEUpdate = func(_ context.Context, _ string, _, _ bool) error {
		return errUnwired("IDEUpdate (cmd/update)")
	}
)

func errUnwired(name string) error {
	return &unwiredError{name: name}
}

type unwiredError struct {
	name string
}

func (e *unwiredError) Error() string {
	return "setup step unwired: " + e.name + " (cmd/root.go must wire it)"
}
