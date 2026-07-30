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

var (
	followFlag bool
	tailFlag   string
)

var logsCmd = &cobra.Command{
	Use:   "logs <stack>",
	Short: "View logs from a Compose stack",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary).
			MarginBottom(1)
		if !ui.DisableProgress {
			fmt.Fprintln(log.Out, headerStyle.Render("📜 Docker Compose Service Logs"))
		}
		stackName := args[0]
		cfg := config.GetContainersConfig()
		mgr := containers.NewManager(cfg.Path)

		if err := mgr.Logs(stackName, followFlag, tailFlag); err != nil {
			return fmt.Errorf("failed to fetch logs for stack %s: %w", stackName, err)
		}
		return nil
	},
}

func init() {
	logsCmd.Flags().BoolVarP(&followFlag, "follow", "f", false, "Follow log output")
	logsCmd.Flags().StringVar(&tailFlag, "tail", "100", "Number of lines to show from the end of the logs")
}
