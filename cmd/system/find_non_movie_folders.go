package system

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/log"
	"github.com/spf13/cobra"
)

var FindNonMovieFoldersCmd = &cobra.Command{
	Use:   "findNonMovieFolders [directory]",
	Short: "Find and optionally delete non-movie folders",
	Long: `This command searches recursively through the supplied directory for directories
that do not contain video files (mp4, mkv, avi, mov, wmv, flv, webm, mpeg, mpg, m4v).
It identifies top-level subdirectories within the supplied directory that lack
any such files anywhere within their structure.

It can optionally delete these identified directories if the --dry-run flag is set to false.
By default, --dry-run is true.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		log.Start("Scanning for non-movie folders...")

		directory := args[0]
		isDryRun, _ := cmd.Flags().GetBool("dry-run")
		isVerbose := utils.IsVerbose(cmd)

		// Validate directory exists
		if _, err := os.Stat(directory); os.IsNotExist(err) {
			log.Error("Provided directory does not exist: %s", directory)
			return
		}

		log.Verbose(isVerbose, "Searching for directories in: %s", directory)
		log.Verbose(isVerbose, "Dry run mode: %t", isDryRun)

		spinner := utils.NewProgressSpinner("Scanning directories...")
		spinner.Start()
		defer spinner.Stop()

		nonMovieFolders, err := findNonMovieFolders(isVerbose, directory, func(done, total int) {
			progress := 0.0
			if total > 0 {
				progress = float64(done) / float64(total)
			}
			spinner.SetProgressBar(progress, fmt.Sprintf("Scanning... (%d/%d)", done, total))
		})

		if err != nil {
			log.Error("Error finding non-movie folders: %s", err)
			return
		}

		// Update spinner message after scanning is complete
		spinner.UpdateMessage("Scan complete.")
		spinner.SetProgressBar(1.0)

		log.Verbose(isVerbose, "Found %d potential non-movie folders.", len(nonMovieFolders))

		if len(nonMovieFolders) == 0 {
			log.Success("No non-movie folders found in %s.", directory)
			return
		}

		log.Message("Processing %d non-movie folder(s)...", len(nonMovieFolders))

		deletedCount := 0
		skippedCount := 0

		for _, folder := range nonMovieFolders {
			// List files for logging purposes, handle potential errors
			listFilesCmd := exec.Command("find", folder, "-type", "f")
			filesToDeleteBytes, listErr := listFilesCmd.Output()
			if listErr != nil {
				log.Warn("Could not list files in folder %s (may be empty or permission issue): %s", folder, listErr)
			}

			filesList := strings.Split(strings.TrimSpace(string(filesToDeleteBytes)), "\n")
			actualFiles := []string{}
			if len(filesList) > 0 && filesList[0] != "" {
				actualFiles = filesList
			}

			if isVerbose && len(actualFiles) > 0 {
				log.Info("Files within %s:", folder)
				for _, file := range actualFiles {
					log.Info("  - %s", file)
				}
			} else if isVerbose {
				log.Info("Folder %s is empty or contains only empty subdirectories.", folder)
			}

			if isDryRun {
				log.Message("[Dry Run] Would delete folder: %s (%d files/items)", folder, len(actualFiles))
				skippedCount++
			} else {
				log.Warn("Deleting non-movie folder: %s (%d files/items)", folder, len(actualFiles))
				deleteCmd := exec.Command("rm", "-rf", "--", folder)
				if err := deleteCmd.Run(); err != nil {
					log.Error("Error deleting folder %s: %s", folder, err)
					skippedCount++
				} else {
					log.Success("Deleted: %s", folder)
					deletedCount++
				}
			}
		}

		// Final summary
		if isDryRun {
			log.Success("Dry run complete. %d folder(s) identified for deletion.", skippedCount)
		} else {
			log.Success("Processing complete. Deleted %d folder(s), skipped %d due to errors.", deletedCount, skippedCount)
		}
	},
}

// findNonMovieFolders returns a list of top-level directories under rootDir
// that do not contain any movie files anywhere inside them.
func findNonMovieFolders(isVerbose bool, rootDir string, progress func(done, total int)) ([]string, error) {
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

	for _, entry := range dirEntries {
		dirPath := filepath.Join(rootDir, entry.Name())
		log.Verbose(isVerbose, "Checking directory: %s", dirPath)

		// Use find to search recursively for any movie file.
		// REMOVED "-quit". Let find search the whole directory.
		checkCmd := exec.Command("find", dirPath, "-type", "f", "(",
			"-iname", "*.mp4", "-o",
			"-iname", "*.mkv", "-o",
			"-iname", "*.avi", "-o",
			"-iname", "*.mov", "-o",
			"-iname", "*.wmv", "-o",
			"-iname", "*.flv", "-o",
			"-iname", "*.webm", "-o",
			"-iname", "*.mpeg", "-o",
			"-iname", "*.mpg", "-o",
			"-iname", "*.m4v",
			")", "-print") // Removed -quit

		// --- Logic Change: Rely primarily on output length ---
		output, err := checkCmd.Output()

		// Check for critical errors running find itself (not just "not found")
		if err != nil {
			// If it's an ExitError, find ran but exited non-zero. This is OKAY if output is empty.
			// If it's NOT an ExitError, it's a more serious problem (e.g., find not found, permissions on dirPath itself).
			if _, ok := err.(*exec.ExitError); !ok {
				// This is NOT an ExitError, log it as a warning and skip.
				log.Warn("Error executing find command in %s: %v. Skipping directory.", dirPath, err)
				// Update progress before skipping
				done++
				if progress != nil {
					progress(done, total)
				}
				continue // Skip this directory
			}
			// If it IS an ExitError, we proceed to check len(output) below.
			// Find might exit non-zero if it encounters permission errors *within* subdirs,
			// but it might still have found files (output > 0) or not (output == 0).
		}

		// At this point, find either exited cleanly (err == nil) or with an ExitError.
		// The definitive check is now the length of the output.
		if len(output) == 0 {
			// No output means no movie files were found anywhere in dirPath.
			log.Verbose(isVerbose, "No movie files found in: %s", dirPath)
			nonMovieFolders = append(nonMovieFolders, dirPath)
		} else {
			// Output exists, means at least one movie file was found.
			log.Verbose(isVerbose, "Movie file(s) found in %s.", dirPath)
			// Optionally log the first found file for debugging:
			// log.Verbose(isVerbose, "Movie file(s) found in %s (e.g., %s)", dirPath, strings.Split(string(output), "\n")[0])
		}

		// Update progress after checking each directory
		done++
		if progress != nil {
			progress(done, total)
		}
	} // End of loop through directories

	return nonMovieFolders, nil
}

// Ensure utils.IsVerbose exists or replace with direct flag check
// func IsVerbose(cmd *cobra.Command) bool {
// 	verbose, _ := cmd.Flags().GetBool("verbose") // Assuming a "verbose" flag exists
// 	return verbose
// }

// Ensure init function adds flags correctly (likely in system.go)
// func init() {
// 	FindNonMovieFoldersCmd.Flags().Bool("dry-run", true, "Perform a dry run without deleting folders")
// 	FindNonMovieFoldersCmd.Flags().BoolP("verbose", "v", false, "Enable verbose logging") // Example verbose flag
// }
