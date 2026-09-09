package ssh

import (
	"os"
	"os/exec"
)

// Mockable system interaction points, migrated from cmd/system/system.go.
// Prefer internal/paths and internal/execx for new code.
var (
	execCommand = exec.Command
	lookPath    = exec.LookPath
	userHomeDir = os.UserHomeDir
	stat        = os.Stat
)
