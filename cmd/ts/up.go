package ts

import (
	"os/exec"

	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/log"
	"github.com/spf13/cobra"
)

var UpCmd = &cobra.Command{
	Use:   "up",
	Short: "bring up the tailscale service",
	Long:  `This call 'sudo tailscale up' under the hood..`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Start("Bringing up the tailscale service")
		tsUpCmd := exec.Command("sudo", "tailscale", "up")
		err := utils.StartChildProcess(tsUpCmd)
		if err != nil {
			return err // Return the error for Cobra to handle
		}
		return nil // Indicate success
	},
}
