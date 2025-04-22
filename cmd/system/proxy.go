package system

import (
	"fmt"
	"os"

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
		fmt.Println("Current proxy configuration:")
		fmt.Println("Config proxy:", proxy)
		fmt.Println("Enabled:", enabled)
		fmt.Println("-------------------------------------------------")
		fmt.Println("System environment variables:")
		fmt.Println("ALL_PROXY:", os.Getenv("ALL_PROXY"))
		fmt.Println("HTTP_PROXY:", os.Getenv("HTTP_PROXY"))
		fmt.Println("HTTPS_PROXY:", os.Getenv("HTTPS_PROXY"))
		fmt.Println("GLOBAL_AGENT_HTTP_PROXY:", os.Getenv("GLOBAL_AGENT_HTTP_PROXY"))
		fmt.Println("NO_PROXY:", os.Getenv("NO_PROXY"))
		fmt.Println("-------------------------------------------------")
		fmt.Println("Lowercase environment variables:")
		fmt.Println("http_proxy:", os.Getenv("http_proxy"))
		fmt.Println("https_proxy:", os.Getenv("https_proxy"))
		fmt.Println("no_proxy:", os.Getenv("no_proxy"))
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
