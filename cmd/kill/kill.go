package kill

import (
	"github.com/spf13/cobra"
)

// KillCmd groups process termination helpers (by port or PID).
var KillCmd = &cobra.Command{
	Use:   "kill",
	Short: "Find and kill processes by port or PID",
	Long:  `Terminate processes listening on a port or running under a PID, with interactive pickers and dry-run support.`,
	RunE: func(cmd *cobra.Command, _args []string) error {
		return cmd.Help()
	},
}

func init() {
	KillCmd.AddCommand(KillPortCmd)
	KillCmd.AddCommand(KillProcessCmd)
}
