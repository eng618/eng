package ts

import (
	"os/exec"

	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/log"
	"github.com/spf13/cobra"
)

var DownCmd = &cobra.Command{
	Use:   "down",
	Short: "take down the tailscale service",
	Long:  `This call 'sudo tailscale down' under the hood..`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Taking down the tailscale service")
		tsDownCmd := exec.Command("sudo", "tailscale", "down")
		utils.StartChildProcess(tsDownCmd)
	},
}
