package config

import (
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/utils/config"
	"github.com/eng618/eng/internal/utils/log"
)

// GitDevPathCmd defines the command for setting the development folder path.
var GitDevPathCmd = &cobra.Command{
	Use:   "git-dev-path",
	Short: "Update config git dev path",
	Long:  `This command sets the path to the development folder containing git repositories in the config file.`,
	Run: func(cmd *cobra.Command, _args []string) {
		log.Start("Checking for development folder path in config file...")
		config.GitDevPath()
	},
}
