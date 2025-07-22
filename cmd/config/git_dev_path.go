package config

import (
	"github.com/eng618/eng/utils/config"
	"github.com/eng618/eng/utils/log"
	"github.com/spf13/cobra"
)

// GitDevPathCmd defines the command for setting the development folder path.
var GitDevPathCmd = &cobra.Command{
	Use:   "git-dev-path",
	Short: "Set development folder path",
	Long:  `This command sets the path to the development folder containing git repositories in the config file.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Checking for development folder path in config file...")
		config.GitDevPath()
	},
}
