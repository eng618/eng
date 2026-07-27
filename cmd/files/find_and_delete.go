package files

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

// FileTypeCategory represents a category of file types with their extensions.
type FileTypeCategory struct {
	Name       string
	Extensions map[string]bool
	Options    []string // For display in the survey
}

// FileMatch tracks a matched file path and its size in bytes.
type FileMatch struct {
	Path      string
	SizeBytes int64
}

// FormatLabel returns a user-friendly display string with path and human-readable size.
func (f FileMatch) FormatLabel() string {
	if f.SizeBytes > 0 {
		return fmt.Sprintf("%s (%s)", f.Path, humanize.Bytes(uint64(f.SizeBytes)))
	}
	return f.Path
}

// FindAndDeleteCmd scans a directory for selected file types and deletes them after confirmation.
var (
	globPattern    string
	extension      string
	filename       string
	listExtensions bool
)

var FindAndDeleteCmd = &cobra.Command{
	Use:     "findAndDelete [directory]",
	Aliases: []string{"find-and-delete", "delete-files", "clean-files"},
	Short:   "Find and delete files of selected types, or list extensions",
	Long: `Recursively scan the provided directory for files of types selected by the user
and delete them after an interactive confirmation. Use --list-extensions to list
all file extensions in the directory instead. Use --filename to target a specific
filename, --glob for glob patterns, or --ext for file extensions.
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary).
			MarginBottom(1)
		if !ui.DisableProgress {
			fmt.Println(headerStyle.Render("🧹 Find & Delete Files"))
		}

		dir := args[0]
		isVerbose := cmdutil.IsVerbose(cmd)

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

			var extLines []string
			extLines = append(extLines, fmt.Sprintf("Found %s file extension(s) in %s:",
				theme.PrimaryText.Bold(true).Render(fmt.Sprintf("%d", len(extensions))),
				theme.BoldText.Render(dir),
			))
			for _, ext := range extensions {
				extLines = append(extLines, fmt.Sprintf("  • %s", theme.PrimaryText.Render(ext)))
			}

			if !ui.DisableProgress {
				fmt.Println(theme.InfoBox.Render(strings.Join(extLines, "\n")))
			} else {
				log.Message("File extensions found in %s:", dir)
				for _, ext := range extensions {
					log.Message("  - %s", ext)
				}
			}
			return
		}

		matchFn, err := buildMatchFunction(globPattern, extension, filename)
		if err != nil {
			log.Error("Error building match function: %v", err)
			return
		}
		if matchFn == nil {
			options := []string{
				"JSON files (.json)",
				"Video files (.mp4, .mov, .avi, .mkv, .m4v, .wmv, .3gp)",
				"Image files (.jpg, .jpeg, .png, .gif)",
				"Microsoft documents (.doc, .docx, .xls, .xlsx, .ppt, .pptx)",
				"Archive files (.zip, .rar, .7z, .tar, .gz, .bz2)",
				"Audio files (.mp3, .wav, .flac, .aac, .ogg)",
				"PDF documents (.pdf)",
				"Text files (.txt, .md, .rtf)",
				"Log files (.log)",
				"Temporary files (.tmp, .temp, .swp)",
				"Backup files (.bak, .backup, .old)",
				"Executable files (.exe, .msi, .dmg, .pkg, .deb, .rpm)",
				"System files (.DS_Store)",
			}
			selected, err := ui.MultiSelect("Select file types to find and delete:", options, nil)
			if err != nil {
				log.Error("Error collecting selection: %v", err)
				return
			}

			if len(selected) == 0 {
				log.Info("No file types selected. Nothing to do.")
				return
			}

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
					extLookup[".m4v"] = true
					extLookup[".wmv"] = true
					extLookup[".3gp"] = true
				}
				if strings.Contains(s, "image") {
					extLookup[".jpg"] = true
					extLookup[".jpeg"] = true
					extLookup[".png"] = true
					extLookup[".gif"] = true
				}
				if strings.Contains(s, "microsoft") {
					extLookup[".doc"] = true
					extLookup[".docx"] = true
					extLookup[".xls"] = true
					extLookup[".xlsx"] = true
					extLookup[".ppt"] = true
					extLookup[".pptx"] = true
				}
				if strings.Contains(s, "archive") {
					extLookup[".zip"] = true
					extLookup[".rar"] = true
					extLookup[".7z"] = true
					extLookup[".tar"] = true
					extLookup[".gz"] = true
					extLookup[".bz2"] = true
				}
				if strings.Contains(s, "audio") {
					extLookup[".mp3"] = true
					extLookup[".wav"] = true
					extLookup[".flac"] = true
					extLookup[".aac"] = true
					extLookup[".ogg"] = true
				}
				if strings.Contains(s, "pdf") {
					extLookup[".pdf"] = true
				}
				if strings.Contains(s, "text") {
					extLookup[".txt"] = true
					extLookup[".md"] = true
					extLookup[".rtf"] = true
				}
				if strings.Contains(s, "log") {
					extLookup[".log"] = true
				}
				if strings.Contains(s, "temporary") || strings.Contains(s, "temp") {
					extLookup[".tmp"] = true
					extLookup[".temp"] = true
					extLookup[".swp"] = true
				}
				if strings.Contains(s, "backup") {
					extLookup[".bak"] = true
					extLookup[".backup"] = true
					extLookup[".old"] = true
				}
				if strings.Contains(s, "executable") {
					extLookup[".exe"] = true
					extLookup[".msi"] = true
					extLookup[".dmg"] = true
					extLookup[".pkg"] = true
					extLookup[".deb"] = true
					extLookup[".rpm"] = true
				}
				if strings.Contains(s, "system") {
					extLookup[".ds_store"] = true
				}
			}
			matchFn = func(name string) bool {
				return extLookup[strings.ToLower(filepath.Ext(name))]
			}
		}

		var spinner *ui.Spinner
		if !ui.DisableProgress {
			spinner = ui.NewProgressSpinner("Scanning directories for matching files...")
		}

		fileMatches, totalSize, walkErr := ScanFileMatches(dir, matchFn, spinner)
		if spinner != nil {
			spinner.Stop()
		}

		if walkErr != nil {
			log.Error("Error scanning directory: %v", walkErr)
			return
		}

		if len(fileMatches) == 0 {
			theme.SuccessMessage(fmt.Sprintf("No matching files found in %s.", dir))
			return
		}

		sizeFormatted := humanize.Bytes(uint64(totalSize))

		// Render Callout Box
		var boxLines []string
		boxLines = append(boxLines, fmt.Sprintf("Found %s matching file(s) reclaiming %s space in %s:",
			theme.PrimaryText.Bold(true).Render(fmt.Sprintf("%d", len(fileMatches))),
			theme.SuccessText.Bold(true).Render(sizeFormatted),
			theme.BoldText.Render(dir),
		))

		if !ui.DisableProgress {
			fmt.Println(theme.InfoBox.Render(strings.Join(boxLines, "\n")))
		}

		// Prepare MultiSelect options with itemized file sizes
		labelMap := make(map[string]FileMatch)
		var options []string
		for _, fm := range fileMatches {
			label := fm.FormatLabel()
			labelMap[label] = fm
			options = append(options, label)
		}

		confirmPromptMsg := fmt.Sprintf("Select files to delete (%s total matched):", sizeFormatted)
		selectedLabels, err := ui.MultiSelect(confirmPromptMsg, options, options)
		if err != nil {
			log.Error("Error during confirmation prompt: %v", err)
			return
		}

		if len(selectedLabels) == 0 {
			log.Info("Deletion canceled or no files selected.")
			return
		}

		var selectedFiles []FileMatch
		var selectedTotalBytes int64
		for _, label := range selectedLabels {
			if fm, ok := labelMap[label]; ok {
				selectedFiles = append(selectedFiles, fm)
				selectedTotalBytes += fm.SizeBytes
			}
		}

		totalToDelete := len(selectedFiles)
		var deleteSpinner *ui.Spinner
		if !ui.DisableProgress {
			deleteSpinner = ui.NewProgressSpinner(fmt.Sprintf("Deleting %d selected file(s)...", totalToDelete))
		}

		var deletedCount int
		var freedBytes int64
		var errorCount int

		for i, fm := range selectedFiles {
			ratio := float64(i+1) / float64(totalToDelete)
			statusMsg := fmt.Sprintf("[%d/%d] Deleting %s", i+1, totalToDelete, filepath.Base(fm.Path))

			if deleteSpinner != nil {
				deleteSpinner.SetProgressBar(ratio, statusMsg)
			}

			if err := os.Remove(fm.Path); err != nil && !os.IsNotExist(err) {
				errorCount++
				if deleteSpinner != nil {
					deleteSpinner.Logf("  %s Failed %s: %v\n", theme.ErrorText.Render("✗"), fm.Path, err)
				} else {
					log.Error("Failed to delete %s: %v", fm.Path, err)
				}
			} else {
				deletedCount++
				freedBytes += fm.SizeBytes
				sizeStr := ""
				if fm.SizeBytes > 0 {
					sizeStr = fmt.Sprintf(" (freed %s)", humanize.Bytes(uint64(fm.SizeBytes)))
				}
				if deleteSpinner != nil {
					deleteSpinner.Logf("  %s Deleted %s%s\n", theme.SuccessText.Render("✓"), fm.Path, sizeStr)
				} else {
					if isVerbose {
						log.Success("Deleted %s%s", fm.Path, sizeStr)
					}
				}
			}
		}

		if deleteSpinner != nil {
			deleteSpinner.SetProgressBar(1.0, "Deletion complete")
			deleteSpinner.Stop()
		}

		freedStr := humanize.Bytes(uint64(freedBytes))
		if errorCount > 0 {
			theme.WarningMessage(
				fmt.Sprintf(
					"Deleted %d file(s) freeing %s, but encountered %d error(s).",
					deletedCount,
					freedStr,
					errorCount,
				),
			)
		} else {
			theme.SuccessMessage(
				fmt.Sprintf("Successfully deleted %d file(s) freeing %s of disk space!", deletedCount, freedStr),
			)
		}
	},
}

// deleteFiles deletes the given files in parallel and returns counts of successes and errors.
// Maintained for backward compatibility with unit tests.
func deleteFiles(files []string, isVerbose bool) (deleted, errors int64) {
	var wg sync.WaitGroup
	var deletedCount, errorCount atomic.Int64
	workerCount := 4
	fileChan := make(chan string, len(files))

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for f := range fileChan {
				if f == "" {
					continue
				}
				if err := os.Remove(f); err != nil {
					if !os.IsNotExist(err) {
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

	for _, f := range files {
		fileChan <- f
	}
	close(fileChan)
	wg.Wait()

	return deletedCount.Load(), errorCount.Load()
}

// ScanFileMatches walks dir recursively and returns FileMatch structs containing path and size.
func ScanFileMatches(dir string, matchFn func(name string) bool, spinner *ui.Spinner) ([]FileMatch, int64, error) {
	var matches []FileMatch
	var totalSize int64
	var filesProcessed int

	walkErr := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d == nil {
			return nil
		}

		if d.IsDir() {
			return nil
		}

		filesProcessed++
		if spinner != nil && filesProcessed%50 == 0 {
			spinner.SetProgressBar(
				0.5,
				fmt.Sprintf("Scanning files... (%d processed, %d matched)", filesProcessed, len(matches)),
			)
		}

		if matchFn(d.Name()) {
			info, err := d.Info()
			var sz int64
			if err == nil {
				sz = info.Size()
			}
			matches = append(matches, FileMatch{
				Path:      path,
				SizeBytes: sz,
			})
			totalSize += sz
		}
		return nil
	})

	return matches, totalSize, walkErr
}

// ScanFiles walks dir recursively and returns files that match the provided function.
// Maintained for backward compatibility.
func ScanFiles(dir string, matchFn func(name string) bool, spinner *ui.Spinner) ([]string, int64, error) {
	fileMatches, totalSize, err := ScanFileMatches(dir, matchFn, spinner)
	var paths []string
	for _, fm := range fileMatches {
		paths = append(paths, fm.Path)
	}
	return paths, totalSize, err
}

func buildMatchFunction(globPattern, extension, filename string) (func(name string) bool, error) {
	if globPattern != "" {
		pattern := globPattern
		if _, err := filepath.Match(pattern, "test"); err != nil {
			return nil, fmt.Errorf("invalid glob pattern '%s': %w", pattern, err)
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

	if filename != "" {
		return func(name string) bool {
			return name == filename
		}, nil
	}

	return nil, nil
}

func ListExtensions(dir string) ([]string, error) {
	extSet := make(map[string]bool)
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d == nil {
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
