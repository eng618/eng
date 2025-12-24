package system

import (
	"fmt"
	"os"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/utils/config"
	"github.com/eng618/eng/utils/log"
)

const (
	msgUpdatedProxyConfigurations = "Updated proxy configurations:"
	msgFailedEnableProxyFmt       = "Failed to enable proxy: %v"
)

var ProxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Show or configure system proxies",
	Long:  `This command displays and manages multiple proxy configurations and allows enabling, disabling, or setting them via subcommands.`,
	Run: func(cmd *cobra.Command, args []string) {
		listProxyConfigurations(cmd)
	},
}

// Common function to list proxy configurations.
func listProxyConfigurations(cmd *cobra.Command) {
	compact, _ := cmd.Flags().GetBool("compact")
	showEnv, _ := cmd.Flags().GetBool("env")
	showLowercaseEnv, _ := cmd.Flags().GetBool("lowercase-env")
	proxies, activeIndex := config.GetProxyConfigs()

	renderProxyList(compact, proxies)
	if showEnv {
		renderEnv(compact, showLowercaseEnv)
	}
	renderActive(compact, proxies, activeIndex)
	renderNote(compact)
}

func renderProxyList(compact bool, proxies []config.ProxyConfig) {
	if compact {
		fmt.Println("Proxies (★ active, • inactive):")
		if len(proxies) == 0 {
			fmt.Println("- none configured")
			return
		}
		for _, p := range proxies {
			fmt.Printf("- %s\n", config.FormatProxyOption(p))
		}
		return
	}
	fmt.Println("Proxy Configurations (★ active, • inactive):")
	fmt.Println("-------------------------------------------------")
	if len(proxies) == 0 {
		fmt.Println("No proxy configurations found.")
	} else {
		for i, p := range proxies {
			fmt.Printf("%d. %s\n", i+1, config.FormatProxyOption(p))
		}
	}
	fmt.Println("-------------------------------------------------")
}

func renderEnv(compact bool, showLowercase bool) {
	if !compact {
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
		fmt.Println("-------------------------------------------------")
		return
	}
	all := os.Getenv("ALL_PROXY")
	http := os.Getenv("HTTP_PROXY")
	https := os.Getenv("HTTPS_PROXY")
	global := os.Getenv("GLOBAL_AGENT_HTTP_PROXY")
	noProxy := os.Getenv("NO_PROXY")
	same := all == http && http == https && https == global
	if same {
		fmt.Printf("Env: ALL/HTTP/HTTPS/GLOBAL=%s, NO_PROXY=%s\n", all, noProxy)
	} else {
		fmt.Printf("Env: ALL=%s HTTP=%s HTTPS=%s GLOBAL=%s NO_PROXY=%s\n", all, http, https, global, noProxy)
	}
	if showLowercase {
		lhttp := os.Getenv("http_proxy")
		lhttps := os.Getenv("https_proxy")
		lno := os.Getenv("no_proxy")
		fmt.Printf("Env (lowercase): http=%s https=%s no=%s\n", lhttp, lhttps, lno)
	}
}

func renderActive(compact bool, proxies []config.ProxyConfig, activeIndex int) {
	if activeIndex >= 0 && activeIndex < len(proxies) {
		if compact {
			fmt.Printf("Active: %s\n", config.FormatProxyOption(proxies[activeIndex]))
		} else {
			fmt.Printf("\nActive proxy: %s\n", config.FormatProxyOption(proxies[activeIndex]))
		}
		return
	}
	if compact {
		fmt.Println("Active: none")
	} else {
		fmt.Println("\nNo active proxy configured.")
	}
}

func renderNote(compact bool) {
	if compact {
		return
	}
	fmt.Println("\nNote: Environment variable changes only affect the current process.")
	fmt.Println("For system-wide changes, you may need to restart your terminal or source your profile.")
	fmt.Println("To apply in your current shell, you can run:")
	fmt.Println("  eval $(eng system proxy --export)")
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new proxy configuration",
	Long:  `Add a new proxy configuration with a title and address.`,
	Run: func(cmd *cobra.Command, args []string) {
		config.AddOrUpdateProxy()
		fmt.Println(msgUpdatedProxyConfigurations)
		listProxyConfigurations(cmd)
	},
}

var enableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Interactively select a proxy to enable",
	Long:  `Displays an interactive prompt to select a proxy configuration to enable.`,
	Run: func(cmd *cobra.Command, args []string) {
		proxies, _ := config.GetProxyConfigs()

		idxFlag, _ := cmd.Flags().GetInt("index")
		titleFlag, _ := cmd.Flags().GetString("title")
		quietFlag, _ := cmd.Flags().GetBool("quiet")

		// If no proxies, add one first interactively
		if len(proxies) == 0 {
			log.Info("No proxy configurations found. Adding a new one...")
			proxies, _ = config.AddOrUpdateProxy()
		}

		selectedIndex := -1
		if idxFlag >= 0 && idxFlag < len(proxies) {
			selectedIndex = idxFlag
		} else if titleFlag != "" {
			selectedIndex = config.FindProxyIndexByTitle(proxies, titleFlag)
			if selectedIndex < 0 {
				log.Error("No proxy found with title '%s'", titleFlag)
				return
			}
		} else {
			// Fall back to interactive selection
			var err error
			selectedIndex, err = config.SelectProxy(proxies)
			if err != nil {
				log.Error("Failed to select proxy: %v", err)
				return
			}
		}

		if _, err := config.EnableProxy(selectedIndex, proxies); err != nil {
			log.Error(msgFailedEnableProxyFmt, err)
			return
		}

		log.Success("Proxy '%s' selected and enabled", proxies[selectedIndex].Title)
		if !quietFlag {
			listProxyConfigurations(cmd)
		}
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
		listProxyConfigurations(cmd)
	},
}

// Add a new export subcommand to enable easy exporting of proxy settings to shell.
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export proxy settings as environment variables for the current shell",
	Long:  `Generates shell commands to export proxy settings as environment variables for the current shell.`,
	Run: func(cmd *cobra.Command, args []string) {
		proxies, activeIndex := config.GetProxyConfigs()

		if activeIndex >= 0 && activeIndex < len(proxies) {
			proxyValue := proxies[activeIndex].Value
			fmt.Printf("export ALL_PROXY='%s'\n", proxyValue)
			fmt.Printf("export HTTP_PROXY='%s'\n", proxyValue)
			fmt.Printf("export HTTPS_PROXY='%s'\n", proxyValue)
			fmt.Printf("export GLOBAL_AGENT_HTTP_PROXY='%s'\n", proxyValue)
			fmt.Printf("export http_proxy='%s'\n", proxyValue)
			fmt.Printf("export https_proxy='%s'\n", proxyValue)

			// Combine default no_proxy with any custom settings
			noProxyValue := "localhost,127.0.0.1,::1,.local"
			if proxies[activeIndex].NoProxy != "" {
				noProxyValue = noProxyValue + "," + proxies[activeIndex].NoProxy
			}

			fmt.Printf("export NO_PROXY='%s'\n", noProxyValue)
			fmt.Printf("export no_proxy='%s'\n", noProxyValue)
		} else {
			// If no active proxy, output commands to unset variables
			fmt.Println("unset ALL_PROXY")
			fmt.Println("unset HTTP_PROXY")
			fmt.Println("unset HTTPS_PROXY")
			fmt.Println("unset GLOBAL_AGENT_HTTP_PROXY")
			fmt.Println("unset NO_PROXY")
			fmt.Println("unset http_proxy")
			fmt.Println("unset https_proxy")
			fmt.Println("unset no_proxy")
		}
	},
}

