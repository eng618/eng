package system

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
)

var statusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"ls", "list"},
	Short:   "Show active proxy status, profiles, and shell env vars",
	Long:    `Displays current active proxy status, list of stored configurations, and environment variable states.`,
	Run: func(cmd *cobra.Command, args []string) {
		listProxyConfigurations(cmd)
	},
}

var addCmd = &cobra.Command{
	Use:     "add [title] [url]",
	Aliases: []string{"create", "new"},
	Short:   "Add a new proxy configuration",
	Long:    `Add a new proxy configuration with a title, proxy address, and optional bypass domains.`,
	Run: func(cmd *cobra.Command, args []string) {
		titleFlag, _ := cmd.Flags().GetString("title")
		urlFlag, _ := cmd.Flags().GetString("url")
		if urlFlag == "" {
			urlFlag, _ = cmd.Flags().GetString("value")
		}
		noProxyFlag, _ := cmd.Flags().GetString("no-proxy")
		enableAfter, _ := cmd.Flags().GetBool("enable")

		var titleVal, urlVal string
		if len(args) > 0 {
			titleVal = args[0]
		} else {
			titleVal = titleFlag
		}
		if len(args) > 1 {
			urlVal = args[1]
		} else {
			urlVal = urlFlag
		}

		var proxies []config.ProxyConfig
		var idx int
		if titleVal != "" && urlVal != "" {
			var err error
			proxies, idx, err = config.AddOrUpdateProxyWithValues(titleVal, urlVal, noProxyFlag)
			if err != nil {
				log.Error("Failed to add proxy: %v", err)
				return
			}
			log.Success("Proxy '%s' added successfully", titleVal)
		} else {
			proxies, idx = config.AddOrUpdateProxy()
		}

		if idx >= 0 && enableAfter {
			if _, err := config.EnableProxy(idx, proxies); err != nil {
				log.Error(msgFailedEnableProxyFmt, err)
				return
			}
			log.Success("Proxy '%s' enabled", proxies[idx].Title)
		}

		fmt.Fprintln(log.Out, msgUpdatedProxyConfigurations)
		listProxyConfigurations(cmd)
	},
}

