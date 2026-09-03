package system

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

const (
	msgUpdatedProxyConfigurations = "Updated proxy configurations:"
	msgFailedEnableProxyFmt       = "Failed to enable proxy: %v"
)

var ProxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Show or configure system proxies",
	Long:  `Display, switch, test, and manage multiple proxy configurations with rich visual feedback.`,
	Run: func(cmd *cobra.Command, args []string) {
		listProxyConfigurations(cmd)
	},
}

var statusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"ls", "list"},
	Short:   "Show active proxy status, profiles, and shell env vars",
	Long:    `Displays current active proxy status, list of stored configurations, and environment variable states.`,
	Run: func(cmd *cobra.Command, args []string) {
		listProxyConfigurations(cmd)
	},
}

// Common function to list proxy configurations.
func listProxyConfigurations(cmd *cobra.Command) {
	var compact bool
	var showEnv bool
	var showLowercaseEnv bool
	if cmd != nil {
		compact, _ = cmd.Flags().GetBool("compact")
		showEnv, _ = cmd.Flags().GetBool("env")
		showLowercaseEnv, _ = cmd.Flags().GetBool("lowercase-env")
	} else {
		compact = false
		showEnv = false
		showLowercaseEnv = false
	}
	proxies, activeIndex := config.GetProxyConfigs()

	renderProxyList(compact, proxies)
	if showEnv {
		renderEnv(compact, showLowercaseEnv)
	}
	renderActive(compact, proxies, activeIndex)
	renderNote(compact)
}

func renderProxyList(compact bool, proxies []config.ProxyConfig) {
	header := theme.PrimaryText.Bold(true).Render("🌐 Proxy Configurations (★ active, • inactive):")
	fmt.Println(header)
	if len(proxies) == 0 {
		fmt.Println(
			theme.MutedText.Render("  No proxy configurations found. Use 'eng system proxy add' to create one."),
		)
		return
	}
	for i, p := range proxies {
		prefix := theme.MutedText.Render(fmt.Sprintf("%d.", i+1))
		if compact {
			prefix = theme.MutedText.Render("•")
		}
		fmt.Printf("  %s %s\n", prefix, config.FormatProxyOption(p))
	}
}

func renderEnv(compact, showLowercase bool) {
	if !compact {
		fmt.Println(theme.PrimaryText.Bold(true).Render("\nSystem environment variables:"))
		printEnvVar("ALL_PROXY", os.Getenv("ALL_PROXY"))
		printEnvVar("HTTP_PROXY", os.Getenv("HTTP_PROXY"))
		printEnvVar("HTTPS_PROXY", os.Getenv("HTTPS_PROXY"))
		printEnvVar("GLOBAL_AGENT_HTTP_PROXY", os.Getenv("GLOBAL_AGENT_HTTP_PROXY"))
		printEnvVar("NO_PROXY", os.Getenv("NO_PROXY"))

		fmt.Println(theme.PrimaryText.Bold(true).Render("\nLowercase environment variables:"))
		printEnvVar("http_proxy", os.Getenv("http_proxy"))
		printEnvVar("https_proxy", os.Getenv("https_proxy"))
		printEnvVar("no_proxy", os.Getenv("no_proxy"))
		return
	}

	all := os.Getenv("ALL_PROXY")
	http := os.Getenv("HTTP_PROXY")
	https := os.Getenv("HTTPS_PROXY")
	global := os.Getenv("GLOBAL_AGENT_HTTP_PROXY")
	noProxy := os.Getenv("NO_PROXY")

	same := all == http && http == https && https == global
	prefix := theme.MutedText.Render("Env:")
	if same {
		fmt.Printf(
			"%s ALL/HTTP/HTTPS/GLOBAL=%s, NO_PROXY=%s\n",
			prefix,
			theme.SuccessText.Render(all),
			theme.SuccessText.Render(noProxy),
		)
	} else {
		fmt.Printf(
			"%s ALL=%s HTTP=%s HTTPS=%s GLOBAL=%s NO_PROXY=%s\n",
			prefix,
			theme.SuccessText.Render(all),
			theme.SuccessText.Render(http),
			theme.SuccessText.Render(https),
			theme.SuccessText.Render(global),
			theme.SuccessText.Render(noProxy),
		)
	}
	if showLowercase {
		lhttp := os.Getenv("http_proxy")
		lhttps := os.Getenv("https_proxy")
		lno := os.Getenv("no_proxy")
		fmt.Printf(
			"%s http=%s https=%s no=%s\n",
			theme.MutedText.Render("Env (lowercase):"),
			theme.SuccessText.Render(lhttp),
			theme.SuccessText.Render(lhttps),
			theme.SuccessText.Render(lno),
		)
	}
}

