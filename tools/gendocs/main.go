package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"github.com/eng618/eng/cmd"
	"github.com/eng618/eng/internal/ui/theme"
)

func disableAutoGenTag(c *cobra.Command) {
	c.DisableAutoGenTag = true
	for _, child := range c.Commands() {
		disableAutoGenTag(child)
	}
}

func main() {
	outputDir := "./docs/reference/cli"
	if len(os.Args) > 1 {
		outputDir = os.Args[1]
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory %s: %v\n", outputDir, err)
		os.Exit(1)
	}

	// Create temporary directory for fresh generation
	tmpDir, err := os.MkdirTemp("", "eng-gendocs-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating temp directory: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	rootCmd := cmd.GetRootCommand()
	disableAutoGenTag(rootCmd)

	if err := doc.GenMarkdownTree(rootCmd, tmpDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating markdown documentation: %v\n", err)
		os.Exit(1)
	}

	// Compare generated files with outputDir and only write when content actually changed
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading generated docs: %v\n", err)
		os.Exit(1)
	}

	updatedCount := 0
	createdCount := 0
	unchangedCount := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		genFilePath := filepath.Join(tmpDir, entry.Name())
		destFilePath := filepath.Join(outputDir, entry.Name())

		newContent, err := os.ReadFile(genFilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", genFilePath, err)
			continue
		}

		existingContent, err := os.ReadFile(destFilePath)
		if err == nil && bytes.Equal(newContent, existingContent) {
			// Content is identical: do not touch the destination file so mtime/git status is preserved
			unchangedCount++
			continue
		}

		if os.IsNotExist(err) {
			createdCount++
		} else {
			updatedCount++
		}

		if err := os.WriteFile(destFilePath, newContent, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", destFilePath, err)
			os.Exit(1)
		}
	}

	absPath, _ := filepath.Abs(outputDir)
	theme.SuccessMessage(fmt.Sprintf("Documentation synchronized at: %s (Created: %d, Updated: %d, Unchanged: %d)",
		absPath, createdCount, updatedCount, unchangedCount))
}
