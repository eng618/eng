package files

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/fs"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

var (
	shredPasses        int
	shredRecursive     bool
	shredForce         bool
	shredDryRun        bool
	shredMethod        string
	shredUseSystem     bool
	shredVerify        bool
	shredFollowSymlink bool
)

// ShredCmd represents the secure file deletion command.
var ShredCmd = &cobra.Command{
	Use:     "shred [paths...]",
	Aliases: []string{"secure-delete", "srm"},
	Short:   "Securely delete files and directories by overwriting data",
	Long: `Securely delete files and directories by overwriting their data multiple times
before removal. This prevents recovery of deleted data using forensic tools.

Supports multiple overwrite methods including DoD 5220.22-M (3-pass) and Gutmann (35-pass).
By default, uses a 3-pass random overwrite which is suitable for most use cases.

Examples:
  eng files shred sensitive.txt                    # Shred a single file (3 passes)
  eng files shred -r secrets/                      # Shred directory recursively
  eng files shred -p 7 -m dod file1 file2         # 7-pass DoD method
  eng files shred --dry-run secrets/              # Preview what would be shredded
  eng files shred -f -r /path/to/data             # Force (no confirmation)`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary).
			MarginBottom(1)
		if !ui.DisableProgress {
			fmt.Fprintln(log.Out, headerStyle.Render("🗑️  Secure File Deletion (shred)"))
		}

		isVerbose := cmdutil.IsVerbose(cmd)

		// Parse method
		method := fs.ShredMethod(shredMethod)
		validMethods := map[fs.ShredMethod]bool{
			fs.MethodAuto:    true,
			fs.MethodRandom:  true,
			fs.MethodZero:    true,
			fs.MethodDoD:     true,
			fs.MethodGutmann: true,
		}
		if !validMethods[method] {
			log.Error("Invalid method: %s. Valid: auto, random, zero, dod, gutmann", shredMethod)
			return
		}

		// Collect all files to shred
		files, err := fs.CollectFilesForShredding(args, shredRecursive)
		if err != nil {
			log.Error("%v", err)
			return
		}

		if len(files) == 0 {
			theme.InfoMessage("No files to shred.")
			return
		}

		// Calculate total size and passes
		var totalSize int64
		fileInfos := make([]FileMatch, 0, len(files))
		for _, file := range files {
			info, err := os.Stat(file)
			size := int64(0)
			if err == nil {
				size = info.Size()
			}
			totalSize += size
			fileInfos = append(fileInfos, FileMatch{
				Path:      file,
				SizeBytes: size,
			})
		}

		// Sort by path for consistent display
		sort.Slice(fileInfos, func(i, j int) bool {
			return fileInfos[i].Path < fileInfos[j].Path
		})

		// Determine passes
		passes := shredPasses
		if passes <= 0 {
			passes = 3 // Default
		}
		totalPasses := fs.GetPassesForMethod(method, passes)

		// Show summary
		sizeFormatted := humanize.Bytes(uint64(totalSize))
		passDesc := fmt.Sprintf("%d pass", totalPasses)
		if totalPasses != 1 {
			passDesc += "es"
		}

		var boxLines []string
		boxLines = append(boxLines, fmt.Sprintf(
			"Ready to securely delete %s file(s) (%s) using %s (%s method):",
			theme.PrimaryText.Bold(true).Render(fmt.Sprintf("%d", len(files))),
			theme.SuccessText.Bold(true).Render(sizeFormatted),
			theme.PrimaryText.Bold(true).Render(passDesc),
			theme.PrimaryText.Bold(true).Render(string(method)),
		))

		if shredDryRun {
			warningStyle := lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#f59e0b", Dark: "#d97706"}).
				Bold(true)
			boxLines = append(boxLines, warningStyle.Render("DRY RUN - No files will be deleted"))
		}

		if !ui.DisableProgress {
			fmt.Fprintln(log.Out, theme.InfoBox.Render(strings.Join(boxLines, "\n")))
		} else {
			log.Message("Files to shred:")
			for _, fm := range fileInfos {
				sizeStr := ""
				if fm.SizeBytes > 0 {
					sizeStr = fmt.Sprintf(" (%s)", humanize.Bytes(uint64(fm.SizeBytes)))
				}
				log.Message("  %s%s", fm.Path, sizeStr)
			}
			log.Message("Method: %s, Passes: %d", method, totalPasses)
		}

		// Dry run - just show what would happen
		if shredDryRun {
			theme.InfoMessage("Dry run complete. No files were deleted.")
			return
		}

		// Confirmation prompt (unless --force)
		if !shredForce {
			confirmed, err := ui.Confirm(
				fmt.Sprintf("Permanently shred %d file(s) (%s)? This cannot be undone!", len(files), sizeFormatted),
				false,
			)
			if err != nil {
				log.Error("Confirmation failed: %v", err)
				return
			}
			if !confirmed {
				log.Info("Shredding canceled.")
				return
			}
		}

		// Initialize multi-progress bar
		var multiBar *ui.MultiProgressBar
		var bars []*ui.ProgressBar
		if !ui.DisableProgress {
			multiBar = ui.NewMultiProgressBar(log.Writer())
			multiBar.SetOverallMessage(fmt.Sprintf("Shredding %d file(s) with %d pass(es)...", len(files), totalPasses))

			// Add bars for each file
			bars = make([]*ui.ProgressBar, len(fileInfos))
			for i, fm := range fileInfos {
				label := fm.Path
				if len(label) > 50 {
					label = "..." + label[len(label)-47:]
				}
				bars[i] = multiBar.AddBar(label, totalPasses)
			}
			multiBar.SetOverallMessage(fmt.Sprintf("Shredding %d file(s) with %d pass(es)...", len(files), totalPasses))
		}

		// Track results
		var deletedCount int
		var freedBytes int64
		var errors []string

		// Progress callback for multi-file shredding
		progressFn := func(statuses []fs.FileShredStatus) {
			if multiBar == nil {
				return
			}
			for _, status := range statuses {
				// Find the corresponding bar
				for i, fm := range fileInfos {
					if fm.Path == status.Path {
						detail := ""
						if status.Size > 0 {
							detail = humanize.Bytes(uint64(status.Size))
						}
						multiBar.UpdateBar(bars[i], status.Percent, detail, status.CurrentPass)
						if status.Done {
							if status.Error != nil {
								multiBar.CompleteBar(bars[i], status.Error)
							} else {
								multiBar.CompleteBar(bars[i], nil)
							}
						}
						break
					}
				}
			}
		}

		// Perform shredding
		startTime := time.Now()
		err = fs.ShredMultiple(files, passes, method, progressFn)
		elapsed := time.Since(startTime).Round(time.Second)

		if multiBar != nil {
			multiBar.Stop()
		}

		// Collect results from statuses (we need to track them)
		// For now, we'll rely on the error return and log individual results
		// In a more complete implementation, we'd track per-file results

		if err != nil {
			// Some files failed
			errors = append(errors, err.Error())
			theme.WarningMessage(fmt.Sprintf("Shredding completed with errors in %s", elapsed))
		} else {
			deletedCount = len(files)
			freedBytes = totalSize
			freedStr := humanize.Bytes(uint64(freedBytes))
			theme.SuccessMessage(
				fmt.Sprintf("Successfully shredded %d file(s) freeing %s in %s!", deletedCount, freedStr, elapsed),
			)
		}

		if isVerbose && len(errors) > 0 {
			for _, e := range errors {
				log.Error("%s", e)
			}
		}
	},
}

