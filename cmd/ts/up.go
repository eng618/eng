package ts

import (
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/utils"
	"github.com/eng618/eng/internal/utils/log"
)

var UpCmd = &cobra.Command{
	Use:   "up",
	Short: "bring up the tailscale service",
	Long:  `This call 'sudo tailscale up' under the hood..`,
	RunE: func(cmd *cobra.Command, _args []string) error {
		log.Start("Bringing up the tailscale service")
		tsUpCmd := exec.Command("sudo", "tailscale", "up")
		err := utils.StartChildProcess(tsUpCmd)
		if err != nil {
			return err // Return the error for Cobra to handle
		}
		return nil // Indicate success
	},
}
