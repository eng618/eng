package compose

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/containers"
)

var ComposeCmd = &cobra.Command{
	Use:     "compose",
	Aliases: []string{"swarm", "stack"},
	Short:   "Manage Docker Compose swarms and services",
	Long:    `Audit, inspect, start, stop, pull, and monitor Docker Compose service stacks.`,
	Example: `  eng compose list
  eng compose status --all
  eng compose up media -e dev`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// completeStackNames offers discovered compose stack names for shell completion.
// It never errors: on discovery failure it returns no completions.
func completeStackNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	cfg := config.GetContainersConfig()
	stacks, err := containers.NewManager(cfg.Path).DiscoverStacks()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var names []string
	for _, s := range stacks {
		if strings.HasPrefix(s.Name, toComplete) {
			names = append(names, s.Name)
		}
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func init() {
	ComposeCmd.AddCommand(listCmd)
	ComposeCmd.AddCommand(upCmd)
	ComposeCmd.AddCommand(downCmd)
	ComposeCmd.AddCommand(pullCmd)
	ComposeCmd.AddCommand(statusCmd)
	ComposeCmd.AddCommand(logsCmd)
	ComposeCmd.AddCommand(cleanCmd)

	for _, c := range []*cobra.Command{upCmd, downCmd, pullCmd, statusCmd, logsCmd} {
		c.ValidArgsFunction = completeStackNames
	}
}