func printEnvVar(key, val string) {
	if val == "" {
		val = theme.MutedText.Render("<empty>")
	} else {
		val = theme.SuccessText.Render(val)
	}
	fmt.Printf("  %s %s\n", theme.MutedText.Render(key+":"), val)
}

func renderActive(compact bool, proxies []config.ProxyConfig, activeIndex int) {
	if activeIndex >= 0 && activeIndex < len(proxies) {
		activeStr := theme.SuccessText.Bold(true).Render(config.FormatProxyOption(proxies[activeIndex]))
		if compact {
			fmt.Printf("\n%s %s\n", theme.PrimaryText.Render("Active:"), activeStr)
		} else {
			fmt.Printf("\n%s %s\n", theme.PrimaryText.Render("Active proxy:"), activeStr)
		}
		return
	}

	noneStr := theme.MutedText.Render("none")
	if compact {
		fmt.Printf("\n%s %s\n", theme.PrimaryText.Render("Active:"), noneStr)
	} else {
		fmt.Println(theme.MutedText.Render("\nNo active proxy configured."))
	}
}

func renderNote(compact bool) {
	if compact {
		return
	}
	fmt.Println()
	theme.InfoMessage("Environment variable changes only affect the current process.")
	fmt.Println(
		theme.MutedText.Render(
			"  For system-wide changes, apply environment variables to your current shell using:",
		),
	)
	fmt.Println(theme.SuccessText.Bold(true).Render("    eval $(eng system proxy export)"))
}

func resolveProxyIndex(target string, idxFlag int, titleFlag string, proxies []config.ProxyConfig) int {
	t := strings.TrimSpace(target)
	if t != "" {
		if idx, err := strconv.Atoi(t); err == nil {
			if idx > 0 && idx <= len(proxies) {
				return idx - 1
			}
			if idx == 0 && len(proxies) > 0 {
				return 0
			}
		}
		matched := config.FindProxyIndexByTitle(proxies, t)
		if matched >= 0 {
			return matched
		}
	}
	if idxFlag >= 0 && idxFlag < len(proxies) {
		return idxFlag
	}
	if titleFlag != "" {
		return config.FindProxyIndexByTitle(proxies, titleFlag)
	}
	return -1
}

