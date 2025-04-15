package system

import (
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
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		verbose := utils.IsVerbose(cmd)

		if verbose {
			log.Verbose(verbose, "Searching for directories in: %s", directory)
		}

		findCmd := exec.Command("find", directory, "-type", "d")
		folders, err := findCmd.Output()
		if err != nil {
			log.Error("Error finding directories: %s", err)
			return
		}

		if verbose {
			log.Verbose(verbose, "Found directories: %s", strings.TrimSpace(string(folders)))
		}

		for _, folder := range strings.Split(string(folders), "\n") {
			if folder == "" {
				continue
			}

			if verbose {
				log.Verbose(verbose, "Checking folder: %s", folder)
			}

			checkCmd := exec.Command("find", folder, "-type", "f", "-iregex", ".*\\.(mp4|mkv|avi|mov|wmv|flv|webm|mpeg|mpg)")
			files, err := checkCmd.Output()
			if err != nil {
				log.Error("Error checking folder %s: %s", folder, err)
				continue
			}

			if verbose {
				log.Verbose(verbose, "Files found in folder %s: %s", folder, strings.TrimSpace(string(files)))
			}

			if strings.TrimSpace(string(files)) == "" {
				if dryRun {
					log.Message("Dry-run: Found non-movie folder: %s", folder)
				} else {
					log.Message("Deleting non-movie folder: %s", folder)
					deleteCmd := exec.Command("rm", "-rf", folder)
					if err := deleteCmd.Run(); err != nil {
						log.Error("Error deleting folder %s: %s", folder, err)
					}
				}
			}
		}
	},
}
