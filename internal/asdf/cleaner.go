package asdf

import (
	"os"
	"path/filepath"
	"slices"
)

// GetASDFDataDir resolves the location of the asdf data directory.
func GetASDFDataDir(homeDir string) string {
	if envDir := os.Getenv("ASDF_DATA_DIR"); envDir != "" {
		return envDir
	}
	dotAsdf := filepath.Join(homeDir, ".asdf")
	if _, err := os.Stat(dotAsdf); err == nil {
		return dotAsdf
	}
	return filepath.Join(homeDir, ".local", "share", "asdf")
}

// CalculateDirSize recursively calculates the size in bytes of a directory,
// ignoring symbolic links to prevent double counting.
func CalculateDirSize(dirPath string) int64 {
	var totalSize int64
	_ = filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
		if err != nil || d == nil {
			return nil
		}
		// Skip symlinks
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err == nil {
				totalSize += info.Size()
			}
		}
		return nil
	})
	return totalSize
}

// FilterRemovableVersions compares installed versions against protected versions in .tool-versions
// and returns a list of CleanupTargets eligible for removal.
// If targetPlugin is non-empty, only versions for that specific plugin are returned.
// If asdfDataDir is non-empty, disk size for each removable target will be calculated.
func FilterRemovableVersions(installed map[string][]string, protected ToolVersions, targetPlugin string, asdfDataDir string) []CleanupTarget {
	var targets []CleanupTarget

	for plugin, versions := range installed {
		// If a target plugin was specified, skip other plugins
		if targetPlugin != "" && plugin != targetPlugin {
			continue
		}

		protectedSet := protected[plugin]

		for _, ver := range versions {
			// If version is protected in .tool-versions, skip it
			if slices.Contains(protectedSet, ver) {
				continue
			}

			var size int64
			if asdfDataDir != "" {
				installDir := filepath.Join(asdfDataDir, "installs", plugin, ver)
				size = CalculateDirSize(installDir)
			}

			targets = append(targets, CleanupTarget{
				Plugin:    plugin,
				Version:   ver,
				SizeBytes: size,
			})
		}
	}

	return targets
}
