package system

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/immich"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

var (
	immichJSONFlag    bool
	immichPagerFlag   bool
	immichRetention   int
	immichRestoreFile string
	immichAutoConfirm bool
	immichLogsFollow  bool
	immichLogsTail    int
	immichLogsService string
)

// ImmichCmd represents the Immich stack management command.
var ImmichCmd = &cobra.Command{
	Use:     "immich",
	Aliases: []string{"photos"},
	Short:   "Manage Immich photo stack, database, backups, and lifecycle",
	Long: `Inspect Immich service health, monitor container states, query database metrics,
execute verified backups, perform disaster recovery, and manage systemd services.`,
	RunE: func(cmd *cobra.Command, _args []string) error {
		return cmd.Help()
	},
}

var immichStatusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"ps", "info"},
	Short:   "Display comprehensive health, metrics, and backup status",
	RunE: func(cmd *cobra.Command, _args []string) error {
		ctx := cmd.Context()
		mgr := immich.NewManager("")

		status, err := mgr.GetStatus(ctx)
		if err != nil {
			return fmt.Errorf("failed to probe immich status: %w", err)
		}

		if immichJSONFlag {
			data, err := json.MarshalIndent(status, "", "  ")
			if err != nil {
				return err
			}
			fmt.Fprintln(log.Out, string(data))
			return nil
		}

		termWidth := ui.GetTerminalWidth()
		rendered := immich.RenderStatus(status, termWidth)

		if immichPagerFlag {
			return ui.RunContainersPager(rendered)
		}

		fmt.Fprint(log.Out, rendered)
		return nil
	},
}

var immichBackupCmd = &cobra.Command{
	Use:     "backup",
	Aliases: []string{"dump"},
	Short:   "Run verified database backup and configuration snapshot",
	RunE: func(cmd *cobra.Command, _args []string) error {
		ctx := cmd.Context()
		mgr := immich.NewManager("")

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

		res, err := mgr.RunBackup(ctx, immichRetention)
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

var immichRestoreCmd = &cobra.Command{
	Use:   "restore [backup-file]",
	Short: "Restore Immich database and configuration from backup",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		mgr := immich.NewManager("")

		targetFile := immichRestoreFile
		if len(args) > 0 {
			targetFile = args[0]
		}

		return mgr.RunRestore(ctx, targetFile, immichAutoConfirm)
	},
}

var immichStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start Immich service stack via systemd",
	RunE: func(cmd *cobra.Command, _args []string) error {
		ctx := cmd.Context()
		mgr := immich.NewManager("")

		if err := mgr.Start(ctx); err != nil {
			return err
		}
		theme.SuccessMessage("Immich service started successfully.")
		return nil
	},
}

var immichStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Gracefully stop Immich service stack",
	RunE: func(cmd *cobra.Command, _args []string) error {
		ctx := cmd.Context()
		mgr := immich.NewManager("")

		if err := mgr.Stop(ctx); err != nil {
			return err
		}
		theme.SuccessMessage("Immich service stopped gracefully.")
		return nil
	},
}

var immichRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart Immich service stack via systemd",
	RunE: func(cmd *cobra.Command, _args []string) error {
		ctx := cmd.Context()
		mgr := immich.NewManager("")

		if err := mgr.Restart(ctx); err != nil {
			return err
		}
		theme.SuccessMessage("Immich service restarted successfully.")
		return nil
	},
}

var immichLogsCmd = &cobra.Command{
	Use:     "logs",
	Aliases: []string{"log"},
	Short:   "View live Immich service or container logs",
	RunE: func(cmd *cobra.Command, _args []string) error {
		ctx := cmd.Context()
		mgr := immich.NewManager("")

		log.Verbose(false, "Streaming Immich logs...")
		return mgr.Logs(ctx, immichLogsService, immichLogsFollow, immichLogsTail)
	},
}

func init() {
	immichStatusCmd.Flags().BoolVar(&immichJSONFlag, "json", false, "Output status in JSON format")
	immichStatusCmd.Flags().BoolVarP(&immichPagerFlag, "pager", "p", false, "Open status inside scrollable viewport")

	immichBackupCmd.Flags().IntVarP(&immichRetention, "retention", "r", 0, "Override retention period (days)")

	immichRestoreCmd.Flags().StringVarP(&immichRestoreFile, "file", "f", "", "Specific backup archive path to restore")
	immichRestoreCmd.Flags().BoolVarP(&immichAutoConfirm, "yes", "y", false, "Auto-confirm restoration prompt")

	immichLogsCmd.Flags().BoolVarP(&immichLogsFollow, "follow", "f", true, "Follow log stream")
	immichLogsCmd.Flags().IntVarP(&immichLogsTail, "tail", "n", 50, "Number of lines to show from end of logs")
	immichLogsCmd.Flags().
		StringVarP(&immichLogsService, "service", "s", "", "Filter to specific container (e.g. server, database, ml, redis)")

	ImmichCmd.AddCommand(immichStatusCmd)
	ImmichCmd.AddCommand(immichBackupCmd)
	ImmichCmd.AddCommand(immichRestoreCmd)
	ImmichCmd.AddCommand(immichStartCmd)
	ImmichCmd.AddCommand(immichStopCmd)
	ImmichCmd.AddCommand(immichRestartCmd)
	ImmichCmd.AddCommand(immichLogsCmd)
}
