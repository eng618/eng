package compose

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
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
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary).
			MarginBottom(1)
		if !ui.DisableProgress {
			fmt.Fprintln(log.Out, headerStyle.Render("🐳 Discovered Docker Compose Stacks"))
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

		var boxLines []string
		boxLines = append(boxLines, fmt.Sprintf("Discovered %s compose stack(s) under %s:",
			theme.PrimaryText.Bold(true).Render(fmt.Sprintf("%d", len(stacks))),
			theme.BoldText.Render(cfg.Path),
		))
		boxLines = append(boxLines, "")
		boxLines = append(boxLines, fmt.Sprintf("  %-20s %-35s %s",
			theme.BoldText.Render("Stack"),
			theme.BoldText.Render("Path"),
			theme.BoldText.Render("Services"),
		))
		boxLines = append(boxLines, "  "+strings.Repeat("─", 65))

		for _, s := range stacks {
			svcs := strings.Join(s.Services, ", ")
			if svcs == "" {
				svcs = "-"
			}
			boxLines = append(boxLines, fmt.Sprintf("  %-20s %-35s %s",
				theme.PrimaryText.Render(s.Name),
				theme.MutedText.Render(s.Path),
				theme.BaseText.Render(svcs),
			))
		}

		if !ui.DisableProgress {
			boxStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(theme.Primary).
				Padding(0, 1).
				MarginBottom(1)
			fmt.Fprintln(log.Out, boxStyle.Render(strings.Join(boxLines, "\n")))
		} else {
			for _, line := range boxLines {
				log.Info("%s", line)
			}
		}

		theme.SuccessMessage(fmt.Sprintf("Listed %d Docker Compose stack(s)", len(stacks)))
		return nil
	},
}
