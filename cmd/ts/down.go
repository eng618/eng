package ts

import (
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/log"
)

var DownCmd = &cobra.Command{
	Use:   "down",
	Short: "take down the tailscale service",
	Long:  `This call 'sudo tailscale down' under the hood..`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Start("Taking down the tailscale service")
		tsDownCmd := exec.Command("sudo", "tailscale", "down")
		err := utils.StartChildProcess(tsDownCmd)
		if err != nil {
			return err // Return the error for Cobra to handle
		}
		return nil // Indicate success
	},
}
