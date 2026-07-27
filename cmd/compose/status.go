package compose

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/containers"
	"github.com/eng618/eng/internal/ui/theme"
)

var (
	jsonFlag      bool
	allFlagStatus bool
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

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "STACK\tCONTAINERS\tSTATUS\tCOMPOSE FILE")
		fmt.Fprintln(w, "-----\t----------\t------\t------------")

		for _, s := range stacks {
			fmt.Fprintf(w, "%s\t%d\t%s\t%s\n", s.Name, s.Containers, s.Status, s.File)
		}
		w.Flush()
		return nil
	},
}

func init() {
	statusCmd.Flags().BoolVar(&jsonFlag, "json", false, "Output status in JSON format")
	statusCmd.Flags().BoolVarP(&allFlagStatus, "all", "a", false, "Include all stacks")
}
