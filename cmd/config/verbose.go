package config

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// VerboseCmd defines the command for managing the verbose output setting.
var VerboseCmd = &cobra.Command{
	Use:   "verbose",
	Short: "Show or set the current verbose flag value",
	Long:  `This command checks, enables, or disables verbose output for this CLI session and saves it to the config file.`,
	Run: func(cmd *cobra.Command, args []string) {
		enable, _ := cmd.Flags().GetBool("enable")
		disable, _ := cmd.Flags().GetBool("disable")

		if enable && disable {
			fmt.Println("Cannot enable and disable verbose mode at the same time.")
			return
		}
		if enable {
			viper.Set("verbose", true)
			_ = viper.WriteConfig()
			fmt.Println("Verbose mode enabled and saved to config.")
			return
		}
		if disable {
			viper.Set("verbose", false)
			_ = viper.WriteConfig()
			fmt.Println("Verbose mode disabled and saved to config.")
			return
		}

		verbose := viper.GetBool("verbose")
		if verbose {
			fmt.Println("Verbose mode is enabled.")
			fmt.Println("To disable: eng config verbose --disable")
		} else {
			fmt.Println("Verbose mode is disabled.")
			fmt.Println("To enable: eng config verbose --enable")
		}
	},
}

func init() {
	VerboseCmd.Flags().Bool("enable", false, "Enable verbose mode and save to config")
	VerboseCmd.Flags().Bool("disable", false, "Disable verbose mode and save to config")
}
