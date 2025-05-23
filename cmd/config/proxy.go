package config

import (
	"github.com/spf13/cobra"
)

// ProxyConfigCmd represents the command for interactively managing proxy configurations
var ProxyConfigCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Interactively manage proxy configurations",
	Long:  `Launch an interactive terminal UI for managing proxy configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		StartProxyUI()
	},
}
