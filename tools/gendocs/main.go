package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra/doc"

	"github.com/eng618/eng/cmd"
	"github.com/eng618/eng/internal/ui/theme"
)

func main() {
	outputDir := "./docs/commands"
	if len(os.Args) > 1 {
		outputDir = os.Args[1]
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory %s: %v\n", outputDir, err)
		os.Exit(1)
	}

	rootCmd := cmd.GetRootCommand()
	if err := doc.GenMarkdownTree(rootCmd, outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating markdown documentation: %v\n", err)
		os.Exit(1)
	}

	absPath, _ := filepath.Abs(outputDir)
	theme.SuccessMessage(fmt.Sprintf("Generated CLI documentation at: %s", absPath))
}
