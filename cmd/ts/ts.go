package ts

import (
	"fmt"

	"github.com/spf13/cobra"
)

var TailscaleCmd = &cobra.Command{
	Use:   "tailscale",
	Short: "A helper for the tailscale command",
	Long:  `This command will help manage various aspects of Tailscale.`,
	Run: func(cmd *cobra.Command, _args []string) {
		fmt.Println("tailscale called")
	},
	Aliases: []string{"ts"},
}

func init() {
	TailscaleCmd.AddCommand(UpCmd)
	TailscaleCmd.AddCommand(DownCmd)
}
