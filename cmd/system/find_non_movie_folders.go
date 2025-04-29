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

		nonMovieFolders, err := findNonMovieFolders(isVerbose, directory, func(done, total int) {
			progress := 0.0
			if total > 0 {
				progress = float64(done) / float64(total)
			}
			spinner.SetProgressBar(progress, fmt.Sprintf("Scanning... (%d/%d)", done, total))
		})

		// --- Explicitly Stop Spinner ---
		if err != nil {
			spinner.Stop() // Stop spinner even if there was an error during scan
			log.Error("Error finding non-movie folders: %s", err)
			return
		}

		// Update spinner message to final state and stop it *before* printing results
		spinner.UpdateMessage("Scan complete.")
		spinner.SetProgressBar(1.0) // Ensure it shows 100%
		spinner.Stop()              // Stop the spinner now!
		// Give the terminal a brief moment to process the spinner stop, might help prevent race conditions
		time.Sleep(100 * time.Millisecond)

		// --- End Spinner Stop ---

		log.Verbose(isVerbose, "Found %d potential non-movie folders.", len(nonMovieFolders))

		if len(nonMovieFolders) == 0 {
			log.Success("No non-movie folders found in %s.", directory)
			return
		}

		// Now print the processing message *after* the spinner is gone
		log.Message("\nProcessing %d non-movie folder(s)...", len(nonMovieFolders)) // Added newline for clarity

		deletedCount := 0
		skippedCount := 0
		errorMessages := []string{}

		for _, folder := range nonMovieFolders {
			// List files for logging purposes, handle potential errors
			listFilesCmd := exec.Command("find", folder, "-type", "f")
			filesToDeleteBytes, listErr := listFilesCmd.Output()

			// Prepare file list string for logging
			var filesListStr string
			var fileCount int
			if listErr != nil {
				// Capture the specific error for later display
				errMsg := fmt.Sprintf("Could not list files in folder %s (may be empty or permission issue): %s", folder, listErr)
				log.Warn(errMsg) // Log immediately as warning
				errorMessages = append(errorMessages, errMsg) // Also collect if needed later
				filesListStr = "(error listing files)"
				fileCount = 0 // Assume 0 if we can't list
			} else {
				filesList := strings.Split(strings.TrimSpace(string(filesToDeleteBytes)), "\n")
				actualFiles := []string{}
				if len(filesList) > 0 && filesList[0] != "" {
					actualFiles = filesList
				}
				fileCount = len(actualFiles)

				if isVerbose && fileCount > 0 {
					log.Info("Files within %s:", folder)
					for _, file := range actualFiles {
						log.Info("  - %s", file)
					}
				} else if isVerbose {
					log.Info("Folder %s is empty or contains only empty subdirectories.", folder)
				}
				filesListStr = fmt.Sprintf("%d files/items", fileCount)
			}


			if isDryRun {
				log.Message("[Dry Run] Would delete folder: %s (%s)", folder, filesListStr)
				skippedCount++
			} else {
				log.Warn("Deleting non-movie folder: %s (%s)", folder, filesListStr)
				deleteCmd := exec.Command("rm", "-rf", "--", folder)
				if err := deleteCmd.Run(); err != nil {
					errMsg := fmt.Sprintf("Error deleting folder %s: %s", folder, err)
					log.Error(errMsg)
					errorMessages = append(errorMessages, errMsg) // Collect error
					skippedCount++
				} else {
					log.Success("Deleted: %s", folder)
					deletedCount++
				}
			}
		}

		// Final summary
		log.Message("") // Add a blank line before summary for better spacing
		if len(errorMessages) > 0 {
			log.Warn("Encountered %d error(s) during processing.", len(errorMessages))
			// Optionally print collected errors again here if needed
			// for _, errMsg := range errorMessages {
			// 	log.Warn("  - %s", errMsg)
			// }
		}

		if isDryRun {
			log.Success("Dry run complete. %d folder(s) identified for deletion.", skippedCount)
		} else {
			log.Success("Processing complete. Deleted %d folder(s), skipped %d due to errors.", deletedCount, skippedCount)
		}
	},
}

// findNonMovieFolders function remains the same...
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

		output, err := checkCmd.Output()

		if err != nil {
			if _, ok := err.(*exec.ExitError); !ok {
				log.Warn("Error executing find command in %s: %v. Skipping directory.", dirPath, err)
				done++
				if progress != nil {
					progress(done, total)
				}
				continue
			}
		}

		if len(output) == 0 {
			log.Verbose(isVerbose, "No movie files found in: %s", dirPath)
			nonMovieFolders = append(nonMovieFolders, dirPath)
		} else {
			log.Verbose(isVerbose, "Movie file(s) found in %s.", dirPath)
		}

		done++
		if progress != nil {
			progress(done, total)
		}
	} // End of loop through directories

	return nonMovieFolders, nil
}
