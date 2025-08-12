// Package codemod provides helpers for codemods and project automation.
// This file contains the copilot command and related functionality.
package codemod

import (
	"os"

	"github.com/eng618/eng/utils/log"
	"github.com/spf13/cobra"
)

// CopilotSetupCmd creates a base custom Copilot instructions file in .github/copilot-instructions.md.
var CopilotSetupCmd = &cobra.Command{
	Use:   "copilot",
	Short: "Setup custom Copilot instructions file",
	Long:  `Create a base custom Copilot instructions file at .github/copilot-instructions.md. By default, this command assumes you are in the root of a Git repository. Use --force to bypass this check.`,
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")
		
		// Check if we're in a git repository unless force is used
		if !force {
			if !isGitRepository() {
				log.Error("Not in a Git repository. Use --force to bypass this check.")
				return
			}
		}

		if err := createCopilotInstructions(); err != nil {
			log.Error("Failed to create Copilot instructions: %v", err)
			return
		}

		log.Success("Copilot instructions file created at .github/copilot-instructions.md")
	},
}

func init() {
	CopilotSetupCmd.Flags().Bool("force", false, "Create the file even if not in a Git repository")
}

// isGitRepository checks if the current directory is a Git repository.
func isGitRepository() bool {
	if _, err := os.Stat(".git"); err == nil {
		return true
	}
	
	// Check if we're inside a git worktree by running git rev-parse
	cmd := execCommand("git", "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil
}

// createCopilotInstructions creates the .github/copilot-instructions.md file with base template.
func createCopilotInstructions() error {
	log.Info("Creating .github/copilot-instructions.md...")
	
	// Create .github directory if it doesn't exist
	if err := os.MkdirAll(".github", 0755); err != nil {
		return err
	}
	
	// Check if file already exists
	filePath := ".github/copilot-instructions.md"
	if _, err := os.Stat(filePath); err == nil {
		log.Info("Copilot instructions file already exists, skipping creation.")
		return nil
	}
	
	// Base template for copilot instructions
	template := `# Copilot Custom Instructions

<!-- Use this file to provide workspace-specific custom instructions to Copilot. For more details, visit https://code.visualstudio.com/docs/copilot/copilot-customization#_use-a-githubcopilotinstructionsmd-file -->

## General

- Keep the README.md file up to date with the latest information about the project.
- Use clear and concise language in comments and documentation.
- Always document your code.
- Follow established patterns and conventions in the codebase.

## Code Quality

- Write tests for new functionality.
- Update tests when modifying existing code.
- Use meaningful variable and function names.
- Keep functions small and focused on a single responsibility.
- Follow the DRY (Don't Repeat Yourself) principle.

## Comments and Documentation

- Use JSDoc style comments for JavaScript and TypeScript.
- Use GoDoc style comments for Go.
- Use Python docstrings for Python.
- Use docstrings for Dart.
- Use Ruby style comments for Ruby.
- Use PHPDoc style comments for PHP.
- Use XML comments for C#.
- Use block comments for shell scripts.
- Use block comments for Makefiles.

## git Commit messages

- Use conventional commit messages
- Follow the commit message guidelines
- See https://www.conventionalcommits.org/en/v1.0.0/
- Start commit messages with a type (e.g., feat, fix, docs, style, refactor, test, chore)
- Optionally include a scope in parentheses after the type (e.g., feat(parser): ...)
- Use a short, imperative description after the type/scope
- Separate the body from the header with a blank line if additional context is needed
- Reference issues or pull requests in the footer if applicable
- Mention breaking changes in the footer if present

## Project Specific

<!-- Add project-specific instructions here -->
`
	
	return os.WriteFile(filePath, []byte(template), 0644)
}
