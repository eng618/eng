package ts

import (
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/log"
)

var execCommand = exec.Command

var TailscaleCmd = &cobra.Command{
	Use:   "tailscale",
	Short: "A helper for the tailscale command",
	Long:  `This command will help manage various aspects of Tailscale.`,
	Run: func(cmd *cobra.Command, _args []string) {
		log.Info("tailscale called")
	},
	Aliases: []string{"ts"},
}

func init() {
	TailscaleCmd.AddCommand(UpCmd)
	TailscaleCmd.AddCommand(DownCmd)
}
