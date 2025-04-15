package system

import (
	"fmt"

	"github.com/eng618/eng/utils/config"
	"github.com/eng618/eng/utils/log"
	"github.com/spf13/cobra"
)

var ProxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Show or configure the system proxy",
	Long:  `This command displays the current system proxy and allows enabling or disabling it via flags or prompts.`,
	Run: func(cmd *cobra.Command, args []string) {
		proxy, enabled := config.ProxyConfig()
		fmt.Printf("Current proxy: %s\nEnabled: %v\n", proxy, enabled)
	},
}

func init() {
	ProxyCmd.Flags().Bool("enable", false, "Enable the proxy")
	ProxyCmd.Flags().Bool("disable", false, "Disable the proxy")
	ProxyCmd.Flags().Bool("set", false, "Interactively set the proxy value")
	ProxyCmd.PreRun = func(cmd *cobra.Command, args []string) {
		enable, _ := cmd.Flags().GetBool("enable")
		disable, _ := cmd.Flags().GetBool("disable")
		set, _ := cmd.Flags().GetBool("set")
		if enable && disable {
			log.Error("Cannot enable and disable proxy at the same time.")
			err := cmd.Help()
			cobra.CheckErr(err)
			cmd.SilenceUsage = true
			return
		}
		if enable {
			config.SetProxyEnabled(true)
		}
		if disable {
			config.SetProxyEnabled(false)
		}
		if set {
			config.ProxyConfig()
		}
	}
}
