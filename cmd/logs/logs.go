package logs

import (
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/runlog"
	"github.com/eng618/eng/internal/ui/theme"
)

// LogsCmd groups session-log inspection commands.
var LogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View session logs from verbose commands",
	Long: `Verbose commands (git bulk syncs, system updates) write full detail logs
to a file while the terminal shows only a clean summary.

Use these subcommands to inspect past runs without re-running them.`,
	Example: `  eng logs              # List recent session logs
  eng logs show         # Show the latest session log
  eng logs show git-sync-all --tail 50
  eng logs clean        # Delete all session logs`,
	Run: func(cmd *cobra.Command, _args []string) {
		listLogs()
	},
}

func init() {
	LogsCmd.AddCommand(ListCmd)
	LogsCmd.AddCommand(ShowCmd)
	LogsCmd.AddCommand(CleanCmd)
}

func listLogs() {
	entries, err := runlog.List()
	if err != nil {
		log.Error("Failed to list session logs: %s", err)
		return
	}
	if len(entries) == 0 {
		log.Info("No session logs yet. Run a verbose command like 'eng git sync-all' first.")
		return
	}

	var boxLines []string
	boxLines = append(boxLines, fmt.Sprintf("Recent sessions (%d, newest first):", len(entries)))
	for _, e := range entries {
		cmd, _ := runlog.ParseName(e.Name)
		boxLines = append(boxLines, fmt.Sprintf("  • %s  %s  %s  (%s)",
			theme.PrimaryText.Render(cmd),
			theme.MutedText.Render(e.ModTime.Format("2006-01-02 15:04:05")),
			theme.MutedText.Render(humanize.Bytes(uint64(e.Size))),
			theme.BoldText.Render(e.Name),
		))
	}
	fmt.Fprintln(log.Out, theme.InfoBox.Render(strings.Join(boxLines, "\n")))
}
