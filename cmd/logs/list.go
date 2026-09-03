package logs

import (
	"github.com/spf13/cobra"
)

// ListCmd lists recent session log files.
var ListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List recent session logs",
	Long:    `Lists captured session logs, newest first, with command, time, and size.`,
	Example: `  eng logs list`,
	Run: func(cmd *cobra.Command, _args []string) {
		listLogs()
	},
}
