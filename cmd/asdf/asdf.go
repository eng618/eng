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
	Long:  `Command suite for managing asdf version manager plugins, checking project requirements, updating root .tool-versions, and pruning outdated tool installs.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default to running prune if no subcommand is provided
		return runPrune(cmd, args)
	},
}

func init() {
	AsdfCmd.AddCommand(PruneCmd)
	AsdfCmd.AddCommand(CheckCmd)
	AsdfCmd.AddCommand(UpdateLatestCmd)
	AsdfCmd.AddCommand(StatusCmd)

	// Bind flags to AsdfCmd as well so running `eng asdf [flags]` works identically to `eng asdf prune [flags]`
	bindPruneFlags(AsdfCmd)
	bindPruneFlags(PruneCmd)
}
