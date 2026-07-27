package compose

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/containers"
	"github.com/eng618/eng/internal/ui/theme"
)

var (
	allFlagDown bool
	volumesFlag bool
)

var downCmd = &cobra.Command{
	Use:   "down [stack...]",
	Short: "Spin down one or more Compose stacks",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 && !allFlagDown {
			return fmt.Errorf("specify at least one stack name or use --all (-a)")
		}

		cfg := config.GetContainersConfig()
		mgr := containers.NewManager(cfg.Path)

		targets := args
		if allFlagDown {
			targets = []string{"all"}
		}

		theme.InfoMessage("Spinning down stack(s)...")
		if err := mgr.Down(targets, volumesFlag); err != nil {
			return err
		}

		theme.SuccessMessage("Compose stack(s) stopped successfully.")
		return nil
	},
}

func init() {
	downCmd.Flags().BoolVarP(&allFlagDown, "all", "a", false, "Spin down all discovered stacks")
	downCmd.Flags().BoolVarP(&volumesFlag, "volumes", "V", false, "Remove named volumes declared in compose file")
}
