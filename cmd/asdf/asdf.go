package asdf

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var (
	execCommand = exec.Command
	lookPath    = exec.LookPath
	userHomeDir = os.UserHomeDir
)

// AsdfCmd is the root command for asdf management.
var AsdfCmd = &cobra.Command{
	Use:   "asdf",
	Short: "Manage asdf version manager plugins and installs",
	Long:  `Command suite for managing asdf version manager plugins and cleaning up outdated tool installs.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default to running cleanup if no subcommand is provided
		return runCleanup(cmd, args)
	},
}

func init() {
	AsdfCmd.AddCommand(CleanupCmd)

	// Bind flags to AsdfCmd as well so running `eng asdf [flags]` works identically to `eng asdf cleanup [flags]`
	bindCleanupFlags(AsdfCmd)
	bindCleanupFlags(CleanupCmd)
}
