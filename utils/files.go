package utils

import (
	"io"
	"os"
	"path/filepath"

	"github.com/eng618/eng/utils/log"
)

// SyncDirectory checks the source directory, and recursively syncs it's files to the destination directory.
// It checks the file last modification time to verify if the file has been modified,
// and copies the newer file from the source to the destination.
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

// copyFile copies the src file path to the destination file path.
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
