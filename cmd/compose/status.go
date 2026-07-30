package compose

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/containers"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

var (
	jsonFlag      bool
	allFlagStatus bool
	detailsFlag   bool
)

var statusCmd = &cobra.Command{
	Use:     "status [stack...]",
	Aliases: []string{"ps"},
	Short:   "Show status of Compose stacks and services",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.GetContainersConfig()
		mgr := containers.NewManager(cfg.Path)

		targets := args
		if len(args) == 0 || allFlagStatus {
			targets = []string{"all"}
		}

		if detailsFlag {
			details, err := mgr.ContainerDetails(targets)
			if err != nil {
				return err
			}

			if jsonFlag {
				out, err := json.MarshalIndent(details, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(out))
				return nil
			}

			theme.InfoMessage("Compose Swarms Status (Detailed):")
			fmt.Println()

			for stackName, containersList := range details {
				fmt.Println(ui.RenderContainerTable(stackName, containersList))
				fmt.Println()
			}
			return nil
		}

		stacks, err := mgr.Status(targets)
		if err != nil {
			return err
		}

		if jsonFlag {
			out, err := json.MarshalIndent(stacks, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(out))
			return nil
		}

		theme.InfoMessage("Compose Swarms Status:")
		fmt.Println()
		fmt.Println(ui.RenderStackTable(stacks))
		return nil
	},
}

func init() {
	statusCmd.Flags().BoolVar(&jsonFlag, "json", false, "Output status in JSON format")
	statusCmd.Flags().BoolVarP(&allFlagStatus, "all", "a", false, "Include all stacks")
	statusCmd.Flags().BoolVarP(&detailsFlag, "details", "d", false, "Show detailed container inspection")
}
