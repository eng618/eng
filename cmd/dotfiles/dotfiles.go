package dotfiles

import (
	"github.com/eng618/eng/utils/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var DotfilesCmd = &cobra.Command{
	Use:   "dotfiles",
	Short: "Manage dotfiles",
	Long:  `This command is used to facilitate the management of private hidden dot files.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("dotfiles called")

		isVerbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			log.Error("Failed to parse verbose flag: %s", err)
			return
		}

		log.Verbose(isVerbose, "Configuration values:")
		log.Verbose(isVerbose, "dotfiles.repoPath: %s", viper.GetString("dotfiles.repoPath"))
		log.Verbose(isVerbose, "dotfiles.worktree: %s", viper.GetString("dotfiles.worktree"))
	},
}

func init() {
	DotfilesCmd.AddCommand(SyncCmd)
	DotfilesCmd.AddCommand(FetchCmd)
}
