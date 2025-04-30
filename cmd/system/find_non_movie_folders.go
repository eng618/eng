package system

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/log"
	"github.com/spf13/cobra"
)

// FindNonMovieFoldersCmd defines the cobra command for finding and optionally deleting
// directories that do not contain common video file types.
var FindNonMovieFoldersCmd = &cobra.Command{
	Use:   "findNonMovieFolders [directory]",
	Short: "Find and optionally delete non-movie folders",
	Long: `This command searches recursively through the supplied directory for directories
that do not contain video files (mp4, mkv, avi, mov, wmv, flv, webm, mpeg, mpg, m4v).
It identifies top-level subdirectories within the supplied directory that lack
any such files anywhere within their structure.

It lists the files within the identified folders before taking action.
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

		nonMovieFolders, err := findNonMovieFolders(isVerbose, directory, func(done, total int) {
			progress := 0.0
			if total > 0 {
				progress = float64(done) / float64(total)
			}
			spinner.SetProgressBar(progress, fmt.Sprintf("Scanning... (%d/%d)", done, total))
		})

		// Explicitly Stop Spinner before printing results
		if err != nil {
			spinner.Stop() // Stop spinner even if there was an error during scan
			log.Error("Error finding non-movie folders: %s", err)
			return
		}
		spinner.UpdateMessage("Scan complete.")
		spinner.SetProgressBar(1.0) // Ensure it shows 100%
		spinner.Stop()
		time.Sleep(100 * time.Millisecond) // Allow terminal to process spinner stop

		log.Verbose(isVerbose, "Found %d potential non-movie folders.", len(nonMovieFolders))

		if len(nonMovieFolders) == 0 {
			log.Success("No non-movie folders found in %s.", directory)
			return
		}

		log.Message("\nProcessing %d non-movie folder(s)...", len(nonMovieFolders))

		deletedCount := 0
		skippedCount := 0
		errorMessages := []string{}

		for _, folder := range nonMovieFolders {
			// List files within the folder for detailed output.
			// Use find to get relative paths from within the folder itself.
			listFilesCmd := exec.Command("find", ".", "-type", "f") // Find files relative to the folder
			listFilesCmd.Dir = folder                               // Execute the command *inside* the target folder

			filesToDeleteBytes, listErr := listFilesCmd.Output()

			var actualFiles []string
			var fileCount int
			var listErrorString string

			if listErr != nil {
				// Capture the specific error for later display if listing fails.
				errMsg := fmt.Sprintf("Could not list files in folder %s (may be empty or permission issue): %s", folder, listErr)
				log.Warn(errMsg)
				errorMessages = append(errorMessages, errMsg)
				listErrorString = "(error listing files)"
				fileCount = 0 // Assume 0 if we can't list
			} else {
				filesList := strings.Split(strings.TrimSpace(string(filesToDeleteBytes)), "\n")
				// Filter out empty strings which can happen if the output is empty.
				if len(filesList) > 0 && filesList[0] != "" {
					actualFiles = filesList
				}
				fileCount = len(actualFiles)
				listErrorString = "" // No error string needed if successful
			}

			// Output the folder and its contents (or status)
			if isDryRun {
				log.Message("[Dry Run] Would delete folder: %s", folder)
				if listErrorString != "" {
					log.Message("  %s", listErrorString) // Show error if listing failed
				} else if fileCount > 0 {
					for _, file := range actualFiles {
						// Construct a display path relative to the parent of the folder being processed.
						displayPath := filepath.Join(filepath.Base(folder), strings.TrimPrefix(file, "./"))
						log.Message("  - %s", displayPath)
					}
				} else {
					log.Message("  (Contains no files or only empty subdirectories)")
				}
				skippedCount++
			} else { // Not Dry Run
				log.Warn("Preparing to delete non-movie folder: %s", folder)
				if listErrorString != "" {
					log.Warn("  %s", listErrorString) // Show list error before attempting delete
				} else if fileCount > 0 {
					log.Message("  Files within this folder:")
					for _, file := range actualFiles {
						displayPath := filepath.Join(filepath.Base(folder), strings.TrimPrefix(file, "./"))
						log.Message("    - %s", displayPath)
					}
				} else {
					log.Message("  (Folder contains no files or only empty subdirectories)")
				}

				// Perform the deletion
				log.Warn("Executing delete: rm -rf %s", folder)
				deleteCmd := exec.Command("rm", "-rf", "--", folder)
				if err := deleteCmd.Run(); err != nil {
					errMsg := fmt.Sprintf("Error deleting folder %s: %s", folder, err)
					log.Error(errMsg)
					errorMessages = append(errorMessages, errMsg)
					skippedCount++
				} else {
					log.Success("Deleted: %s", folder)
					deletedCount++
				}
			}
			log.Message("") // Add a blank line between folder processing for readability
		}

		// Final summary
		if len(errorMessages) > 0 {
			log.Warn("Encountered %d error(s) during processing.", len(errorMessages))
		}

		if isDryRun {
			log.Success("Dry run complete. %d folder(s) identified for deletion.", skippedCount)
		} else {
			log.Success("Processing complete. Deleted %d folder(s), skipped %d due to errors.", deletedCount, skippedCount)
		}
	},
}

// findNonMovieFolders scans the immediate subdirectories of rootDir.
// It returns a slice of paths to subdirectories that do not contain any files
// matching common video extensions (recursively within each subdirectory).
//
// Parameters:
//   - isVerbose: If true, logs verbose messages during the scan.
//   - rootDir: The path to the directory whose subdirectories will be scanned.
//   - progress: A callback function to report progress (done, total). Can be nil.
//
// Returns:
//   - A slice of strings, where each string is the absolute path to a non-movie folder.
//   - An error if reading the root directory fails or if there's an unexpected issue executing 'find'.
func findNonMovieFolders(isVerbose bool, rootDir string, progress func(done, total int)) ([]string, error) {
	var nonMovieFolders []string

	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", rootDir, err)
	}

	// Filter for directories only
	var dirEntries []os.DirEntry
	for _, entry := range entries {
		if entry.IsDir() {
			dirEntries = append(dirEntries, entry)
		}
	}

	total := len(dirEntries)
	done := 0

	if progress != nil {
		progress(done, total) // Initial progress report (0/total)
	}

	for _, entry := range dirEntries {
		dirPath := filepath.Join(rootDir, entry.Name())
		log.Verbose(isVerbose, "Checking directory: %s", dirPath)

		// Use find to search recursively for *any* movie file within the subdirectory.
		// The command uses -quit to exit immediately after finding the first match for efficiency.
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
			")", "-print", "-quit")

		output, err := checkCmd.Output()

		// Check the error type. An ExitError is expected if find exits due to -quit
		// (finding a file) or if it finds nothing. Any other error is unexpected.
		if err != nil {
			exitErr, ok := err.(*exec.ExitError)
			// If it's NOT an ExitError, it's some other problem (permissions, command not found etc.)
			if !ok {
				log.Warn("Unexpected error executing find command in %s: %v. Skipping directory.", dirPath, err)
				done++
				if progress != nil {
					progress(done, total)
				}
				continue // Skip this directory due to the error
			}
			// If it *is* an ExitError, we proceed to check the output length.
			// An ExitError with non-empty output means -quit worked (movie file found).
			// An ExitError with empty output means nothing was found.
			_ = exitErr // Avoid unused variable error if not debugging
		}

		// If the output is empty, no movie files were found.
		if len(strings.TrimSpace(string(output))) == 0 {
			log.Verbose(isVerbose, "No movie files found in: %s", dirPath)
			nonMovieFolders = append(nonMovieFolders, dirPath)
		} else {
			log.Verbose(isVerbose, "Movie file(s) found in %s.", dirPath)
		}

		done++
		if progress != nil {
			progress(done, total)
		}
	}

	return nonMovieFolders, nil
}
