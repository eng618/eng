package immich

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

// commandFlags holds flag state for one built command tree. Each call to
// NewCommand gets a fresh set, so multiple registrations of the tree
// (e.g. `eng immich` and `eng system immich`) stay fully independent.
type commandFlags struct {
	json        bool
	pager       bool
	retention   int
	restoreFile string
	autoConfirm bool
	logsFollow  bool
	logsTail    int
	logsService string
}

// NewCommand builds the Immich command tree (status, backup, restore,
// start, stop, restart, logs). Callers in cmd/ attach the returned root to
// their own parent command. All runtime behavior lives here against the
// Manager in this package, so no command package needs to import another.
func NewCommand() *cobra.Command {
	f := &commandFlags{logsFollow: true, logsTail: 50}

	root := &cobra.Command{
		Use:     "immich",
		Aliases: []string{"photos"},
		Short:   "Manage Immich photo stack, database, backups, and lifecycle",
		Long: `Inspect Immich service health, monitor container states, query database metrics,
execute verified backups, perform disaster recovery, and manage systemd services.`,
		RunE: func(cmd *cobra.Command, _args []string) error {
			return cmd.Help()
		},
	}

	statusCmd := &cobra.Command{
		Use:     "status",
		Aliases: []string{"ps", "info"},
		Short:   "Display comprehensive health, metrics, and backup status",
		RunE: func(cmd *cobra.Command, _args []string) error {
			ctx := cmd.Context()
			mgr := NewManager("")

			status, err := mgr.GetStatus(ctx)
			if err != nil {
				return fmt.Errorf("failed to probe immich status: %w", err)
			}

			if f.json {
				data, err := json.MarshalIndent(status, "", "  ")
				if err != nil {
					return err
				}
				fmt.Fprintln(log.Out, string(data))
				return nil
			}

			termWidth := ui.GetTerminalWidth()
			rendered := RenderStatus(status, termWidth)

			if f.pager {
				return ui.RunContainersPager(rendered)
			}

			fmt.Fprint(log.Out, rendered)
			return nil
		},
	}

	backupCmd := &cobra.Command{
		Use:     "backup",
		Aliases: []string{"dump"},
		Short:   "Run verified database backup and configuration snapshot",
		RunE: func(cmd *cobra.Command, _args []string) error {
			ctx := cmd.Context()
			mgr := NewManager("")

			headerStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Primary).
				MarginBottom(1)
			if !ui.DisableProgress {
				fmt.Fprintln(log.Out, headerStyle.Render("📦 Running Immich Full Backup"))
			}

			var spinner *ui.Spinner
			if !ui.DisableProgress {
				spinner = ui.NewSpinner("Dumping PostgreSQL, validating archive, and generating checksums...")
				spinner.Start()
			}

			res, err := mgr.RunBackup(ctx, f.retention)
			if spinner != nil {
				spinner.Stop()
			}

			if err != nil {
				return fmt.Errorf("backup failed: %w", err)
			}

			theme.SuccessMessage("Immich backup completed successfully!")
			fmt.Fprintf(log.Out, "\n  %s %s (%s)\n", theme.BoldText.Render("DB Archive:"), res.BackupFile, res.Size)
			fmt.Fprintf(log.Out, "  %s %s\n", theme.BoldText.Render("Checksum:"), res.ChecksumFile)
			fmt.Fprintf(log.Out, "  %s %s\n", theme.BoldText.Render("Config Snapshot:"), res.ConfigArchive)
			fmt.Fprintf(log.Out, "  %s %s\n\n", theme.BoldText.Render("Duration:"), res.Duration.Round(100*1000000))
			return nil
		},
	}

	restoreCmd := &cobra.Command{
		Use:   "restore [backup-file]",
		Short: "Restore Immich database and configuration from backup",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			mgr := NewManager("")

			targetFile := f.restoreFile
			if len(args) > 0 {
				targetFile = args[0]
			}

			return mgr.RunRestore(ctx, targetFile, f.autoConfirm)
		},
	}

	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start Immich service stack via systemd",
		RunE: func(cmd *cobra.Command, _args []string) error {
			ctx := cmd.Context()
			mgr := NewManager("")

			if err := mgr.Start(ctx); err != nil {
				return err
			}
			theme.SuccessMessage("Immich service started successfully.")
			return nil
		},
	}

	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: "Gracefully stop Immich service stack",
		RunE: func(cmd *cobra.Command, _args []string) error {
			ctx := cmd.Context()
			mgr := NewManager("")

			if err := mgr.Stop(ctx); err != nil {
				return err
			}
			theme.SuccessMessage("Immich service stopped gracefully.")
			return nil
		},
	}

	restartCmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart Immich service stack via systemd",
		RunE: func(cmd *cobra.Command, _args []string) error {
			ctx := cmd.Context()
			mgr := NewManager("")

			if err := mgr.Restart(ctx); err != nil {
				return err
			}
			theme.SuccessMessage("Immich service restarted successfully.")
			return nil
		},
	}

	logsCmd := &cobra.Command{
		Use:     "logs",
		Aliases: []string{"log"},
		Short:   "View live Immich service or container logs",
		RunE: func(cmd *cobra.Command, _args []string) error {
			ctx := cmd.Context()
			mgr := NewManager("")

			log.Verbose(false, "Streaming Immich logs...")
			return mgr.Logs(ctx, f.logsService, f.logsFollow, f.logsTail)
		},
	}

	statusCmd.Flags().BoolVar(&f.json, "json", false, "Output status in JSON format")
	statusCmd.Flags().BoolVarP(&f.pager, "pager", "p", false, "Open status inside scrollable viewport")

	backupCmd.Flags().IntVarP(&f.retention, "retention", "r", 0, "Override retention period (days)")

	restoreCmd.Flags().StringVarP(&f.restoreFile, "file", "f", "", "Specific backup archive path to restore")
	restoreCmd.Flags().BoolVarP(&f.autoConfirm, "yes", "y", false, "Auto-confirm restoration prompt")

	logsCmd.Flags().BoolVarP(&f.logsFollow, "follow", "f", true, "Follow log stream")
	logsCmd.Flags().IntVarP(&f.logsTail, "tail", "n", 50, "Number of lines to show from end of logs")
	logsCmd.Flags().
		StringVarP(&f.logsService, "service", "s", "", "Filter to specific container (e.g. server, database, ml, redis)")

	root.AddCommand(statusCmd)
	root.AddCommand(backupCmd)
	root.AddCommand(restoreCmd)
	root.AddCommand(startCmd)
	root.AddCommand(stopCmd)
	root.AddCommand(restartCmd)
	root.AddCommand(logsCmd)
	return root
}