// completeProxyNames offers configured proxy titles for positional args.
// GetProxyConfigs logs to stdout, which would corrupt shell completion
// output, so log writers are silenced while completing and restored after.
func completeProxyNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	prevOut, prevErr := log.Out, log.Err
	log.SetWriters(io.Discard, io.Discard)
	defer log.SetWriters(prevOut, prevErr)

	proxies, _ := config.GetProxyConfigs()
	var names []string
	for _, p := range proxies {
		if strings.HasPrefix(p.Title, toComplete) {
			names = append(names, p.Title)
		}
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

// completeProxyTitles offers configured proxy titles for --title flags.
func completeProxyTitles(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeProxyNames(cmd, args, toComplete)
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

		fmt.Println(msgUpdatedProxyConfigurations)
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
		proxies, activeIndex := config.GetProxyConfigs()

		if activeIndex >= 0 && activeIndex < len(proxies) {
			proxyValue := proxies[activeIndex].Value
			fmt.Printf("export ALL_PROXY='%s'\n", proxyValue)
			fmt.Printf("export HTTP_PROXY='%s'\n", proxyValue)
			fmt.Printf("export HTTPS_PROXY='%s'\n", proxyValue)
			fmt.Printf("export GLOBAL_AGENT_HTTP_PROXY='%s'\n", proxyValue)
			fmt.Printf("export http_proxy='%s'\n", proxyValue)
			fmt.Printf("export https_proxy='%s'\n", proxyValue)

			noProxyValue := "localhost,127.0.0.1,::1,.local"
			if proxies[activeIndex].NoProxy != "" {
				noProxyValue = noProxyValue + "," + proxies[activeIndex].NoProxy
			}

			fmt.Printf("export NO_PROXY='%s'\n", noProxyValue)
			fmt.Printf("export no_proxy='%s'\n", noProxyValue)
		} else {
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
			fmt.Println(msgUpdatedProxyConfigurations)
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

		fmt.Println(msgUpdatedProxyConfigurations)
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

var (
	enableCmd  = useCmd
	disableCmd = offCmd
	setCmd     = editCmd
)

func init() {
	// Subcommands
	ProxyCmd.AddCommand(statusCmd)
	ProxyCmd.AddCommand(useCmd)
	ProxyCmd.AddCommand(offCmd)
	ProxyCmd.AddCommand(addCmd)
	ProxyCmd.AddCommand(editCmd)
	ProxyCmd.AddCommand(removeCmd)
	ProxyCmd.AddCommand(exportCmd)
	ProxyCmd.AddCommand(toggleCmd)
	ProxyCmd.AddCommand(testCmd)

	// Dynamic completion for proxy names (positional args and --title flags).
	// GetProxyConfigs logs to stdout, which would corrupt `eng __complete`
	// output, so silence log writers while completing.
	for _, c := range []*cobra.Command{useCmd, editCmd, removeCmd, testCmd} {
		c.ValidArgsFunction = completeProxyNames
	}

	// Persistent flags to control listing style
	ProxyCmd.PersistentFlags().Bool("compact", true, "Show compact status output")
	ProxyCmd.PersistentFlags().Bool("env", false, "Include environment variables in status output")
	ProxyCmd.PersistentFlags().Bool("lowercase-env", false, "Include lowercase environment vars in compact mode")

	// Flags for add & edit
	addCmd.Flags().String("title", "", "Proxy configuration title")
	addCmd.Flags().String("url", "", "Proxy address (e.g., http://host:port)")
	addCmd.Flags().String("value", "", "Alias for --url")
	addCmd.Flags().String("no-proxy", "", "Additional no_proxy values (comma-separated)")
	addCmd.Flags().Bool("enable", false, "Enable proxy after adding")

	editCmd.Flags().String("title", "", "Proxy configuration title")
	editCmd.Flags().String("url", "", "Proxy address (e.g., http://host:port)")
	editCmd.Flags().String("value", "", "Alias for --url")
	editCmd.Flags().String("no-proxy", "", "Additional no_proxy values (comma-separated)")
	editCmd.Flags().Bool("enable", false, "Enable this proxy after editing")
	editCmd.Flags().Bool("interactive", false, "Use interactive prompts when missing values")

	// Flags for use/enable
	useCmd.Flags().Int("index", -1, "Proxy index to enable")
	useCmd.Flags().String("title", "", "Proxy title to enable")
	useCmd.Flags().Bool("quiet", false, "Suppress status output after enabling")

	// Flags for off/disable
	offCmd.Flags().Bool("quiet", false, "Suppress status output after disabling")

	// Flags for remove
	removeCmd.Flags().Int("index", -1, "Proxy index to remove")
	removeCmd.Flags().String("title", "", "Proxy title to remove")

	// Flags for test
	testCmd.Flags().String("target", "https://1.1.1.1", "Target URL to test proxy against")
	testCmd.Flags().Int("index", -1, "Proxy index to test")
	testCmd.Flags().String("title", "", "Proxy title to test")

	// Flags for toggle
	toggleCmd.Flags().Bool("on", false, "Toggle on (enable a proxy)")
	toggleCmd.Flags().Bool("off", false, "Toggle off (disable all proxies)")
	toggleCmd.Flags().Bool("quiet", false, "Suppress status output after toggling")
	toggleCmd.Flags().Int("index", -1, "Enable proxy by index")
	toggleCmd.Flags().String("title", "", "Enable proxy by title")

	// Flag completion must be registered after the flags exist, otherwise
	// cobra rejects it with "flag does not exist".
	for _, c := range []*cobra.Command{useCmd, editCmd, removeCmd, testCmd, toggleCmd} {
		_ = c.RegisterFlagCompletionFunc("title", completeProxyTitles)
	}
}