var useCmd = &cobra.Command{
	Use:     "use [name|index]",
	Aliases: []string{"enable", "switch", "select"},
	Short:   "Activate a proxy configuration",
	Long:    `Select and enable a proxy configuration interactively or by name/index.`,
	Run: func(cmd *cobra.Command, args []string) {
		proxies, _ := config.GetProxyConfigs()

		idxFlag, _ := cmd.Flags().GetInt("index")
		titleFlag, _ := cmd.Flags().GetString("title")
		quietFlag, _ := cmd.Flags().GetBool("quiet")

		targetArg := ""
		if len(args) > 0 {
			targetArg = args[0]
		}

		// If no proxies, prompt to add one first
		if len(proxies) == 0 {
			log.Info("No proxy configurations found. Adding a new one...")
			proxies, _ = config.AddOrUpdateProxy()
			if len(proxies) == 0 {
				return
			}
		}

		selectedIndex := resolveProxyIndex(targetArg, idxFlag, titleFlag, proxies)
		if selectedIndex < 0 {
			if targetArg != "" || titleFlag != "" {
				log.Error("No proxy configuration found matching identifier")
				return
			}
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

var offCmd = &cobra.Command{
	Use:     "off",
	Aliases: []string{"disable", "unset", "clear"},
	Short:   "Deactivate all proxies and unset environment variables",
	Long:    `Disables all proxy configurations and unsets all shell proxy environment variables.`,
	Run: func(cmd *cobra.Command, args []string) {
		quietFlag, _ := cmd.Flags().GetBool("quiet")
		if err := config.DisableAllProxies(); err != nil {
			log.Error("Failed to disable proxies: %v", err)
			return
		}
		log.Success("All proxies disabled")
		if !quietFlag {
			listProxyConfigurations(cmd)
		}
	},
}

var exportCmd = &cobra.Command{
	Use:     "export",
	Aliases: []string{"shell", "env"},
	Short:   "Export proxy settings as environment variables for current shell",
	Long:    `Generates shell export/unset statements for eval: eval $(eng system proxy export)`,
	Run: func(cmd *cobra.Command, args []string) {
		// Machine-consumed output for eval: silence log chatter while loading
		// config so only export/unset statements reach stdout. Errors stay live.
		prevOut := log.Out
		log.SetWriters(io.Discard, log.Err)
		proxies, activeIndex := config.GetProxyConfigs()
		log.SetWriters(prevOut, log.Err)

		if activeIndex >= 0 && activeIndex < len(proxies) {
			proxyValue := proxies[activeIndex].Value
			fmt.Fprintf(log.Out, "export ALL_PROXY='%s'\n", proxyValue)
			fmt.Fprintf(log.Out, "export HTTP_PROXY='%s'\n", proxyValue)
			fmt.Fprintf(log.Out, "export HTTPS_PROXY='%s'\n", proxyValue)
			fmt.Fprintf(log.Out, "export GLOBAL_AGENT_HTTP_PROXY='%s'\n", proxyValue)
			fmt.Fprintf(log.Out, "export http_proxy='%s'\n", proxyValue)
			fmt.Fprintf(log.Out, "export https_proxy='%s'\n", proxyValue)

			noProxyValue := "localhost,127.0.0.1,::1,.local"
			if proxies[activeIndex].NoProxy != "" {
				noProxyValue = noProxyValue + "," + proxies[activeIndex].NoProxy
			}

			fmt.Fprintf(log.Out, "export NO_PROXY='%s'\n", noProxyValue)
			fmt.Fprintf(log.Out, "export no_proxy='%s'\n", noProxyValue)
		} else {
			fmt.Fprintln(log.Out, "unset ALL_PROXY")
			fmt.Fprintln(log.Out, "unset HTTP_PROXY")
			fmt.Fprintln(log.Out, "unset HTTPS_PROXY")
			fmt.Fprintln(log.Out, "unset GLOBAL_AGENT_HTTP_PROXY")
			fmt.Fprintln(log.Out, "unset NO_PROXY")
			fmt.Fprintln(log.Out, "unset http_proxy")
			fmt.Fprintln(log.Out, "unset https_proxy")
			fmt.Fprintln(log.Out, "unset no_proxy")
		}
	},
}

var toggleCmd = &cobra.Command{
	Use:     "toggle",
	Aliases: []string{"on-off"},
	Short:   "Toggle proxies on or off",
	Long:    `Toggles proxies on or off. When toggling on, select an existing proxy or create a new one.`,
	Run: func(cmd *cobra.Command, args []string) {
		onFlag, _ := cmd.Flags().GetBool("on")
		offFlag, _ := cmd.Flags().GetBool("off")
		quietFlag, _ := cmd.Flags().GetBool("quiet")
		idxFlag, _ := cmd.Flags().GetInt("index")
		titleFlag, _ := cmd.Flags().GetString("title")

		proxies, activeIndex := config.GetProxyConfigs()

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
			if offFlag && !onFlag {
				return
			}
		}

		if doOn {
			var selectedIndex int
			if idxFlag >= 0 && idxFlag < len(proxies) {
				selectedIndex = idxFlag
			} else if titleFlag != "" {
				selectedIndex = config.FindProxyIndexByTitle(proxies, titleFlag)
				if selectedIndex < 0 {
					log.Error("No proxy found with title '%s'", titleFlag)
					return
				}
			} else {
				if len(proxies) == 0 {
					var idx int
					proxies, idx = config.AddOrUpdateProxy()
					selectedIndex = idx
				} else {
					options := make([]string, 0, len(proxies)+1)
					for _, p := range proxies {
						options = append(options, config.FormatProxyOption(p))
					}
					options = append(options, "Create new…")

					selected, err := ui.Select("Select a proxy to enable or create new:", options, "")
					if err != nil {
						log.Error("Selection canceled: %v", err)
						return
					}

					sel := -1
					for i, opt := range options {
						if opt == selected {
							sel = i
							break
						}
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

var editCmd = &cobra.Command{
	Use:     "edit [name|index]",
	Aliases: []string{"update", "set"},
	Short:   "Edit an existing proxy configuration",
	Long:    `Modify an existing proxy configuration via flags or interactively.`,
	Run: func(cmd *cobra.Command, args []string) {
		titleFlag, _ := cmd.Flags().GetString("title")
		valueFlag, _ := cmd.Flags().GetString("value")
		if valueFlag == "" {
			valueFlag, _ = cmd.Flags().GetString("url")
		}
		noProxyFlag, _ := cmd.Flags().GetString("no-proxy")
		enableAfter, _ := cmd.Flags().GetBool("enable")
		interactive, _ := cmd.Flags().GetBool("interactive")

		proxies, _ := config.GetProxyConfigs()
		targetArg := ""
		if len(args) > 0 {
			targetArg = args[0]
		}

		targetIdx := resolveProxyIndex(targetArg, -1, titleFlag, proxies)

		if interactive || (titleFlag == "" && targetIdx < 0 && valueFlag == "") {
			proxies, idx := config.AddOrUpdateProxy()
			if enableAfter && idx >= 0 {
				if _, err := config.EnableProxy(idx, proxies); err != nil {
					log.Error(msgFailedEnableProxyFmt, err)
					return
				}
				log.Success("Proxy '%s' enabled", proxies[idx].Title)
			}
			fmt.Fprintln(log.Out, msgUpdatedProxyConfigurations)
			listProxyConfigurations(cmd)
			return
		}

		targetTitle := titleFlag
		if targetIdx >= 0 && targetIdx < len(proxies) {
			targetTitle = proxies[targetIdx].Title
		}

		if targetTitle == "" {
			log.Error("Please specify a proxy title or index to edit")
			return
		}

		if valueFlag == "" && targetIdx >= 0 {
			valueFlag = proxies[targetIdx].Value
		}

		proxies, idx, err := config.AddOrUpdateProxyWithValues(targetTitle, valueFlag, noProxyFlag)
		if err != nil {
			log.Error("Failed to update proxy: %v", err)
			return
		}
		log.Success("Proxy '%s' updated", targetTitle)

		if enableAfter && idx >= 0 {
			if _, err := config.EnableProxy(idx, proxies); err != nil {
				log.Error(msgFailedEnableProxyFmt, err)
				return
			}
			log.Success("Proxy '%s' enabled", proxies[idx].Title)
		}

		fmt.Fprintln(log.Out, msgUpdatedProxyConfigurations)
		listProxyConfigurations(cmd)
	},
}

var removeCmd = &cobra.Command{
	Use:     "remove [name|index]",
	Aliases: []string{"rm", "delete"},
	Short:   "Remove a proxy configuration",
	Long:    `Deletes a stored proxy configuration profile.`,
	Run: func(cmd *cobra.Command, args []string) {
		proxies, _ := config.GetProxyConfigs()
		if len(proxies) == 0 {
			log.Info("No proxy configurations found to remove.")
			return
		}

		idxFlag, _ := cmd.Flags().GetInt("index")
		titleFlag, _ := cmd.Flags().GetString("title")

		targetArg := ""
		if len(args) > 0 {
			targetArg = args[0]
		}

		targetIdx := resolveProxyIndex(targetArg, idxFlag, titleFlag, proxies)
		if targetIdx < 0 {
			var err error
			targetIdx, err = config.SelectProxy(proxies)
			if err != nil {
				log.Error("Failed to select proxy to remove: %v", err)
				return
			}
		}

		updatedProxies, err := config.RemoveProxy(targetIdx)
		if err != nil {
			log.Error("Failed to remove proxy: %v", err)
			return
		}

		listProxyConfigurations(cmd)
		_ = updatedProxies
	},
}

var testCmd = &cobra.Command{
	Use:     "test [name|index]",
	Aliases: []string{"check", "ping"},
	Short:   "Test HTTP connection through a proxy",
	Long:    `Sends a test HTTP request through the specified proxy or active proxy to verify connectivity.`,
	Run: func(cmd *cobra.Command, args []string) {
		targetURL, _ := cmd.Flags().GetString("target")
		if targetURL == "" {
			targetURL = "https://1.1.1.1"
		}

		proxies, activeIdx := config.GetProxyConfigs()
		if len(proxies) == 0 {
			log.Error("No proxy configurations stored to test.")
			return
		}

		targetArg := ""
		if len(args) > 0 {
			targetArg = args[0]
		}

		idxFlag, _ := cmd.Flags().GetInt("index")
		titleFlag, _ := cmd.Flags().GetString("title")

		testIdx := resolveProxyIndex(targetArg, idxFlag, titleFlag, proxies)
		if testIdx < 0 {
			if activeIdx >= 0 {
				testIdx = activeIdx
			} else {
				var err error
				testIdx, err = config.SelectProxy(proxies)
				if err != nil {
					log.Error("Failed to select proxy for testing: %v", err)
					return
				}
			}
		}

		targetProxy := proxies[testIdx]
		log.Info("Testing connectivity through '%s' (%s) -> %s...", targetProxy.Title, targetProxy.Value, targetURL)

		duration, err := config.TestProxyConnection(targetProxy.Value, targetURL)
		if err != nil {
			log.Error("Connection test failed after %v: %v", duration.Round(100), err)
			return
		}

		log.Success("Connection successful! Response time: %v", duration.Round(100))
	},
}
