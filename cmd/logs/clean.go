package logs

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/runlog"
)

// CleanCmd deletes session log files.
var CleanCmd = &cobra.Command{
	Use:     "clean",
	Short:   "Delete session logs",
	Long:    `Deletes captured session log files. Reports how many were removed.`,
	Example: `  eng logs clean`,
	RunE: func(cmd *cobra.Command, _args []string) error {
		entries, err := runlog.List()
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			log.Info("No session logs to clean.")
			return nil
		}
		removed := 0
		for _, e := range entries {
			if err := os.Remove(e.Path); err != nil {
				log.Warn("Could not remove %s: %v", e.Name, err)
				continue
			}
			removed++
		}
		log.Success("Removed %d session log(s).", removed)
		return nil
	},
}
