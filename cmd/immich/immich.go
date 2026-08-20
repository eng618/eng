package immich

import (
	"github.com/spf13/cobra"

	"github.com/eng618/eng/cmd/system"
)

// ImmichCmd represents the top-level Immich management command.
var ImmichCmd = &cobra.Command{
	Use:     "immich",
	Aliases: []string{"photos"},
	Short:   "Manage Immich photo stack, database, backups, and lifecycle",
	Long: `Inspect Immich service health, monitor container states, query database metrics,
execute verified backups, perform disaster recovery, and manage systemd services.`,
	RunE: func(cmd *cobra.Command, _args []string) error {
		return cmd.Help()
	},
}

func init() {
	// Re-export all subcommands from system.ImmichCmd
	for _, sub := range system.ImmichCmd.Commands() {
		// Clone command to allow independent root attachment
		subClone := *sub
		ImmichCmd.AddCommand(&subClone)
	}
}
