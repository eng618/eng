package auth

import (
	"os/exec"

	"github.com/spf13/cobra"
)

var execCommand = exec.Command
var lookPath = exec.LookPath

// AuthCmd is the parent for gitlab auth commands.
var AuthCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage GitLab authentication for eng",
}
