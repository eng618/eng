// Package codemod provides helpers for codemods and project automation.
package codemod

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/eng618/eng/utils/log"
	"github.com/spf13/cobra"
)

// execCommand is a variable holding the exec.Command function, allowing for test overrides.
var execCommand = exec.Command

// CodemodCmd is the root command for codemod-related helpers and automation.
var CodemodCmd = &cobra.Command{
	Use:   "codemod",
	Short: "Helpers for codemods and project automation",
	Long:  `Run codemods or setup helpers for various project types.`,
}

// LintSetupCmd sets up linting, formatting, and pre-commit hooks for a Node.js project.
var LintSetupCmd = &cobra.Command{
	Use:   "lint-setup",
	Short: "Setup linting and formatting for a Node.js project",
	Long:  `Install and configure linting, formatting, and pre-commit hooks for a Node.js project (eslint, prettier, husky, etc).`,
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := os.Stat("package.json"); errors.Is(err, os.ErrNotExist) {
			log.Error("package.json not found in current directory")
			return
		}

		log.Info("Installing lint/format dependencies via npm...")
			installArgs := []string{"install", "--save-dev",
				"eslint@latest", "@eslint/js@latest",
				"@typescript-eslint/eslint-plugin@latest", "@typescript-eslint/parser@latest",
				"eslint-config-prettier@latest", "eslint-plugin-prettier@latest",
				"@eng618/prettier-config@latest", "globals",
				"echo-eslint-config@latest", "husky@latest", "lint-staged@latest", "prettier@latest",
			}
			installCmd := execCommand("npm", installArgs...)
			installCmd.Stdout = log.Writer()

			// Capture stderr for error analysis
			stderrPipe, err := installCmd.StderrPipe()
			if err != nil {
				log.Error("Failed to get stderr pipe: %v", err)
				return
			}
			if err := installCmd.Start(); err != nil {
				log.Error("npm install failed to start: %v", err)
				return
			}

			// Read stderr output
			stderrBytes, _ := io.ReadAll(stderrPipe)
			err = installCmd.Wait()
			if err != nil {
				stderrStr := string(stderrBytes)
				if strings.Contains(stderrStr, "--legacy-peer-deps") || strings.Contains(stderrStr, "could not resolve dependency") {
					log.Info("npm install failed due to peer deps, retrying with --legacy-peer-deps...")
					installArgs = append(installArgs, "--legacy-peer-deps")
					installCmd2 := execCommand("npm", installArgs...)
					installCmd2.Stdout = log.Writer()
					installCmd2.Stderr = log.ErrorWriter()
					if err2 := installCmd2.Run(); err2 != nil {
						log.Error("npm install with --legacy-peer-deps failed: %v", err2)
						return
					}
				} else {
					log.Error("npm install failed: %v", err)
					return
				}
			}

		log.Info("Writing eslint.config.mjs...")
		eslintConfig := `import echoConfig, { echoGlobalsOverride, echoJestGlobalsOverride } from 'echo-eslint-config';

export default [
  ...echoConfig,
  echoGlobalsOverride,
  echoJestGlobalsOverride,
];
`
		if err := os.WriteFile("eslint.config.mjs", []byte(eslintConfig), 0644); err != nil {
			log.Error("Failed to write eslint.config.mjs: %v", err)
			return
		}

		log.Info("Adding scripts and config to package.json...")
		pkgData, err := os.ReadFile("package.json")
		if err != nil {
			log.Error("Failed to read package.json: %v", err)
			return
		}
		var pkg map[string]interface{}
		if err := json.Unmarshal(pkgData, &pkg); err != nil {
			log.Error("Failed to parse package.json: %v", err)
			return
		}

		// Ensure scripts exists
		scripts, ok := pkg["scripts"].(map[string]interface{})
		if !ok {
			scripts = make(map[string]interface{})
		}
		scripts["format"] = "prettier --write ."
		scripts["format:ci"] = "prettier --check ."
		scripts["lint"] = "eslint . --ext .js,.jsx,.ts,.tsx,.mjs,.cjs"
		scripts["lint:fix"] = "eslint . --fix --ext .js,.jsx,.ts,.tsx,.mjs,.cjs"
		scripts["prepare"] = "husky || echo 'Husky not installed, probably in ci'"
		pkg["scripts"] = scripts

		// Add lint-staged config
		pkg["lint-staged"] = map[string]interface{}{
			"*.{js,ts,md,jsx,tsx,mdx}": []string{"prettier --write", "eslint --cache --fix"},
		}
		// Add prettier config
		pkg["prettier"] = "@eng618/prettier-config"

		// Write back to package.json
		newPkgData, err := json.MarshalIndent(pkg, "", "  ")
		if err != nil {
			log.Error("Failed to marshal package.json: %v", err)
			return
		}
		if err := os.WriteFile("package.json", newPkgData, 0644); err != nil {
			log.Error("Failed to write package.json: %v", err)
			return
		}

		log.Info("Setting up Husky and pre-commit hook...")
		huskyInstall := execCommand("npx", "husky", "init")
		huskyInstall.Stdout = log.Writer()
		huskyInstall.Stderr = log.ErrorWriter()
		if err := huskyInstall.Run(); err != nil {
			log.Error("Failed to run 'npx husky init': %v", err)
			return
		}

		huskyAdd := execCommand("npx", "husky", "add", ".husky/pre-commit", "npx lint-staged")
		huskyAdd.Stdout = log.Writer()
		huskyAdd.Stderr = log.ErrorWriter()
		if err := huskyAdd.Run(); err != nil {
			log.Error("Failed to add pre-commit hook: %v", err)
			return
		}

		log.Success("Linting, formatting, and pre-commit hooks are set up!")
	},
}

func init() {
	CodemodCmd.AddCommand(LintSetupCmd)
}
