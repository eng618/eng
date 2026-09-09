package gpg

import (
	"os"

	"github.com/eng618/eng/internal/execx"
	"github.com/eng618/eng/internal/paths"
	"github.com/eng618/eng/internal/sysinfo"
)

// Mockable system interaction points, migrated from cmd/system/system.go.
// Seams default to the shared internal/execx and internal/paths entry
// points; tests override the package vars per-case as before.
var (
	execCommand  = execx.Command
	lookPath     = execx.LookPath
	userHomeDir  = paths.UserHomeDir
	stat         = os.Stat
	detectDistro = sysinfo.Detect
)

// SetupGPG is the exported entry point for the GPG setup flow,
// so the setup orchestrator can invoke it via root wiring
// without a sibling cmd import.
func SetupGPG(verbose bool) error {
	return setupGPG(verbose)
}
