package compose

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/containers"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

var allFlagPull bool

var pullCmd = &cobra.Command{
	Use:   "pull [stack...]",
	Short: "Pull latest service images for Compose stacks",
	RunE: func(cmd *cobra.Command, args []string) error {
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary).
			MarginBottom(1)
		if !ui.DisableProgress {
			fmt.Fprintln(log.Out, headerStyle.Render("📥 Pulling Docker Compose Service Images"))
		}

		if len(args) == 0 && !allFlagPull {
			return fmt.Errorf("specify at least one stack name or use --all (-a)")
		}

		cfg := config.GetContainersConfig()
		mgr := containers.NewManager(cfg.Path)

		targets := args
		if allFlagPull {
			targets = []string{"all"}
		}

		theme.InfoMessage("Pulling latest images...")
		if err := mgr.Pull(targets); err != nil {
			return err
		}

		theme.SuccessMessage("Compose stack image(s) pulled successfully.")
		return nil
	},
}

func init() {
	pullCmd.Flags().BoolVarP(&allFlagPull, "all", "a", false, "Pull images for all discovered stacks")
}
