package immich

import (
	immichsvc "github.com/eng618/eng/internal/immich"
)

// ImmichCmd represents the top-level Immich management command.
// Built from the same factory as the nested `eng system immich` tree,
// with independent flag state per registration.
var ImmichCmd = immichsvc.NewCommand()
