/*
Copyright © 2024 Eric N. Garcia <eng618@garciaericn.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/log"
)

// systemCmd represents the system command
var systemCmd = &cobra.Command{
	Use:   "system",
	Short: "A command for managing the system",
	Long:  `This command will help manage various aspects of MacOS.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("system called")
	},
}

func init() {
	rootCmd.AddCommand(systemCmd)

	systemCmd.AddCommand(killPort)
	systemCmd.AddCommand(findNonMovieFolders)

	// TODO: Ubuntu update command

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// systemCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// systemCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	findNonMovieFolders.Flags().Bool("dry-run", false, "Only print the directories that do not contain movie files")
}

var killPort = &cobra.Command{
	Use:   "killPort",
	Short: "Kill a supplied port",
	Long:  `This will find what process is running on a supplied port, then kill that process.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			KillString := fmt.Sprintf("kill -9 $(lsof -ti:%s)", strings.Join(args, ","))
			log.Message("This should run: %s", KillString)

			killPortCmd := exec.Command(KillString)
			utils.StartChildProcess(killPortCmd)
		} else {
			log.Warn("You need to supply the port to kill, which will run the following:")
		}
	},
}

// findNonMovieFolders is a Cobra command that searches recursively through a specified
// directory for folders that do not contain video files. It provides an option to 
// delete these folders if the dry-run flag is not set.
//
// Usage:
//   findNonMovieFolders <directory> [flags]
//
// Flags:
//   --dry-run   Perform a dry run without deleting any folders.
//
// Description:
// This command uses the `find` utility to identify all directories within the specified
// directory. For each directory, it checks if it contains any video files with common
// extensions such as .mp4, .mkv, .avi, .mov, .wmv, .flv, .webm, .mpeg, or .mpg. If no
// video files are found in a directory, it is considered a "non-movie folder."
//
// If the `--dry-run` flag is set, the command will only log the non-movie folders it finds
// without deleting them. If the flag is not set, the command will delete the identified
// non-movie folders.
//
// Example:
//   # Perform a dry run to find non-movie folders
//   findNonMovieFolders /path/to/directory --dry-run
//
//   # Find and delete non-movie folders
//   findNonMovieFolders /path/to/directory
var findNonMovieFolders = &cobra.Command{
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
