package compose

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/containers"
	"github.com/eng618/eng/internal/ui/theme"
)

var (
	envFlag    string
	allFlagUp  bool
	detachFlag bool
	buildFlag  bool
)

var upCmd = &cobra.Command{
	Use:   "up [stack...]",
	Short: "Spin up one or more Compose stacks",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 && !allFlagUp {
			return fmt.Errorf("specify at least one stack name or use --all (-a)")
		}

		cfg := config.GetContainersConfig()
		mgr := containers.NewManager(cfg.Path)

		targets := args
		if allFlagUp {
			targets = []string{"all"}
		}

		theme.InfoMessage(fmt.Sprintf("Spinning up stack(s) (env: %s)...", envFlag))
		if err := mgr.Up(targets, envFlag, detachFlag, buildFlag); err != nil {
			return err
		}

		theme.SuccessMessage("Compose stack(s) started successfully.")
		return nil
	},
}

func init() {
	upCmd.Flags().StringVarP(&envFlag, "env", "e", "prod", "Target environment (e.g. dev, staging, prod)")
	upCmd.Flags().BoolVarP(&allFlagUp, "all", "a", false, "Spin up all discovered stacks")
	upCmd.Flags().BoolVarP(&detachFlag, "detach", "d", true, "Run containers in background")
	upCmd.Flags().BoolVar(&buildFlag, "build", false, "Build images before starting containers")
}
