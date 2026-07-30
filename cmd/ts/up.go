package ts

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

var UpCmd = &cobra.Command{
	Use:   "up",
	Short: "bring up the tailscale service",
	Long:  `This call 'sudo tailscale up' under the hood..`,
	RunE: func(cmd *cobra.Command, _args []string) error {
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary).
			MarginBottom(1)
		if !ui.DisableProgress {
			fmt.Fprintln(log.Out, headerStyle.Render("🐉 Tailscale Service Up"))
		}

		log.Start("Bringing up the tailscale service")
		tsUpCmd := execCommand("sudo", "tailscale", "up")
		err := cmdutil.StartChildProcess(tsUpCmd)
		if err != nil {
			return err // Return the error for Cobra to handle
		}
		return nil // Indicate success
	},
}
