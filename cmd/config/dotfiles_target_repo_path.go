package config

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
)

// DotfilesTargetRepoPathCmd represents the command to manage the target repository path.
var DotfilesTargetRepoPathCmd = &cobra.Command{
	Use:   "dotfiles-target-repo-path [path]",
	Short: "Show or set the dotfiles target repo path for copy-changes",
	Long: `Show or set the explicit destination git repository used by 'eng dotfiles copy-changes'.

With a path argument, validates that it exists and saves it, so future copies
no longer depend on dev-folder heuristics. Without arguments, prints the
currently effective path and where it came from.`,
	Example: `  eng config dotfiles-target-repo-path
  eng config dotfiles-target-repo-path ~/Development/dotfiles`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			if err := config.SetTargetRepoPath(args[0]); err != nil {
				log.Error("%s", err)
				return
			}
			log.Success("Target repository path set successfully")
			return
		}
		if explicit := config.GetDotfilesConfig().TargetRepoPath; explicit != "" {
			log.Info("Target repository path (explicit config): %s", os.ExpandEnv(explicit))
			return
		}
		log.Info(
			"No explicit target set (heuristic would resolve: %s). Pass a path to pin it.",
			config.TargetRepoPath(config.GetGitConfig().DevPath),
		)
	},
}
