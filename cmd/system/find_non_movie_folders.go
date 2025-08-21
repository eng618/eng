package system

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
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

It lists the files within the identified folders and prompts for confirmation before deletion.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		log.Start("Scanning for non-movie folders...")

		directory := args[0]
		isVerbose := utils.IsVerbose(cmd)

		// Validate directory exists
		if _, err := os.Stat(directory); os.IsNotExist(err) {
			log.Error("Provided directory does not exist: %s", directory)
			return
		}

		log.Verbose(isVerbose, "Searching for directories in: %s", directory)
		spinner := utils.NewProgressSpinner("Scanning directories...")

		nonMovieFolders, err := findNonMovieFolders(isVerbose, directory, spinner, func(done, total int) {
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

		log.Verbose(isVerbose, "Found %d potential non-movie folders.", len(nonMovieFolders))

		if len(nonMovieFolders) == 0 {
			log.Success("No non-movie folders found in %s.", directory)
			return
		}

		log.Message("\nFound %d non-movie folder(s) that will be deleted:", len(nonMovieFolders))

		// Store folder contents for summary and confirmation
		type FolderContent struct {
			Files []string
			Error string
		}
		folderContents := make(map[string]FolderContent)

		for _, folder := range nonMovieFolders {
			var files []string
			walkErr := filepath.WalkDir(folder, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !d.IsDir() {
					relPath, err := filepath.Rel(folder, path)
					if err != nil {
						return err
					}
					files = append(files, relPath)
				}
				return nil
			})

			listErrorString := ""
			if walkErr != nil {
				errMsg := fmt.Sprintf("Could not list files in folder %s: %s", folder, walkErr)
				log.Warn(errMsg)
				listErrorString = "(error listing files)"
			}
			folderContents[folder] = FolderContent{Files: files, Error: listErrorString}

			log.Message("  - %s", folder)
			if listErrorString != "" {
				log.Message("    %s", listErrorString)
			} else if len(files) > 0 {
				for _, file := range files {
					displayPath := filepath.Join(filepath.Base(folder), file)
					log.Message("    - %s", displayPath)
				}
			} else {
				log.Message("    (Contains no files or only empty subdirectories)")
			}
		}

		log.Message("") // Add a blank line for readability

		if !askForConfirmation("Do you want to delete these folders and their contents?") {
			log.Info("Deletion cancelled by user.")
			return
		}

		log.Start("Deleting non-movie folders...")
		deletedCount := 0
		skippedCount := 0
		errorMessages := []string{}

		for _, folder := range nonMovieFolders {
			log.Warn("Attempting to delete: %s", folder)
			if err := os.RemoveAll(folder); err != nil {
				errMsg := fmt.Sprintf("Error deleting folder %s: %s", folder, err)
				log.Error(errMsg)
				errorMessages = append(errorMessages, errMsg)
				skippedCount++
			} else {
				log.Success("Deleted: %s", folder)
				deletedCount++
			}
		}

		// Final summary
		if len(errorMessages) > 0 {
			log.Warn("Encountered %d error(s) during processing.", len(errorMessages))
		}

		log.Success("Processing complete. Deleted %d folder(s), skipped %d due to errors.", deletedCount, skippedCount)
	},
}

// askForConfirmation prompts the user for a yes/no confirmation using survey.
func askForConfirmation(prompt string) bool {
	confirm := false
	promptConfirm := &survey.Confirm{
		Message: prompt,
		Default: false, // Default to No for safety
	}
	err := survey.AskOne(promptConfirm, &confirm)
	if err != nil {
		// Handle error, e.g., log it and return false for safety
		log.Error("Error during confirmation prompt: %v", err)
		return false
	}
	return confirm
}

// findNonMovieFolders scans the immediate subdirectories of rootDir.
// It returns a slice of paths to subdirectories that do not contain any files
// matching common video extensions (recursively within each subdirectory).
//
// Parameters:
//   - isVerbose: If true, logs verbose messages during the scan.
//   - rootDir: The path to the directory whose subdirectories will be scanned.
//   - spinner: A pointer to the progress spinner.
//   - progress: A callback function to report progress (done, total). Can be nil.
//
// Returns:
//   - A slice of strings, where each string is the absolute path to a non-movie folder.
//   - An error if reading the root directory fails.
func findNonMovieFolders(isVerbose bool, rootDir string, spinner *utils.Spinner, progress func(done, total int)) ([]string, error) {
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

	videoExtensions := map[string]bool{
		".mp4": true, ".mkv": true, ".avi": true, ".mov": true, ".wmv": true,
		".flv": true, ".webm": true, ".mpeg": true, ".mpg": true, ".m4v": true,
	}

	for _, entry := range dirEntries {
		dirPath := filepath.Join(rootDir, entry.Name())
		if isVerbose {
			spinner.Logf("--- Checking directory: %s\n", dirPath)
		}

		foundMovieFile := false
		walkErr := filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err // Propagate errors
			}
			if !d.IsDir() {
				ext := strings.ToLower(filepath.Ext(d.Name()))
				if videoExtensions[ext] {
					foundMovieFile = true
					// Return a special error to stop walking, since we found what we need.
					return filepath.SkipAll
				}
			}
			return nil
		})

		// If walkErr is filepath.SkipAll, it means we found a movie file and stopped early.
		// This is our success condition for finding a movie, not a real error.
		if walkErr != nil && walkErr != filepath.SkipAll {
			log.Warn("Error scanning directory %s: %v. Skipping.", dirPath, walkErr)
		}

		if !foundMovieFile {
			if isVerbose {
				spinner.Logf("--- No movie files found in: %s\n", dirPath)
			}
			nonMovieFolders = append(nonMovieFolders, dirPath)
		} else {
			if isVerbose {
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
