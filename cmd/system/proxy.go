package system

import (
	"io"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
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
	names := make([]string, len(proxies))
	for i, p := range proxies {
		names[i] = p.Title
	}
	return cmdutil.CompletePrefix(names, toComplete)
}

// completeProxyTitles offers configured proxy titles for --title flags.
func completeProxyTitles(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeProxyNames(cmd, args, toComplete)
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
