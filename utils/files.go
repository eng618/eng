package utils

import (
	"io"
	"os"
	"path/filepath"

	"github.com/eng618/eng/utils/log"
)

// SyncDirectory synchronizes the contents of the source directory (srcDir) with the destination directory (destDir).
// It copies files and recursively syncs subdirectories. If isVerbose is true, it logs the files being copied.
//
// Parameters:
//   - srcDir: The source directory to sync from.
//   - destDir: The destination directory to sync to.
//   - isVerbose: A boolean flag to enable verbose logging.
//
// Returns:
//   - error: An error if any occurs during the synchronization process., otherwise nil.
func SyncDirectory(srcDir, destDir string, isVerbose bool) error {
	srcFiles, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}

	for _, de := range srcFiles {
		srcPath := filepath.Join(srcDir, de.Name())
		destPath := filepath.Join(destDir, de.Name())

		if de.IsDir() {
			// Recursively sync subdirectories
			if err := SyncDirectory(srcPath, destPath, isVerbose); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, destPath); err != nil {
				return err
			}

			log.Verbose(isVerbose, "Copied %s to %s\n", srcPath, destPath)
		}
	}

	return nil
}

// copyFile copies a file from srcPath to destPath.
// It opens the source file for reading and the destination file for writing.
// If any error occurs during these operations or during the copy process,
// it returns the error. Otherwise, it returns nil.
//
// Parameters:
//   - srcPath: The path to the source file to be copied.
//   - destPath: The path to the destination file where the content will be copied.
//
// Returns:
//   - error: An error if any occurs during the file operations, otherwise nil.
func copyFile(srcPath, destPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dest, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, src)
	if err != nil {
		return err
	}

	return nil
}
