package files

import (
	"github.com/spf13/cobra"
)

var FilesCmd = &cobra.Command{
	Use:   "files",
	Short: "A command for managing files",
	Long:  `This command will help manage various aspects of file operations on MacOS and Linux systems.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		cobra.CheckErr(err)
	},
}

func init() {
	FilesCmd.AddCommand(FindAndDeleteCmd)

	FindAndDeleteCmd.Flags().StringVarP(&globPattern, "glob", "g", "", "Glob pattern to match files (e.g., '*.bak'). Bypasses extension selection.")
	FindAndDeleteCmd.Flags().StringVarP(&extension, "ext", "e", "", "File extension to match (e.g., '.json'). Bypasses extension selection.")
	FindAndDeleteCmd.Flags().BoolVarP(&listExtensions, "list-extensions", "l", false, "List all file extensions in the directory")
}
