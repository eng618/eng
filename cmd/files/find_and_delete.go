package files

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/log"
)

// FileTypeCategory represents a category of file types with their extensions
type FileTypeCategory struct {
	Name       string
	Extensions map[string]bool
	Options    []string // For display in the survey
}

// FindAndDeleteCmd scans a directory for selected file types and deletes them after confirmation.
var (
	globPattern    string
	extension      string
	listExtensions bool
)

var FindAndDeleteCmd = &cobra.Command{
	Use:   "findAndDelete [directory]",
	Short: "Find and delete files of selected types, or list extensions",
	Long: `Recursively scan the provided directory for files of types selected by the user
and delete them after an interactive confirmation. Use --list-extensions to list
all file extensions in the directory instead.
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dir := args[0]
		isVerbose := utils.IsVerbose(cmd)

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			log.Error("Provided directory does not exist: %s", dir)
			return
		}

		if listExtensions {
			extensions, err := ListExtensions(dir)
			if err != nil {
				log.Error("Error listing extensions: %v", err)
				return
			}
			log.Message("File extensions found in %s:", dir)
			for _, ext := range extensions {
				log.Message("  - %s", ext)
			}
			return
		}

		matchFn, err := buildMatchFunction(globPattern, extension)
		if err != nil {
			log.Error("Error building match function: %v", err)
			return
		}
		if matchFn == nil {
			// Prompt the user for file types to search for (multi-select).
			options := []string{
				"JSON files (.json)",
				"Video files (.mp4, .mov, .avi, .mkv)",
				"Image files (.jpg, .jpeg, .png, .gif)",
				"System files (.DS_Store)",
			}
			var selected []string
			prompt := &survey.MultiSelect{
				Message: "Select file types to find and delete:",
				Options: options,
			}
			if err := survey.AskOne(prompt, &selected); err != nil {
				log.Error("Error collecting selection: %v", err)
				return
			}

			if len(selected) == 0 {
				log.Info("No file types selected. Nothing to do.")
				return
			}

			// Build a lookup for extensions based on selection.
			extLookup := map[string]bool{}
			for _, s := range selected {
				s = strings.ToLower(s)
				if strings.Contains(s, "json") {
					extLookup[".json"] = true
				}
				if strings.Contains(s, "video") {
					extLookup[".mp4"] = true
					extLookup[".mov"] = true
					extLookup[".avi"] = true
					extLookup[".mkv"] = true
				}
				if strings.Contains(s, "image") {
					extLookup[".jpg"] = true
					extLookup[".jpeg"] = true
					extLookup[".png"] = true
					extLookup[".gif"] = true
				}
				if strings.Contains(s, "system") {
					extLookup[".ds_store"] = true
				}
			}
			matchFn = func(name string) bool {
				return extLookup[strings.ToLower(filepath.Ext(name))]
			}
		}

		log.Start("Scanning for files...")
		spinner := utils.NewProgressSpinner("Scanning directories...")
		matches, totalSize, walkErr := ScanFiles(dir, matchFn, spinner)

		if walkErr != nil {
			spinner.Stop()
			log.Error("Error scanning directory: %v", walkErr)
			return
		}

		spinner.UpdateMessage("Scan complete.")
		spinner.SetProgressBar(1.0)
		spinner.Stop()

		if walkErr != nil {
			log.Error("Error scanning directory: %v", walkErr)
			return
		}

		if len(matches) == 0 {
			log.Success("No matching files found in %s.", dir)
			return
		}

		log.Message("\nFound %d file(s) (%.2f MB total):", len(matches), float64(totalSize)/(1024*1024))
		for _, m := range matches {
			if isVerbose {
				log.Message("  - %s", m)
			} else {
				log.Message("  - %s", filepath.Base(m))
			}
		}

		// Confirm deletion
		confirm := false
		promptConfirm := &survey.Confirm{
			Message: fmt.Sprintf("Delete %d file(s) (%.2f MB)?", len(matches), float64(totalSize)/(1024*1024)),
			Default: false,
		}
		if err := survey.AskOne(promptConfirm, &confirm); err != nil {
			log.Error("Error during confirmation prompt: %v", err)
			return
		}
		if !confirm {
			log.Info("Deletion cancelled by user.")
			return
		}

		// Attempt parallel deletion
		var deleted, errors atomic.Int64
		var wg sync.WaitGroup
		workerCount := 4 // Adjust based on system capabilities
		fileChan := make(chan string, len(matches))

		// Start workers
		for i := 0; i < workerCount; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for f := range fileChan {
					if f == "" { // Skip empty strings if any
						continue
					}
					if err := os.Remove(f); err != nil {
						if !os.IsNotExist(err) { // Only log if it's not a "file not found" error
							log.Error("Failed to delete %s: %v", f, err)
							errors.Add(1)
						}
					} else {
						if isVerbose {
							log.Success("Deleted %s", f)
						}
						deleted.Add(1)
					}
				}
			}()
		}

		// Feed files to workers
		for _, f := range matches {
			fileChan <- f
		}
		close(fileChan)

		// Wait for completion
		wg.Wait()

		log.Success("Done. Deleted %d file(s), %d errors.", deleted.Load(), errors.Load())
	},
}

// ScanFiles walks dir recursively and returns files that match the provided function.
// Also returns the total size of matched files. spinner may be nil.
// buildMatchFunction creates a file matching function based on provided patterns
func buildMatchFunction(globPattern, extension string) (func(name string) bool, error) {
	if globPattern != "" {
		pattern := globPattern
		// Validate pattern before returning the function
		if _, err := filepath.Match(pattern, "test"); err != nil {
			return nil, fmt.Errorf("invalid glob pattern '%s': %v", pattern, err)
		}
		return func(name string) bool {
			matched, _ := filepath.Match(pattern, filepath.Base(name))
			return matched
		}, nil
	}

	if extension != "" {
		ext := strings.ToLower(extension)
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		return func(name string) bool {
			return strings.ToLower(filepath.Ext(name)) == ext
		}, nil
	}

	return nil, nil
}

// deleteFiles deletes the given files in parallel and returns counts of successes and errors
func deleteFiles(files []string, isVerbose bool) (deleted, errors int64) {
	var wg sync.WaitGroup
	var deletedCount, errorCount atomic.Int64
	workerCount := 4 // Adjust based on system capabilities
	fileChan := make(chan string, len(files))

	// Start workers
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for f := range fileChan {
				if f == "" { // Skip empty strings if any
					continue
				}
				if err := os.Remove(f); err != nil {
					if !os.IsNotExist(err) { // Only log if it's not a "file not found" error
						log.Error("Failed to delete %s: %v", f, err)
						errorCount.Add(1)
					}
				} else {
					if isVerbose {
						log.Success("Deleted %s", f)
					}
					deletedCount.Add(1)
				}
			}
		}()
	}

	// Feed files to workers
	for _, f := range files {
		fileChan <- f
	}
	close(fileChan)

	// Wait for completion
	wg.Wait()

	return deletedCount.Load(), errorCount.Load()
}

func ScanFiles(dir string, matchFn func(name string) bool, spinner *utils.Spinner) ([]string, int64, error) {
	var matches []string
	var totalSize int64
	var filesProcessed, totalFiles atomic.Int64

	// First, count total files for progress reporting
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			log.Warn("Error accessing path %s: %v", path, err)
			return filepath.SkipDir
		}
		if !d.IsDir() {
			totalFiles.Add(1)
		}
		return nil
	})

	if err != nil {
		return nil, 0, fmt.Errorf("error counting files: %w", err)
	}

	// Now do the actual scanning
	walkErr := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// log and continue
			log.Warn("Error accessing path %s: %v", path, err)
			return nil
		}

		if err != nil {
			log.Warn("Error accessing path %s: %v", path, err)
			return filepath.SkipDir
		}

		if !d.IsDir() {
			filesProcessed.Add(1)
			if spinner != nil && filesProcessed.Load()%50 == 0 {
				progress := float64(filesProcessed.Load()) / float64(totalFiles.Load())
				spinner.SetProgressBar(progress, fmt.Sprintf("Scanning... (%d/%d files, %d matches)",
					filesProcessed.Load(), totalFiles.Load(), len(matches)))
			}
		}

		if d.IsDir() {
			return nil
		}

		if matchFn(d.Name()) {
			info, err := d.Info()
			if err != nil {
				log.Warn("Error getting file info for %s: %v", path, err)
				return nil
			}
			matches = append(matches, path)
			totalSize += info.Size()
		}
		return nil
	})

	return matches, totalSize, walkErr
}

func ListExtensions(dir string) ([]string, error) {
	extSet := make(map[string]bool)
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			log.Warn("Error accessing path %s: %v", path, err)
			return nil
		}
		if !d.IsDir() {
			ext := strings.ToLower(filepath.Ext(d.Name()))
			if ext != "" {
				extSet[ext] = true
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	var extensions []string
	for ext := range extSet {
		extensions = append(extensions, ext)
	}
	sort.Strings(extensions)
	return extensions, nil
}
