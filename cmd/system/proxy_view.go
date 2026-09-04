package system

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui/theme"
)

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
	fmt.Fprintln(log.Out, header)
	if len(proxies) == 0 {
		fmt.Fprintln(
			log.Out,
			theme.MutedText.Render("  No proxy configurations found. Use 'eng system proxy add' to create one."),
		)
		return
	}
	for i, p := range proxies {
		prefix := theme.MutedText.Render(fmt.Sprintf("%d.", i+1))
		if compact {
			prefix = theme.MutedText.Render("•")
		}
		fmt.Fprintf(log.Out, "  %s %s\n", prefix, config.FormatProxyOption(p))
	}
}

func renderEnv(compact, showLowercase bool) {
	if !compact {
		fmt.Fprintln(log.Out, theme.PrimaryText.Bold(true).Render("\nSystem environment variables:"))
		printEnvVar("ALL_PROXY", os.Getenv("ALL_PROXY"))
		printEnvVar("HTTP_PROXY", os.Getenv("HTTP_PROXY"))
		printEnvVar("HTTPS_PROXY", os.Getenv("HTTPS_PROXY"))
		printEnvVar("GLOBAL_AGENT_HTTP_PROXY", os.Getenv("GLOBAL_AGENT_HTTP_PROXY"))
		printEnvVar("NO_PROXY", os.Getenv("NO_PROXY"))

		fmt.Fprintln(log.Out, theme.PrimaryText.Bold(true).Render("\nLowercase environment variables:"))
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
		fmt.Fprintf(
			log.Out,
			"%s ALL/HTTP/HTTPS/GLOBAL=%s, NO_PROXY=%s\n",
			prefix,
			theme.SuccessText.Render(all),
			theme.SuccessText.Render(noProxy),
		)
	} else {
		fmt.Fprintf(
			log.Out,
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
		fmt.Fprintf(
			log.Out,
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
	fmt.Fprintf(log.Out, "  %s %s\n", theme.MutedText.Render(key+":"), val)
}

func renderActive(compact bool, proxies []config.ProxyConfig, activeIndex int) {
	if activeIndex >= 0 && activeIndex < len(proxies) {
		activeStr := theme.SuccessText.Bold(true).Render(config.FormatProxyOption(proxies[activeIndex]))
		if compact {
			fmt.Fprintf(log.Out, "\n%s %s\n", theme.PrimaryText.Render("Active:"), activeStr)
		} else {
			fmt.Fprintf(log.Out, "\n%s %s\n", theme.PrimaryText.Render("Active proxy:"), activeStr)
		}
		return
	}

	noneStr := theme.MutedText.Render("none")
	if compact {
		fmt.Fprintf(log.Out, "\n%s %s\n", theme.PrimaryText.Render("Active:"), noneStr)
	} else {
		fmt.Fprintln(log.Out, theme.MutedText.Render("\nNo active proxy configured."))
	}
}

func renderNote(compact bool) {
	if compact {
		return
	}
	fmt.Fprintln(log.Out)
	theme.InfoMessage("Environment variable changes only affect the current process.")
	fmt.Fprintln(
		log.Out,
		theme.MutedText.Render(
			"  For system-wide changes, apply environment variables to your current shell using:",
		),
	)
	fmt.Fprintln(log.Out, theme.SuccessText.Bold(true).Render("    eval $(eng system proxy export)"))
}
