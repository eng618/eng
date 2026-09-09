package auth

import (
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/execx"
)

var (
	execCommand = execx.Command
	lookPath    = execx.LookPath
)

// AuthCmd is the parent for gitlab auth commands.
var AuthCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage GitLab authentication for eng",
}
