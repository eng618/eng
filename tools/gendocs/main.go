package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	checkOnly := false
	pruneStale := false
	for _, arg := range os.Args[1:] {
		switch arg {
		case "--check":
			checkOnly = true
		case "--prune", "--clean":
			pruneStale = true
		default:
			if !strings.HasPrefix(arg, "-") {
				outputDir = arg
			} else {
				fmt.Fprintf(os.Stderr, "Unknown flag %s (want --check or --prune)\n", arg)
				os.Exit(2)
			}
		}
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

	generated := map[string][]byte{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		content, err := os.ReadFile(filepath.Join(tmpDir, entry.Name()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", entry.Name(), err)
			os.Exit(1)
		}
		generated[entry.Name()] = content
	}

	existing := map[string]bool{}
	if dirEntries, err := os.ReadDir(outputDir); err == nil {
		for _, entry := range dirEntries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".md" {
				existing[entry.Name()] = true
			}
		}
	} else if !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error reading output directory %s: %v\n", outputDir, err)
		os.Exit(1)
	}

	// Stale pages: present in the output dir but no longer generated
	// (removed/renamed commands). Report always; delete only with --prune.
	stale := findStale(generated, existing)

	if checkOnly {
		if drift := checkDrift(outputDir, generated, stale); drift > 0 {
			os.Exit(1)
		}
		theme.SuccessMessage("Documentation is synchronized.")
		return
	}

	createdCount, updatedCount, unchangedCount := syncGenerated(outputDir, generated)

	prunedCount := 0
	if pruneStale {
		for _, name := range stale {
			destFilePath := filepath.Join(outputDir, name)
			if err := os.Remove(destFilePath); err != nil {
				fmt.Fprintf(os.Stderr, "Error removing stale page %s: %v\n", destFilePath, err)
				os.Exit(1)
			}
			prunedCount++
		}
	} else if len(stale) > 0 {
		fmt.Fprintf(os.Stderr, "Stale pages (run with --prune to delete): %s\n", strings.Join(stale, ", "))
	}

	absPath, _ := filepath.Abs(outputDir)
	theme.SuccessMessage(fmt.Sprintf(
		"Documentation synchronized at: %s (Created: %d, Updated: %d, Unchanged: %d, Pruned: %d)",
		absPath,
		createdCount,
		updatedCount,
		unchangedCount,
		prunedCount,
	))
}

// findStale returns output-dir pages with no generated counterpart.
func findStale(generated map[string][]byte, existing map[string]bool) []string {
	var stale []string
	for name := range existing {
		if _, ok := generated[name]; !ok {
			stale = append(stale, name)
		}
	}
	return stale
}

// checkDrift reports content drift and stale pages, returning the count.
func checkDrift(outputDir string, generated map[string][]byte, stale []string) int {
	drift := 0
	for name, newContent := range generated {
		existingContent, err := os.ReadFile(filepath.Join(outputDir, name))
		if err != nil || !bytes.Equal(newContent, existingContent) {
			fmt.Fprintf(os.Stderr, "Drift detected: %s\n", name)
			drift++
		}
	}
	for _, name := range stale {
		fmt.Fprintf(os.Stderr, "Stale page: %s\n", name)
		drift++
	}
	if drift > 0 {
		fmt.Fprintf(os.Stderr, "Docs are out of date (run 'task docs').\n")
	}
	return drift
}

// syncGenerated writes new/changed pages, skipping identical files to
// preserve mtime/git status. Returns created, updated, unchanged counts.
func syncGenerated(outputDir string, generated map[string][]byte) (created, updated, unchanged int) {
	for name, newContent := range generated {
		destFilePath := filepath.Join(outputDir, name)

		existingContent, err := os.ReadFile(destFilePath)
		if err == nil && bytes.Equal(newContent, existingContent) {
			// Content is identical: do not touch the destination file so mtime/git status is preserved
			unchanged++
			continue
		}

		if os.IsNotExist(err) {
			created++
		} else {
			updated++
		}

		if err := os.WriteFile(destFilePath, newContent, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", destFilePath, err)
			os.Exit(1)
		}
	}
	return created, updated, unchanged
}
