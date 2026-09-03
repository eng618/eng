package system

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
	dockerFlagCleanSys    bool
	allImagesFlagCleanSys bool
	olderThanFlagCleanSys string
	journalFlagCleanSys   bool
	journalSizeCleanSys   string
	packagesFlagCleanSys  bool
	asdfFlagCleanSys      bool
	brewFlagCleanSys      bool
	dryRunFlagCleanSys    bool
	yesFlagCleanSys       bool
	timeoutFlagCleanSys   int
)

// CleanCmd represents the system clean command.
var CleanCmd = &cobra.Command{
	Use:     "clean",
	Aliases: []string{"cleanup", "prune", "c"},
	Short:   "Clean and reclaim host system storage across OS and developer tools",
	Long: `Orchestrates cross-platform system maintenance across macOS, Ubuntu, Fedora, and Debian/Raspberry Pi.

Cleans package manager caches (APT/DNF/Brew), vacuums systemd journal logs, prunes Docker container and image layers, and cleans outdated asdf tool versions.

Use the --dry-run (-n) flag to preview operations without modifying system data.
Use the --yes (-y) flag to run all available cleanup operations without prompting.`,
	RunE: func(cmd *cobra.Command, _args []string) error {
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary).
			MarginBottom(1)
		if !ui.DisableProgress {
			fmt.Fprintln(log.Out, headerStyle.Render("🧹 Host & System Storage Maintenance"))
		}

		isVerbose := cmdutil.IsVerbose(cmd)

		opts := cleanup.SystemCleanOptions{
			Docker: dockerFlagCleanSys,
			DockerOpts: cleanup.DockerCleanOptions{
				All:        allImagesFlagCleanSys,
				OlderThan:  olderThanFlagCleanSys,
				BuildCache: true,
				Volumes:    false,
				DryRun:     dryRunFlagCleanSys,
				Verbose:    isVerbose,
			},
			Journal:        journalFlagCleanSys,
			JournalSize:    journalSizeCleanSys,
			Packages:       packagesFlagCleanSys,
			Asdf:           asdfFlagCleanSys,
			Brew:           brewFlagCleanSys,
			DryRun:         dryRunFlagCleanSys,
			Verbose:        isVerbose,
			AutoApprove:    yesFlagCleanSys,
			CleanupTimeout: timeoutFlagCleanSys,
		}

		report, err := cleanup.RunSystemCleanup(opts)
		if err != nil {
			return err
		}

		if report != nil {
			summary := report.RenderSummaryTable()
			if summary != "" {
				fmt.Fprintln(log.Out, summary)
			}
		}

		return nil
	},
}

func init() {
	CleanCmd.Flags().BoolVar(&dockerFlagCleanSys, "docker", true, "Include Docker container & image cleanup")
	CleanCmd.Flags().BoolVar(&allImagesFlagCleanSys, "all-images", false, "Prune all unused Docker images regardless of age")
	CleanCmd.Flags().StringVar(&olderThanFlagCleanSys, "older-than", "168h", "Filter age for unused Docker images (e.g. 168h)")
	CleanCmd.Flags().BoolVar(&journalFlagCleanSys, "journal", true, "Vacuum systemd journal logs")
	CleanCmd.Flags().StringVar(&journalSizeCleanSys, "journal-size", "500M", "Target size for systemd journal vacuuming")
	CleanCmd.Flags().BoolVar(&packagesFlagCleanSys, "packages", true, "Clean OS package manager caches (APT/DNF)")
	CleanCmd.Flags().BoolVar(&asdfFlagCleanSys, "asdf", true, "Clean outdated asdf tool versions")
	CleanCmd.Flags().BoolVar(&brewFlagCleanSys, "brew", true, "Clean Homebrew caches")
	CleanCmd.Flags().BoolVarP(&dryRunFlagCleanSys, "dry-run", "n", false, "Preview operations without modifying the system")
	CleanCmd.Flags().BoolVarP(&yesFlagCleanSys, "yes", "y", false, "Auto-approve all cleanup operations without prompting")
	CleanCmd.Flags().IntVar(&timeoutFlagCleanSys, "cleanup-timeout", 60, "Timeout in seconds for interactive prompt")
}
