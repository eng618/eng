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
	Long:  `This command displays and manages multiple proxy configurations and allows enabling, disabling, or setting them via subcommands.`,
	Run: func(cmd *cobra.Command, args []string) {
		listProxyConfigurations()
	},
}

// Common function to list proxy configurations
func listProxyConfigurations() {
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

	if activeIndex >= 0 && activeIndex < len(proxies) {
		fmt.Printf("\nActive proxy: %s (%s)\n", proxies[activeIndex].Title, proxies[activeIndex].Value)
	} else {
		fmt.Println("\nNo active proxy configured.")
	}
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new proxy configuration",
	Long:  `Add a new proxy configuration with a title and address.`,
	Run: func(cmd *cobra.Command, args []string) {
		config.AddOrUpdateProxy()
		fmt.Println("Updated proxy configurations:")
		listProxyConfigurations()
	},
}

var enableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Interactively select a proxy to enable",
	Long:  `Displays an interactive prompt to select a proxy configuration to enable.`,
	Run: func(cmd *cobra.Command, args []string) {
		proxies, _ := config.GetProxyConfigs()

		if len(proxies) == 0 {
			// If no proxies, add one first
			log.Info("No proxy configurations found. Adding a new one...")
			proxies, _ = config.AddOrUpdateProxy()
		}

		selectedIndex, err := config.SelectProxy(proxies)
		if err != nil {
			log.Error("Failed to select proxy: %v", err)
			return
		}

		_, err = config.EnableProxy(selectedIndex, proxies)
		if err != nil {
			log.Error("Failed to enable proxy: %v", err)
			return
		}

		log.Success("Proxy '%s' selected and enabled", proxies[selectedIndex].Title)
		listProxyConfigurations()
	},
}

var disableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable all proxies",
	Long:  `Disable all proxy configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.DisableAllProxies(); err != nil {
			log.Error("Failed to disable proxies: %v", err)
			return
		}
		log.Success("All proxies disabled")
		listProxyConfigurations()
	},
}

func init() {
	// Add subcommands to the proxy command
	ProxyCmd.AddCommand(addCmd)
	ProxyCmd.AddCommand(enableCmd)
	ProxyCmd.AddCommand(disableCmd)
}
