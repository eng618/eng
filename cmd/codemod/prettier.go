// Package codemod provides helpers for codemods and project automation.
// This file contains the prettier formatting command.
package codemod

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/utils/log"
)

// PrettierCmd formats the current directory with prettier, installing @eng618/prettier-config if needed.
var PrettierCmd = &cobra.Command{
	Use:   "prettier [path]",
	Short: "Format code with prettier using @eng618/prettier-config",
	Long: `Format code with prettier, automatically installing @eng618/prettier-config if not available.
If no path is provided, formats the current directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Determine the target path
		targetPath := "."
		if len(args) > 0 {
			targetPath = args[0]
		}

		if err := runPrettier(targetPath); err != nil {
			log.Error("Failed to run prettier: %v", err)
			return
		}

		log.Success("Prettier formatting completed!")
	},
}

// runPrettier executes the prettier formatting workflow.
func runPrettier(targetPath string) error {
	log.Info("Checking for @eng618/prettier-config...")

	// Check if @eng618/prettier-config is available
	if !canResolvePackage("@eng618/prettier-config") {
		log.Info("Installing @eng618/prettier-config globally...")
		if err := installPrettierConfig(); err != nil {
			return fmt.Errorf("failed to install @eng618/prettier-config: %w", err)
		}
	}

	// Try to use the prettier config if available, fallback to default if not
	if canResolvePackage("@eng618/prettier-config") {
		log.Info("Using @eng618/prettier-config...")
		configPath, err := getConfigPath("@eng618/prettier-config")
		if err == nil && configPath != "" {
			return runPrettierWithConfig(configPath, targetPath)
		}
		log.Warn("Could not resolve config path, using default prettier configuration...")
	}

	log.Info("Using default prettier configuration...")
	return runPrettierDefault(targetPath)
}

// canResolvePackage checks if a Node.js package can be resolved.
func canResolvePackage(packageName string) bool {
	cmd := execCommand(
		"node",
		"-e",
		fmt.Sprintf(
			"try { require.resolve('%s'); console.log('true'); } catch(e) { console.log('false'); }",
			packageName,
		),
	)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "true"
}

// installPrettierConfig installs @eng618/prettier-config globally.
func installPrettierConfig() error {
	cmd := execCommand("npm", "install", "-g", "@eng618/prettier-config")
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()
	if err := cmd.Run(); err != nil {
		return err
	}

	// Refresh the current shell's environment to pick up global npm packages
	npmPrefix, err := getNpmPrefix()
	if err == nil && npmPrefix != "" {
		currentPath := os.Getenv("PATH")
		binPath := filepath.Join(npmPrefix, "bin")
		if !strings.Contains(currentPath, binPath) {
			if err := os.Setenv("PATH", binPath+":"+currentPath); err != nil {
				// Log warning but don't fail the operation
				log.Warn("Failed to update PATH environment variable: %v", err)
			}
		}
	}

	return nil
}

// getNpmPrefix gets the npm global prefix path.
func getNpmPrefix() (string, error) {
	cmd := execCommand("npm", "config", "get", "prefix")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// getConfigPath resolves the path to a package.
func getConfigPath(packageName string) (string, error) {
	cmd := execCommand("node", "-p", fmt.Sprintf("require.resolve('%s')", packageName))
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// runPrettierWithConfig runs prettier with a specific config file.
func runPrettierWithConfig(configPath, targetPath string) error {
	cmd := execCommand("npx", "prettier", "--config", configPath, "--write", targetPath)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()
	return cmd.Run()
}

// runPrettierDefault runs prettier with default configuration.
func runPrettierDefault(targetPath string) error {
	cmd := execCommand("npx", "prettier", "--write", targetPath)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()
	return cmd.Run()
}