var toggleCmd = &cobra.Command{
	Use:   "toggle",
	Short: "Toggle proxies on or off",
	Long:  `Toggles proxies on or off. When toggling on, select an existing proxy or create a new one.`,
	Run: func(cmd *cobra.Command, args []string) {
		onFlag, _ := cmd.Flags().GetBool("on")
		offFlag, _ := cmd.Flags().GetBool("off")
		quietFlag, _ := cmd.Flags().GetBool("quiet")
		idxFlag, _ := cmd.Flags().GetInt("index")
		titleFlag, _ := cmd.Flags().GetString("title")

		proxies, activeIndex := config.GetProxyConfigs()

		// Decide action: explicit flags win; otherwise toggle based on current state
		doOff := offFlag || (!onFlag && activeIndex >= 0)
		doOn := onFlag || (!offFlag && activeIndex < 0)

		if doOff {
			if err := config.DisableAllProxies(); err != nil {
				log.Error("Failed to disable proxies: %v", err)
				return
			}
			log.Success("All proxies disabled")
			if !quietFlag {
				listProxyConfigurations(cmd)
			}
			// If explicitly off, and not also asked to turn on, return.
			if offFlag && !onFlag {
				return
			}
		}

		if doOn {
			// Determine selection path
			selectedIndex := -1

			if idxFlag >= 0 && idxFlag < len(proxies) {
				selectedIndex = idxFlag
			} else if titleFlag != "" {
				selectedIndex = config.FindProxyIndexByTitle(proxies, titleFlag)
				if selectedIndex < 0 {
					log.Error("No proxy found with title '%s'", titleFlag)
					return
				}
			} else {
				// Interactive: existing list plus "Create new…"
				if len(proxies) == 0 {
					// No proxies yet: create new interactively
					var idx int
					proxies, idx = config.AddOrUpdateProxy()
					selectedIndex = idx
				} else {
					// Build options: existing proxies + create new
					options := make([]string, 0, len(proxies)+1)
					for _, p := range proxies {
						options = append(options, config.FormatProxyOption(p))
					}
					options = append(options, "Create new…")

					var sel int
					prompt := &survey.Select{
						Message: "Select a proxy to enable or create new:",
						Options: options,
						Help:    "Use arrow keys to navigate, and Enter to select.",
					}
					if err := survey.AskOne(prompt, &sel); err != nil {
						log.Error("Selection cancelled: %v", err)
						return
					}
					if sel == len(options)-1 {
						var idx int
						proxies, idx = config.AddOrUpdateProxy()
						selectedIndex = idx
					} else {
						selectedIndex = sel
					}
				}
			}

			if _, err := config.EnableProxy(selectedIndex, proxies); err != nil {
				log.Error(msgFailedEnableProxyFmt, err)
				return
			}
			log.Success("Proxy '%s' selected and enabled", proxies[selectedIndex].Title)
			if !quietFlag {
				listProxyConfigurations(cmd)
			}
		}
	},
}

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Add or update a proxy configuration",
	Long:  `Add or update a proxy configuration via flags; optionally enable it. Use --interactive to prompt for missing values.`,
	Run: func(cmd *cobra.Command, args []string) {
		title, _ := cmd.Flags().GetString("title")
		value, _ := cmd.Flags().GetString("value")
		noProxy, _ := cmd.Flags().GetString("no-proxy")
		enableAfter, _ := cmd.Flags().GetBool("enable")
		interactive, _ := cmd.Flags().GetBool("interactive")

		if interactive || title == "" || value == "" {
			// Fall back to interactive add/update
			proxies, idx := config.AddOrUpdateProxy()
			if enableAfter {
				if _, err := config.EnableProxy(idx, proxies); err != nil {
					log.Error(msgFailedEnableProxyFmt, err)
					return
				}
				log.Success("Proxy '%s' enabled", proxies[idx].Title)
			}
			fmt.Println(msgUpdatedProxyConfigurations)
			listProxyConfigurations(cmd)
			return
		}

		proxies, idx, err := config.AddOrUpdateProxyWithValues(title, value, noProxy)
		if err != nil {
			log.Error("Failed to set proxy: %v", err)
			return
		}
		log.Success("Proxy '%s' added/updated", title)

		if enableAfter {
			if _, err := config.EnableProxy(idx, proxies); err != nil {
				log.Error(msgFailedEnableProxyFmt, err)
				return
			}
			log.Success("Proxy '%s' enabled", proxies[idx].Title)
		}

		fmt.Println(msgUpdatedProxyConfigurations)
		listProxyConfigurations(cmd)
	},
}

func init() {
	// Add subcommands to the proxy command
	ProxyCmd.AddCommand(addCmd)
	ProxyCmd.AddCommand(enableCmd)
	ProxyCmd.AddCommand(disableCmd)
	ProxyCmd.AddCommand(exportCmd)
	ProxyCmd.AddCommand(toggleCmd)
	ProxyCmd.AddCommand(setCmd)

	// Persistent flags to control listing style
	ProxyCmd.PersistentFlags().Bool("compact", true, "Show compact status output")
	ProxyCmd.PersistentFlags().Bool("env", false, "Include environment variables in status output")
	ProxyCmd.PersistentFlags().Bool("lowercase-env", false, "Include lowercase environment vars in compact mode")

	// Flags for set command
	setCmd.Flags().String("title", "", "Proxy configuration title")
	setCmd.Flags().String("value", "", "Proxy address (e.g., http://host:port)")
	setCmd.Flags().String("no-proxy", "", "Additional no_proxy values (comma-separated)")
	setCmd.Flags().Bool("enable", false, "Enable this proxy after setting")
	setCmd.Flags().Bool("interactive", false, "Use interactive prompts when missing values")

	// Flags for toggle command
	toggleCmd.Flags().Bool("on", false, "Toggle on (enable a proxy)")
	toggleCmd.Flags().Bool("off", false, "Toggle off (disable all proxies)")
	toggleCmd.Flags().Bool("quiet", false, "Suppress status output after toggling")
	toggleCmd.Flags().Int("index", -1, "Enable proxy by index (non-interactive)")
	toggleCmd.Flags().String("title", "", "Enable proxy by title (non-interactive)")

	// Flags for enable command
	enableCmd.Flags().Int("index", -1, "Enable proxy by index (non-interactive)")
	enableCmd.Flags().String("title", "", "Enable proxy by title (non-interactive)")
	enableCmd.Flags().Bool("quiet", false, "Suppress status output after enabling")
}
