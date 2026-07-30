package files

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/asdf"
	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

// FolderMatch represents a detected non-movie folder with disk size and file count.
type FolderMatch struct {
	Path      string
	SizeBytes int64
	FileCount int
}

// FormatLabel returns a user-friendly display string with folder path, size, and file count.
func (f FolderMatch) FormatLabel() string {
	sizeStr := humanize.Bytes(uint64(f.SizeBytes))
	return fmt.Sprintf("%s (%s, %d file(s))", f.Path, sizeStr, f.FileCount)
}

// FindNonMovieFoldersCmd defines the cobra command for finding and optionally deleting
// directories that do not contain common video file types.
var FindNonMovieFoldersCmd = &cobra.Command{
	Use:     "findNonMovieFolders [directory]",
	Aliases: []string{"find-non-movie-folders", "non-movie-folders", "clean-folders"},
	Short:   "Find and optionally delete non-movie folders",
	Long: `This command searches recursively through the supplied directory for directories
that do not contain video files (mp4, mkv, avi, mov, wmv, flv, webm, mpeg, mpg, m4v).
It identifies top-level subdirectories within the supplied directory that lack
any such files anywhere within their structure.

It calculates folder disk space, lists folder contents, and prompts for confirmation before deletion.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary).
			MarginBottom(1)
		if !ui.DisableProgress {
			fmt.Fprintln(log.Out, headerStyle.Render("🎬 Find Non-Movie Folders"))
		}

		directory := args[0]
		isVerbose := cmdutil.IsVerbose(cmd)

		if _, err := os.Stat(directory); os.IsNotExist(err) {
			log.Error("Provided directory does not exist: %s", directory)
			return
		}

		var spinner *ui.Spinner
		if !ui.DisableProgress {
			spinner = ui.NewProgressSpinner("Scanning directories for non-movie folders...")
		}

		nonMovieFolders, err := findNonMovieFolders(isVerbose, directory, spinner, func(done, total int) {
			progress := 0.0
			if total > 0 {
				progress = float64(done) / float64(total)
			}
			if spinner != nil {
				spinner.SetProgressBar(progress, fmt.Sprintf("Scanning... (%d/%d directories)", done, total))
			}
		})

		if spinner != nil {
			spinner.Stop()
		}

		if err != nil {
			log.Error("Error finding non-movie folders: %s", err)
			return
		}

		if len(nonMovieFolders) == 0 {
			theme.SuccessMessage(fmt.Sprintf("No non-movie folders found in %s.", directory))
			return
		}

		// Inspect each non-movie folder for size and file count
		var folderMatches []FolderMatch
		var totalReclaimableBytes int64
		var totalFiles int

		for _, folder := range nonMovieFolders {
			sz := asdf.CalculateDirSize(folder)
			fc := countFilesInDir(folder)

			folderMatches = append(folderMatches, FolderMatch{
				Path:      folder,
				SizeBytes: sz,
				FileCount: fc,
			})

			totalReclaimableBytes += sz
			totalFiles += fc
		}

		totalSizeStr := humanize.Bytes(uint64(totalReclaimableBytes))

		// Render Callout Box
		var boxLines []string
		boxLines = append(boxLines, fmt.Sprintf("Found %s non-movie folder(s) reclaiming %s space (%d total file(s)):",
			theme.PrimaryText.Bold(true).Render(fmt.Sprintf("%d", len(folderMatches))),
			theme.SuccessText.Bold(true).Render(totalSizeStr),
			totalFiles,
		))

		for _, fm := range folderMatches {
			boxLines = append(boxLines, fmt.Sprintf(
				"  • %s %s",
				theme.BoldText.Render(fm.Path),
				theme.MutedText.Render(
					fmt.Sprintf("(%s, %d files)", humanize.Bytes(uint64(fm.SizeBytes)), fm.FileCount),
				),
			))
		}

		if !ui.DisableProgress {
			fmt.Println(theme.InfoBox.Render(strings.Join(boxLines, "\n")))
		}

		// Prepare MultiSelect options
		labelMap := make(map[string]FolderMatch)
		var options []string
		for _, fm := range folderMatches {
			label := fm.FormatLabel()
			labelMap[label] = fm
			options = append(options, label)
		}

		confirmPromptMsg := fmt.Sprintf("Select non-movie folders to delete (%s total reclaimable):", totalSizeStr)
		selectedLabels, err := ui.MultiSelect(confirmPromptMsg, options, options)
		if err != nil {
			log.Error("Error during confirmation prompt: %v", err)
			return
		}

		if len(selectedLabels) == 0 {
			log.Info("Deletion canceled or no folders selected.")
			return
		}

		var selectedFolders []FolderMatch
		for _, label := range selectedLabels {
			if fm, ok := labelMap[label]; ok {
				selectedFolders = append(selectedFolders, fm)
			}
		}

		totalToDelete := len(selectedFolders)
		var deleteSpinner *ui.Spinner
		if !ui.DisableProgress {
			deleteSpinner = ui.NewProgressSpinner(fmt.Sprintf("Deleting %d non-movie folder(s)...", totalToDelete))
		}

		var deletedCount int
		var freedBytes int64
		var errorCount int

		for i, fm := range selectedFolders {
			ratio := float64(i+1) / float64(totalToDelete)
			statusMsg := fmt.Sprintf("[%d/%d] Removing %s", i+1, totalToDelete, filepath.Base(fm.Path))

			if deleteSpinner != nil {
				deleteSpinner.SetProgressBar(ratio, statusMsg)
			}

			if err := os.RemoveAll(fm.Path); err != nil {
				errorCount++
				if deleteSpinner != nil {
					deleteSpinner.Logf("  %s Failed to remove %s: %v\n", theme.ErrorText.Render("✗"), fm.Path, err)
				} else {
					log.Error("Error deleting folder %s: %v", fm.Path, err)
				}
			} else {
				deletedCount++
				freedBytes += fm.SizeBytes
				sizeStr := humanize.Bytes(uint64(fm.SizeBytes))
				if deleteSpinner != nil {
					deleteSpinner.Logf(
						"  %s Deleted %s (freed %s, %d files)\n",
						theme.SuccessText.Render("✓"),
						fm.Path,
						sizeStr,
						fm.FileCount,
					)
				} else {
					log.Success("Deleted: %s (freed %s)", fm.Path, sizeStr)
				}
			}
		}

		if deleteSpinner != nil {
			deleteSpinner.SetProgressBar(1.0, "Folder deletion complete")
			deleteSpinner.Stop()
		}

		freedStr := humanize.Bytes(uint64(freedBytes))
		if errorCount > 0 {
			theme.WarningMessage(
				fmt.Sprintf(
					"Deleted %d non-movie folder(s) freeing %s, but encountered %d error(s).",
					deletedCount,
					freedStr,
					errorCount,
				),
			)
		} else {
			theme.SuccessMessage(
				fmt.Sprintf(
					"Successfully deleted %d non-movie folder(s) freeing %s of disk space!",
					deletedCount,
					freedStr,
				),
			)
		}
	},
}

func countFilesInDir(dirPath string) int {
	var count int
	_ = filepath.WalkDir(dirPath, func(_ string, d os.DirEntry, err error) error {
		if err != nil || d == nil {
			return nil
		}
		if !d.IsDir() {
			count++
		}
		return nil
	})
	return count
}

func findNonMovieFolders(
	isVerbose bool,
	rootDir string,
	spinner *ui.Spinner,
	progress func(done, total int),
) ([]string, error) {
	var nonMovieFolders []string

	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", rootDir, err)
	}

	var dirEntries []os.DirEntry
	for _, entry := range entries {
		if entry.IsDir() {
			dirEntries = append(dirEntries, entry)
		}
	}

	total := len(dirEntries)
	done := 0

	if progress != nil {
		progress(done, total)
	}

	videoExtensions := map[string]bool{
		".mp4": true, ".mkv": true, ".avi": true, ".mov": true, ".wmv": true,
		".flv": true, ".webm": true, ".mpeg": true, ".mpg": true, ".m4v": true,
	}

	for _, entry := range dirEntries {
		dirPath := filepath.Join(rootDir, entry.Name())
		if isVerbose && spinner != nil {
			spinner.Logf("--- Checking directory: %s\n", dirPath)
		}

		foundMovieFile := false
		walkErr := filepath.WalkDir(dirPath, func(_ string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				ext := strings.ToLower(filepath.Ext(d.Name()))
				if videoExtensions[ext] {
					foundMovieFile = true
					return filepath.SkipAll
				}
			}
			return nil
		})

		if walkErr != nil && !errors.Is(walkErr, filepath.SkipAll) {
			log.Warn("Error scanning directory %s: %v. Skipping.", dirPath, walkErr)
		}

		if !foundMovieFile {
			if isVerbose && spinner != nil {
				spinner.Logf("--- No movie files found in: %s\n", dirPath)
			}
			nonMovieFolders = append(nonMovieFolders, dirPath)
		} else {
			if isVerbose && spinner != nil {
				spinner.Logf("--- Movie file(s) found in %s.\n", dirPath)
			}
		}

		done++
		if progress != nil {
			progress(done, total)
		}
	}

	return nonMovieFolders, nil
}
