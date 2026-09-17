package compose

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/containers"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List discovered Docker Compose stacks",
	RunE: func(cmd *cobra.Command, args []string) error {
		header := theme.PrimaryText.Bold(true).Render("🐳 Discovered Docker Compose Stacks")
		if !ui.DisableProgress {
			fmt.Fprintln(log.Out, header)
		}

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

		var rows [][]string
		for _, s := range stacks {
			svcs := strings.Join(s.Services, ", ")
			if svcs == "" {
				svcs = "-"
			}
			rows = append(rows, []string{s.Name, s.Path, svcs})
		}
		subheader := fmt.Sprintf("Discovered %s compose stack(s) under %s:",
			theme.PrimaryText.Bold(true).Render(fmt.Sprintf("%d", len(stacks))),
			theme.BoldText.Render(cfg.Path),
		)

		if !ui.DisableProgress {
			fmt.Fprintln(log.Out, theme.InfoBox.Render(subheader+"\n"+ui.RenderTable(ui.TableOpts{
				Headers: []string{"STACK", "PATH", "SERVICES"},
				Rows:    rows,
			})))
		} else {
			for _, r := range rows {
				log.Info("%s | %s | %s", r[0], r[1], r[2])
			}
		}

		theme.SuccessMessage(fmt.Sprintf("Listed %d Docker Compose stack(s)", len(stacks)))
		return nil
	},
}
