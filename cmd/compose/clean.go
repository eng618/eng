package compose

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/cleanup"
	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

var (
	allFlagClean        bool
	olderThanFlagClean  string
	buildCacheFlagClean bool
	volumesFlagClean    bool
	dryRunFlagClean     bool
	yesFlagClean        bool
)

var cleanCmd = &cobra.Command{
	Use:     "clean",
	Aliases: []string{"prune", "cleanup", "c"},
	Short:   "Prune unused and dangling Docker images, build cache, and volumes",
	Long: `Prune dangling layers, unused images older than a specified duration (default: 7 days / 168h), BuildKit build cache, and optionally unused volumes to reclaim host storage.

Use the --dry-run (-n) flag to preview reclaimable space without removing data.
Use the --all (-a) flag to remove all unused images regardless of age.
Use the --volumes (-v) flag to also prune unused anonymous volumes.`,
	RunE: func(cmd *cobra.Command, _args []string) error {
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary).
			MarginBottom(1)
		if !ui.DisableProgress {
			fmt.Fprintln(log.Out, headerStyle.Render("🧹 Docker & Compose Storage Prune"))
		}

		isVerbose := cmdutil.IsVerbose(cmd)

		if !yesFlagClean && !dryRunFlagClean {
			confirm, err := ui.Confirm("Are you sure you want to prune unused Docker resources?", true)
			if err != nil || !confirm {
				log.Message("Docker cleanup cancelled.")
				return nil
			}
		}

		opts := cleanup.DockerCleanOptions{
			All:        allFlagClean,
			OlderThan:  olderThanFlagClean,
			BuildCache: buildCacheFlagClean,
			Volumes:    volumesFlagClean,
			DryRun:     dryRunFlagClean,
			Verbose:    isVerbose,
		}

		report, err := cleanup.RunDockerCleanup(opts)
		if err != nil {
			return err
		}

		if report != nil {
			summary := report.RenderSummaryTable()
			if summary != "" {
				fmt.Println(summary)
			}
		}

		return nil
	},
}

func init() {
	cleanCmd.Flags().BoolVarP(&allFlagClean, "all", "a", false, "Prune all unused images regardless of age")
	cleanCmd.Flags().StringVarP(&olderThanFlagClean, "older-than", "o", "168h", "Prune unused images older than duration (e.g. 168h, 72h)")
	cleanCmd.Flags().BoolVarP(&buildCacheFlagClean, "build-cache", "b", true, "Prune BuildKit / Docker build cache")
	cleanCmd.Flags().BoolVar(&volumesFlagClean, "volumes", false, "Prune unused Docker anonymous volumes")
	cleanCmd.Flags().BoolVarP(&dryRunFlagClean, "dry-run", "n", false, "Preview reclaimable space without deleting")
	cleanCmd.Flags().BoolVarP(&yesFlagClean, "yes", "y", false, "Skip confirmation prompt")
}
