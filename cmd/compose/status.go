package compose

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/containers"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

var (
	jsonFlag      bool
	allFlagStatus bool
	detailsFlag   bool
	pagerFlag     bool
)

var statusCmd = &cobra.Command{
	Use:     "status [stack...]",
	Aliases: []string{"ps"},
	Short:   "Show status of Compose stacks and services",
	RunE: func(cmd *cobra.Command, args []string) error {
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary).
			MarginBottom(1)
		if !ui.DisableProgress && !jsonFlag {
			fmt.Fprintln(log.Out, headerStyle.Render("📊 Docker Compose Swarms Status"))
		}
		cfg := config.GetContainersConfig()
		mgr := containers.NewManager(cfg.Path)

		targets := args
		if len(args) == 0 || allFlagStatus {
			targets = []string{"all"}
		}

		termWidth := ui.GetTerminalWidth()

		if detailsFlag {
			details, err := mgr.ContainerDetails(targets)
			if err != nil {
				return err
			}

			if jsonFlag {
				if details == nil {
					details = map[string][]containers.ContainerDetail{}
				}
				out, err := json.MarshalIndent(details, "", "  ")
				if err != nil {
					return err
				}
				fmt.Fprintln(log.Out, string(out))
				return nil
			}

			var rendered string
			for stackName, containersList := range details {
				rendered += ui.RenderContainerTable(stackName, toContainerRows(containersList), termWidth) + "\n\n"
			}

			if pagerFlag {
				return ui.RunContainersPager(rendered)
			}

			theme.InfoMessage("Compose Swarms Status (Detailed):")
			fmt.Fprintln(log.Out)
			fmt.Fprint(log.Out, rendered)
			return nil
		}

		stacks, err := mgr.Status(targets)
		if err != nil {
			return err
		}

		if jsonFlag {
			if stacks == nil {
				stacks = []containers.Stack{}
			}
			out, err := json.MarshalIndent(stacks, "", "  ")
			if err != nil {
				return err
			}
			fmt.Fprintln(log.Out, string(out))
			return nil
		}

		if pagerFlag {
			rendered := ui.RenderStackTable(toStackRows(stacks), termWidth)
			return ui.RunContainersPager(rendered)
		}

		theme.InfoMessage("Compose Swarms Status:")
		fmt.Fprintln(log.Out)
		fmt.Fprintln(log.Out, ui.RenderStackTable(toStackRows(stacks), termWidth))
		return nil
	},
}

// toStackRows maps domain stacks to presentation rows for ui.RenderStackTable.
func toStackRows(stacks []containers.Stack) []ui.StackRow {
	rows := make([]ui.StackRow, len(stacks))
	for i, s := range stacks {
		rows[i] = ui.StackRow{Name: s.Name, Containers: s.Containers, Status: s.Status, File: s.File}
	}
	return rows
}

// toContainerRows maps domain containers to presentation rows for ui.RenderContainerTable.
func toContainerRows(list []containers.ContainerDetail) []ui.ContainerRow {
	rows := make([]ui.ContainerRow, len(list))
	for i, c := range list {
		ports := make([]ui.PortMapping, len(c.Publishers))
		for j, p := range c.Publishers {
			ports[j] = ui.PortMapping{PublishedPort: p.PublishedPort, TargetPort: p.TargetPort}
		}
		rows[i] = ui.ContainerRow{
			Name:    c.Name,
			Service: c.Service,
			State:   c.State,
			Health:  c.Health,
			Image:   c.Image,
			Ports:   ports,
		}
	}
	return rows
}

func init() {
	statusCmd.Flags().BoolVar(&jsonFlag, "json", false, "Output status in JSON format")
	statusCmd.Flags().BoolVarP(&allFlagStatus, "all", "a", false, "Include all stacks")
	statusCmd.Flags().BoolVarP(&detailsFlag, "details", "d", false, "Show detailed container inspection")
	statusCmd.Flags().
		BoolVarP(&pagerFlag, "pager", "p", false, "Open status table inside an interactive scrollable viewport")
}
