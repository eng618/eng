package gpg

import (
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

// SetupGPG is the exported entry point for the GPG setup flow,
// so the setup orchestrator can invoke it via root wiring
// without a sibling cmd import.
func SetupGPG(verbose bool) error {
	return setupGPG(verbose)
}
