package kill

import (
	"github.com/eng618/eng/internal/execx"
)

// Mockable process-execution seams defaulting to the shared
// internal/execx entry points; tests override per-case as needed.
var (
	execCommand = execx.Command
	lookPath    = execx.LookPath
)
