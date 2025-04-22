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
	Short: "Show or configure system proxies",
	Long:  `This command displays and manages multiple proxy configurations and allows enabling, disabling, or setting them via flags or interactive prompts.`,
	Run: func(cmd *cobra.Command, args []string) {
		proxies, activeIndex := config.GetProxyConfigs()

		fmt.Println("Proxy Configurations:")
		fmt.Println("-------------------------------------------------")

		if len(proxies) == 0 {
			fmt.Println("No proxy configurations found.")
		} else {
			for i, p := range proxies {
				status := " "
				if p.Enabled {
					status = "*"
				}
				fmt.Printf("[%s] %d. %s - %s\n", status, i+1, p.Title, p.Value)
			}
		}

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

		// Check if any flags are set that would skip the interactive mode
		add, _ := cmd.Flags().GetBool("add")
		enableIndex, _ := cmd.Flags().GetInt("enable")
		disableAll, _ := cmd.Flags().GetBool("disable")
		select_, _ := cmd.Flags().GetBool("select")

		// If no action flags are set, prompt for selection
		if !add && enableIndex < 0 && !disableAll && !select_ {
			if activeIndex >= 0 && activeIndex < len(proxies) {
				fmt.Printf("\nActive proxy: %s (%s)\n", proxies[activeIndex].Title, proxies[activeIndex].Value)
			} else {
				fmt.Println("\nNo active proxy configured.")
			}
		}
	},
}

func init() {
	ProxyCmd.Flags().Bool("add", false, "Add a new proxy configuration")
	ProxyCmd.Flags().Int("enable", -1, "Enable a specific proxy by index")
	ProxyCmd.Flags().Bool("disable", false, "Disable all proxies")
	ProxyCmd.Flags().Bool("select", false, "Interactively select a proxy to enable")

	ProxyCmd.PreRun = func(cmd *cobra.Command, args []string) {
		add, _ := cmd.Flags().GetBool("add")
		enableIndex, _ := cmd.Flags().GetInt("enable")
		disableAll, _ := cmd.Flags().GetBool("disable")
		select_, _ := cmd.Flags().GetBool("select")

		proxies, _ := config.GetProxyConfigs()

		if add {
			// Add a new proxy configuration
			config.AddOrUpdateProxy()
			return
		}

		if enableIndex >= 0 {
			// Convert from 1-based index (user-friendly) to 0-based index
			index := enableIndex - 1
			if index >= 0 && index < len(proxies) {
				_, err := config.EnableProxy(index, proxies)
				if err != nil {
					log.Error("Failed to enable proxy: %v", err)
					cmd.SilenceUsage = true
					return
				}
			} else {
				log.Error("Invalid proxy index: %d. Valid range is 1-%d", enableIndex, len(proxies))
				cmd.SilenceUsage = true
				return
			}
		}

		if disableAll {
			// Disable all proxies
			if err := config.DisableAllProxies(); err != nil {
				log.Error("Failed to disable proxies: %v", err)
				cmd.SilenceUsage = true
				return
			}
			log.Success("All proxies disabled")
			return
		}

		if select_ {
			// Interactive selection
			if len(proxies) == 0 {
				// If no proxies, add one first
				proxies, _ = config.AddOrUpdateProxy()
			}

			selectedIndex, err := config.SelectProxy(proxies)
			if err != nil {
				log.Error("Failed to select proxy: %v", err)
				cmd.SilenceUsage = true
				return
			}

			_, err = config.EnableProxy(selectedIndex, proxies)
			if err != nil {
				log.Error("Failed to enable proxy: %v", err)
				cmd.SilenceUsage = true
				return
			}
		}
	}
}
