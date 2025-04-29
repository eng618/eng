package system

import (
	"os"
	"os/exec"
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

		directory := args[0]
		isDryRun, _ := cmd.Flags().GetBool("dry-run")
		isVerbose := utils.IsVerbose(cmd)

		log.Verbose(isVerbose, "Searching for directories in: %s", directory)

		nonMovieFolders, err := findNonMovieFolders(isVerbose, directory)
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
	},
}

// findNonMovieFolders returns a list of directories under rootDir that do not contain any movie files.
func findNonMovieFolders(isVerbose bool, rootDir string) ([]string, error) {
	var nonMovieFolders []string

	log.Verbose(isVerbose, "Running: find %s -type d", rootDir)
	findCmd := exec.Command("find", rootDir, "-type", "d")
	combinedOutput, err := findCmd.CombinedOutput()
	outputStr := string(combinedOutput)
	if err != nil {
		log.Verbose(isVerbose, "find command error: %v", err)
		// Continue processing even if find returns error (e.g., missing dirs)
	}

	for _, folder := range strings.Split(outputStr, "\n") {
		if folder == "" || strings.Contains(folder, "No such file or directory") {
			continue
		}
		
		log.Verbose(isVerbose, "Checking folder: %s", folder)
		if _, err := os.Stat(folder); err != nil {
			log.Verbose(isVerbose, "Skipping missing folder: %s (%v)", folder, err)
			continue
		}

		log.Verbose(isVerbose, "Running: find %s -type f (movie extensions)", folder)
		checkCmd := exec.Command("find", folder, "-type", "f", "(",
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
			log.Verbose(isVerbose, "Error checking for movie files in %s: %v", folder, err)
			continue
		}
		if len(strings.TrimSpace(string(files))) == 0 {
			log.Verbose(isVerbose, "No movie files found in: %s", folder)
			nonMovieFolders = append(nonMovieFolders, folder)
		} else {
			log.Verbose(isVerbose, "Movie files found in: %s", folder)
		}
	}
	return nonMovieFolders, nil
}
