package compose

import (
	"github.com/spf13/cobra"
)

var ComposeCmd = &cobra.Command{
	Use:     "compose",
	Aliases: []string{"swarm", "stack"},
	Short:   "Manage Docker Compose swarms and services",
	Long:    `Audit, inspect, start, stop, pull, and monitor Docker Compose service stacks.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	ComposeCmd.AddCommand(listCmd)
	ComposeCmd.AddCommand(upCmd)
	ComposeCmd.AddCommand(downCmd)
	ComposeCmd.AddCommand(pullCmd)
	ComposeCmd.AddCommand(statusCmd)
	ComposeCmd.AddCommand(logsCmd)
	ComposeCmd.AddCommand(cleanCmd)
}
