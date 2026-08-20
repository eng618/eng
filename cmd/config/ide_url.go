package config

import (
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
)

// IdeURLCmd defines the command for setting the Antigravity IDE download URL.
var IdeURLCmd = &cobra.Command{
	Use:     "ide-url [url]",
	Aliases: []string{"antigravity-url", "antigravity-ide-url", "ide-download-url"},
	Short:   "Update Antigravity IDE download URL",
	Long:    `This command configures the direct download URL for Antigravity IDE updates in $HOME/.eng.yaml.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Checking Antigravity IDE download URL in config file...")
		if len(args) > 0 {
			config.IdeURL(args[0])
		} else {
			config.IdeURL()
		}
	},
}
