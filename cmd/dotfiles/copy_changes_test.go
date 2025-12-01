package dotfiles

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestCopyChangesCmd_MissingConfig(t *testing.T) {
	// Ensure viper has no config for dotfiles
	viper.Reset()

	// Run command; should return early without panic
	cmd := &cobra.Command{}
	CopyChangesCmd.Run(cmd, []string{})
}

func TestGetModifiedFiles(t *testing.T) {
	viper.Reset()
	viper.Set("dotfiles.repoPath", "/tmp/repo")
	viper.Set("dotfiles.worktree", "/tmp/worktree")

	// Mock the function to return some files
	original := getModifiedFilesFunc
	getModifiedFilesFunc = func(repoPath, worktreePath string) ([]string, error) {
		return []string{".tool-versions", ".zshrc"}, nil
	}
	defer func() { getModifiedFilesFunc = original }()

	files, err := getModifiedFilesFunc("/tmp/repo", "/tmp/worktree")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
	if files[0] != ".tool-versions" {
		t.Fatalf("expected .tool-versions, got %s", files[0])
	}
}

func TestCopyFile(t *testing.T) {
	// Create temp files
	srcDir := t.TempDir()
	destDir := t.TempDir()

	srcFile := filepath.Join(srcDir, "test.txt")
	destFile := filepath.Join(destDir, "test.txt")

	content := "test content"
	if err := os.WriteFile(srcFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write src file: %v", err)
	}

	if err := copyFile(srcFile, destFile, false); err != nil {
		t.Fatalf("failed to copy file: %v", err)
	}

	destContent, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("failed to read dest file: %v", err)
	}

	if string(destContent) != content {
		t.Fatalf("content mismatch: got %s, want %s", destContent, content)
	}
}
