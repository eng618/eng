package immich

import (
	immichsvc "github.com/eng618/eng/internal/immich"
)

// ImmichCmd represents the top-level Immich management command.
var ImmichCmd = immichsvc.NewCommand()
