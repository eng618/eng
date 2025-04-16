// Package utils_test contains unit tests for the utils package.
package utils_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/eng618/eng/utils"
)

func TestSyncDirectory_CopiesFilesAndSubdirs(t *testing.T) {
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "src")
	destDir := filepath.Join(dir, "dest")

	// Ensure destination directory exists
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		t.Fatalf("failed to create destDir: %v", err)
	}

	// Create source directory structure
	if err := os.MkdirAll(filepath.Join(srcDir, "subdir"), 0o755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("failed to write file1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "subdir", "file2.txt"), []byte("world"), 0o644); err != nil {
		t.Fatalf("failed to write file2: %v", err)
	}

	// Run SyncDirectory
	if err := utils.SyncDirectory(srcDir, destDir, false); err != nil {
		t.Fatalf("SyncDirectory failed: %v", err)
	}

	// Check that files exist in destDir
	b, err := os.ReadFile(filepath.Join(destDir, "file1.txt"))
	if err != nil || string(b) != "hello" {
		t.Errorf("file1.txt not copied correctly: %v, content: %s", err, string(b))
	}
	b, err = os.ReadFile(filepath.Join(destDir, "subdir", "file2.txt"))
	if err != nil || string(b) != "world" {
		t.Errorf("file2.txt not copied correctly: %v, content: %s", err, string(b))
	}
}
