package system

import (
	"github.com/eng618/eng/internal/immich"
)

// ImmichCmd is the Immich stack management command nested under system.
// The tree is built by internal/immich so the top-level `eng immich`
// command shares one implementation with independent flag state.
var ImmichCmd = immich.NewCommand()
