package ts

import (
	"os/exec"
	"github.com/spf13/cobra"
	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/log"
)

var UpCmd = &cobra.Command{
	Use:   "up",
	Short: "bring up the tailscale service",
	Long:  `This call 'sudo tailscale up' under the hood..`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Bringing up the tailscale service")
		tsUpCmd := exec.Command("sudo", "tailscale", "up")
		utils.StartChildProcess(tsUpCmd)
	},
}
