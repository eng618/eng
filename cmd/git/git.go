// Package git provides cobra commands for managing multiple git repositories
// in a development folder.
package git

import (
	"fmt"
	"os"

	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// GitCmd serves as the base command for all git repository management operations.
// It doesn't perform any action itself but groups subcommands like sync-all, status-all, etc.
var GitCmd = &cobra.Command{
	Use:   "git",
	Short: "Manage multiple git repositories",
	Long:  `This command is used to facilitate the management of multiple git repositories in your development folder.`,
	Run: func(cmd *cobra.Command, args []string) {
		showInfo, _ := cmd.Flags().GetBool("info")
		isVerbose := utils.IsVerbose(cmd)

		if showInfo {
			log.Info("Current git repository management configuration:")
			devPath := viper.GetString("git.devPath")

			if devPath == "" {
				log.Warn("  Development Path (git.devPath): Not Set")
				log.Info("  Use 'eng config git-dev-path' to set your development folder path")
			} else {
				log.Info("  Development Path (git.devPath): %s", devPath)
			}
			return // Don't show help if info flag is used
		}

		// If no subcommand is given, print the help information.
		if len(args) == 0 {
			log.Verbose(isVerbose, "No subcommand provided, showing help.")
			err := cmd.Help()
			cobra.CheckErr(err)
		} else {
			log.Verbose(isVerbose, "Subcommand '%s' provided.", args[0])
		}
	},
}

func init() {
	GitCmd.Flags().BoolP("info", "i", false, "Show current git repository management configuration")
	GitCmd.PersistentFlags().BoolP("current", "c", false, "Use current working directory instead of configured development path")

	GitCmd.AddCommand(SyncAllCmd)
	GitCmd.AddCommand(StatusAllCmd)
	GitCmd.AddCommand(FetchAllCmd)
	GitCmd.AddCommand(ListCmd)
	GitCmd.AddCommand(PullAllCmd)
	GitCmd.AddCommand(PushAllCmd)
	GitCmd.AddCommand(BranchAllCmd)
	GitCmd.AddCommand(StashAllCmd)
	GitCmd.AddCommand(CleanAllCmd)
}

// getWorkingPath returns either the current working directory (if --current flag is used)
// or the configured development path from the config file.
func getWorkingPath(cmd *cobra.Command) (string, error) {
	// Check for the persistent flag on the root git command or inherited
	useCurrent, _ := cmd.PersistentFlags().GetBool("current")
	if !useCurrent {
		// Try to get from local flags if not found in persistent flags
		useCurrent, _ = cmd.Flags().GetBool("current")
	}

	if useCurrent {
		devPath, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current working directory: %w", err)
		}
		log.Info("Using current directory: %s", devPath)
		return devPath, nil
	}

	devPath := viper.GetString("git.devPath")
	if devPath == "" {
		return "", fmt.Errorf("git.devPath is not set in the configuration file. Use 'eng config git-dev-path' to set your development folder path, or use --current flag")
	}

	// Expand environment variables
	devPath = os.ExpandEnv(devPath)
	return devPath, nil
}
