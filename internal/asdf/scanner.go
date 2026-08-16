package asdf

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// DefaultIgnoredDirs is a list of directory names to skip during recursive scanning for .tool-versions.
var DefaultIgnoredDirs = []string{
	".git",
	".github",
	".vscode",
	".idea",
	"node_modules",
	"vendor",
	".venv",
	"venv",
	"target",
	"dist",
	"build",
	".cache",
	".asdf",
	".cargo",
}

// FindToolVersionFiles recursively scans the given root directories for .tool-versions files,
// skipping common build, dependency, and VCS directories.
func FindToolVersionFiles(rootDirs []string) ([]string, error) {
	return FindToolVersionFilesWithProgress(rootDirs, nil)
}

// FindToolVersionFilesWithProgress recursively scans the given root directories for .tool-versions files,
// invoking onProgress (if non-nil) whenever a directory is inspected or a file is found.
func FindToolVersionFilesWithProgress(
	rootDirs []string,
	onProgress func(currentDir string, foundCount int),
) ([]string, error) {
	var foundFiles []string
	seenPaths := make(map[string]bool)

	for _, root := range rootDirs {
		if root == "" {
			continue
		}

		absRoot, err := filepath.Abs(root)
		if err != nil {
			absRoot = root
		}

		info, err := os.Stat(absRoot)
		if err != nil || !info.IsDir() {
			// If root is a file named .tool-versions, record it
			if err == nil && !info.IsDir() && filepath.Base(absRoot) == ".tool-versions" {
				if !seenPaths[absRoot] {
					seenPaths[absRoot] = true
					foundFiles = append(foundFiles, absRoot)
					if onProgress != nil {
						onProgress(absRoot, len(foundFiles))
					}
				}
			}
			continue
		}

		if onProgress != nil {
			onProgress(absRoot, len(foundFiles))
		}

		err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil || d == nil {
				return nil
			}

			if d.IsDir() {
				dirName := d.Name()
				// Skip root dir check
				if path != absRoot && slices.Contains(DefaultIgnoredDirs, dirName) {
					return filepath.SkipDir
				}
				if onProgress != nil && path != absRoot {
					onProgress(path, len(foundFiles))
				}
				return nil
			}

			if d.Name() == ".tool-versions" {
				if !seenPaths[path] {
					seenPaths[path] = true
					foundFiles = append(foundFiles, path)
					if onProgress != nil {
						onProgress(path, len(foundFiles))
					}
				}
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("error searching for .tool-versions in %s: %w", absRoot, err)
		}
	}

	return foundFiles, nil
}

// FileProtectionSummary tracks versions protected by a specific .tool-versions file.
type FileProtectionSummary struct {
	FilePath  string
	Protected ToolVersions
}

// ParseAndMergeToolVersions parses a list of .tool-versions files and merges
// all protected versions into a single ToolVersions map. It also returns per-file protection summaries.
func ParseAndMergeToolVersions(filePaths []string) (ToolVersions, []FileProtectionSummary, error) {
	merged := make(ToolVersions)
	var summaries []FileProtectionSummary

	for _, path := range filePaths {
		tv, err := ParseToolVersionsFile(path)
		if err != nil {
			// Skip unreadable files gracefully
			continue
		}

		if len(tv) == 0 {
			continue
		}

		summaries = append(summaries, FileProtectionSummary{
			FilePath:  path,
			Protected: tv,
		})

		for plugin, versions := range tv {
			for _, v := range versions {
				if !slices.Contains(merged[plugin], v) {
					merged[plugin] = append(merged[plugin], v)
				}
			}
		}
	}

	return merged, summaries, nil
}

// FormatFileSummary returns a formatted label describing protected versions in a file.
func (s FileProtectionSummary) FormatFileSummary(homeDir string) string {
	displayPath := s.FilePath
	if homeDir != "" && strings.HasPrefix(s.FilePath, homeDir) {
		displayPath = "~" + strings.TrimPrefix(s.FilePath, homeDir)
	}

	var items []string
	for plugin, versions := range s.Protected {
		items = append(items, fmt.Sprintf("%s @ %s", plugin, strings.Join(versions, ", ")))
	}

	return fmt.Sprintf("%s (%s)", displayPath, strings.Join(items, "; "))
}