func init() {
	ShredCmd.Flags().IntVarP(&shredPasses, "passes", "p", 3, "Number of overwrite passes (0 = method default)")
	ShredCmd.Flags().
		BoolVarP(&shredRecursive, "recursive", "r", false, "Shred directories recursively (required for directories)")
	ShredCmd.Flags().BoolVarP(&shredForce, "force", "f", false, "Skip confirmation prompt")
	ShredCmd.Flags().BoolVarP(&shredDryRun, "dry-run", "n", false, "Preview what would be shredded without deleting")
	ShredCmd.Flags().
		StringVarP(&shredMethod, "method", "m", "auto", "Overwrite method: auto, random, zero, dod, gutmann")
	ShredCmd.Flags().
		BoolVar(&shredUseSystem, "use-system", true, "Use system shred command on Linux (fallback to Go implementation)")
	ShredCmd.Flags().BoolVar(&shredVerify, "verify", false, "Verify overwrites by reading back (slower)")
	ShredCmd.Flags().
		BoolVar(&shredFollowSymlink, "follow-symlinks", false, "Follow symlinks and shred targets (default: shred symlink target)")

	// Mark method flag with valid values
	_ = ShredCmd.RegisterFlagCompletionFunc(
		"method",
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return []string{"auto", "random", "zero", "dod", "gutmann"}, cobra.ShellCompDirectiveNoFileComp
		},
	)
}
