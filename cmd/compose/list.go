package compose

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/containers"
	"github.com/eng618/eng/internal/ui/theme"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List discovered Docker Compose stacks",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.GetContainersConfig()
		mgr := containers.NewManager(cfg.Path)

		stacks, err := mgr.DiscoverStacks()
		if err != nil {
			return err
		}

		if len(stacks) == 0 {
			theme.WarningMessage(fmt.Sprintf("No compose stacks found under %s", cfg.Path))
			return nil
		}

		theme.InfoMessage(fmt.Sprintf("Discovered %d compose stacks under %s", len(stacks), cfg.Path))
		fmt.Println()

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "STACK\tPATH\tSERVICES")
		fmt.Fprintln(w, "-----\t----\t--------")

		for _, s := range stacks {
			svcs := strings.Join(s.Services, ", ")
			if svcs == "" {
				svcs = "-"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", s.Name, s.Path, svcs)
		}
		w.Flush()
		return nil
	},
}
