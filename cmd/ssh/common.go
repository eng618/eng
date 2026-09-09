package ssh

import (
	"os"

	"github.com/eng618/eng/internal/execx"
	"github.com/eng618/eng/internal/paths"
)

// Mockable system interaction points, migrated from cmd/system/system.go.
// Seams default to the shared internal/execx and internal/paths entry
// points; tests override the package vars per-case as before.
var (
	execCommand = execx.Command
	lookPath    = execx.LookPath
	userHomeDir = paths.UserHomeDir
	stat        = os.Stat
)
