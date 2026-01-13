package utils

import (
	"io"
	"os"
	"path/filepath"

	"github.com/eng618/eng/internal/utils/log"
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
			// Ensure destination subdirectory exists
			if err := os.MkdirAll(destPath, 0o755); err != nil {
				return err
			}
			// Recursively sync subdirectories
			if err := SyncDirectory(srcPath, destPath, isVerbose); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, destPath, isVerbose); err != nil {
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
func copyFile(srcPath, destPath string, isVerbose bool) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := src.Close(); err != nil {
			log.Error("Error closing source file: %v", err)
		}
		log.Verbose(isVerbose, "Closed source file: %s\n", srcPath)
	}()

	dest, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := dest.Close(); err != nil {
			log.Error("Error closing destination file: %v", err)
		}
		log.Verbose(isVerbose, "Closed destination file: %s\n", destPath)
	}()

	_, err = io.Copy(dest, src)
	if err != nil {
		return err
	}

	return nil
}
