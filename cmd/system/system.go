package system

import (
	"github.com/spf13/cobra"
)

var SystemCmd = &cobra.Command{
	Use:   "system",
	Short: "A command for managing the system",
	Long:  `This command will help manage various aspects of MacOS and Linux systems.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		cobra.CheckErr(err)
	},
}

func init() {
	SystemCmd.AddCommand(KillPortCmd)
	SystemCmd.AddCommand(FindNonMovieFoldersCmd)
	SystemCmd.AddCommand(UpdateSystemCmd)
	SystemCmd.AddCommand(ProxyCmd)

	// Add flags for subcommands if needed
	FindNonMovieFoldersCmd.Flags().Bool("dry-run", false, "Only print the directories that do not contain movie files")
}
