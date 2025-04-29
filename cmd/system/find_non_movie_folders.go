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
	Use:   "findNonMovieFolders",
	Short: "Find and optionally delete non-movie folders",
	Long:  `This command searches recursively through the supplied directory for directories that do not contain video files. It can also delete these directories if the dry-run flag is not set.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Warn("You need to supply the directory to search for non-movie folders.")
			return
		}

		log.Start("Scanning for non-movie folders...")

		directory := args[0]
		isDryRun, _ := cmd.Flags().GetBool("dry-run")
		isVerbose := utils.IsVerbose(cmd)

		log.Verbose(isVerbose, "Searching for directories in: %s", directory)

		spinner := utils.NewProgressSpinner("Scanning for non-movie folders...")
		spinner.Start()
		defer spinner.Stop()

		nonMovieFolders, err := findNonMovieFolders(isVerbose, directory, func(done, total int) {
			spinner.SetProgressBar(float64(done)/float64(total), fmt.Sprintf("Scanning... (%d/%d)", done, total))
		})
		spinner.Stop()
		if err != nil {
			log.Error("Error finding non-movie folders: %s", err)
			return
		}

		log.Verbose(isVerbose, "Found non-movie folders: %s", strings.Join(nonMovieFolders, ", "))

		for _, folder := range nonMovieFolders {
			// No movie files found, log all files that would be deleted
			listFilesCmd := exec.Command("find", folder, "-type", "f")
			filesToDelete, err := listFilesCmd.Output()
			if err != nil {
				log.Error("Error listing files in folder %s: %s", folder, err)
				continue
			}

			if isVerbose {
				for _, file := range strings.Split(string(filesToDelete), "\n") {
					if file != "" {
						log.Warn("Would delete file: %s", file)
					}
				}
				log.Warn("Would delete folder: %s", folder)
			}

			if isDryRun {
				log.Message("Dry-run: Found non-movie folder: %s", folder)
			} else {
				log.Message("Deleting non-movie folder: %s", folder)
				deleteCmd := exec.Command("rm", "-rf", folder)
				if err := deleteCmd.Run(); err != nil {
					log.Error("Error deleting folder %s: %s", folder, err)
				}
			}
		}
		log.Success("Scan complete.")
	},
}

// findNonMovieFolders returns a list of top-level directories under rootDir that do not contain any movie files anywhere inside them.
func findNonMovieFolders(isVerbose bool, rootDir string, progress func(done, total int)) ([]string, error) {
	var nonMovieFolders []string

	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, err
	}

	total := 0
	for _, entry := range entries {
		if entry.IsDir() {
			total++
		}
	}
	done := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirPath := filepath.Join(rootDir, entry.Name())
		log.Verbose(isVerbose, "Checking top-level directory: %s", dirPath)

		// Recursively search for movie files in this directory
		checkCmd := exec.Command("find", dirPath, "-type", "f", "(",
			"-iname", "*.mp4", "-o",
			"-iname", "*.mkv", "-o",
			"-iname", "*.avi", "-o",
			"-iname", "*.mov", "-o",
			"-iname", "*.wmv", "-o",
			"-iname", "*.flv", "-o",
			"-iname", "*.webm", "-o",
			"-iname", "*.mpeg", "-o",
			"-iname", "*.mpg",
			")")
		files, err := checkCmd.Output()
		if err != nil {
			log.Verbose(isVerbose, "Error checking for movie files in %s: %v", dirPath, err)
			continue
		}
		if len(strings.TrimSpace(string(files))) == 0 {
			log.Verbose(isVerbose, "No movie files found in: %s", dirPath)
			nonMovieFolders = append(nonMovieFolders, dirPath)
		} else {
			log.Verbose(isVerbose, "Movie files found in: %s", dirPath)
		}
		done++
		if progress != nil {
			progress(done, total)
		}
	}
	return nonMovieFolders, nil
}
