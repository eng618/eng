// Package git provides cobra commands for managing multiple git repositories
// in a development folder.
package git

import (
	"fmt"
	"os"
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

// GitCommandSetup holds common variables and flags resolved during git command initialization.
type GitCommandSetup struct {
	IsVerbose bool
	DryRun    bool
	DevPath   string
}

// setupGitCommand abstracts the duplicated setup and validation steps shared across git commands.
func setupGitCommand(cmd *cobra.Command) (*GitCommandSetup, error) {
	isVerbose := cmdutil.IsVerbose(cmd)
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	devPath, err := getWorkingPath(cmd)
	if err != nil {
		return nil, err
	}

	log.Verbose(isVerbose, "Development path: %s", devPath)

	if dryRun {
		log.Info("Dry run mode - no actual git operations will be performed")
	}

	return &GitCommandSetup{
		IsVerbose: isVerbose,
		DryRun:    dryRun,
		DevPath:   devPath,
	}, nil
}

// printHeader prints a stylized header consistently to log.Out if progress display is enabled.
func printHeader(text string) {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary).
		MarginBottom(1)
	if !ui.DisableProgress {
		fmt.Fprintln(log.Out, headerStyle.Render(text))
	}
}

// GitCmd serves as the base command for all git repository management operations.
// It doesn't perform any action itself but groups subcommands like sync-all, status-all, etc.
var GitCmd = &cobra.Command{
	Use:   "git",
	Short: "Manage multiple git repositories",
	Long:  `This command is used to facilitate the management of multiple git repositories in your development folder.`,
	Run: func(cmd *cobra.Command, args []string) {
		showInfo, _ := cmd.Flags().GetBool("info")
		isVerbose := cmdutil.IsVerbose(cmd)

		if showInfo {
			log.Info("Current git repository management configuration:")
			gitCfg := config.GetGitConfig()
			devPath := gitCfg.DevPath

			if devPath == "" {
				log.Warn("  Development Path: Not Set")
				log.Info("  Use 'eng config git-dev-path' to set your development folder path")
			} else {
				log.Info("  Development Path: %s", devPath)
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
	GitCmd.PersistentFlags().
		BoolP("current", "c", false, "Use current working directory instead of configured development path")

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

// getBoolFlag safely checks if a flag exists anywhere in the command's local,
// persistent, or inherited flag sets and returns its boolean value.
func getBoolFlag(cmd *cobra.Command, name string) bool {
	if f := cmd.Flag(name); f != nil {
		if val, err := strconv.ParseBool(f.Value.String()); err == nil {
			return val
		}
	}
	return false
}

// getWorkingPath returns either the current working directory (if --current flag is used)
// or the configured development path from the config file.
func getWorkingPath(cmd *cobra.Command) (string, error) {
	useCurrent := getBoolFlag(cmd, "current")

	if useCurrent {
		devPath, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current working directory: %w", err)
		}
		log.Info("Using current directory: %s", devPath)
		return devPath, nil
	}

	gitCfg := config.GetGitConfig()
	devPath := gitCfg.DevPath
	if devPath == "" {
		return "", fmt.Errorf(
			"development folder path is not set. Run 'eng config git-dev-path ~/Development' (or 'eng config edit --interactive'), use the --current flag, or run 'eng doctor' to diagnose",
		)
	}

	// Expand environment variables
	devPath = os.ExpandEnv(devPath)
	return devPath, nil
}
