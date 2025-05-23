// Package codemod provides helpers for codemods and project automation.
package codemod

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"sort"
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

// installLintDependencies installs lint/format dependencies via npm and handles peer dep errors.
func installLintDependencies() error {
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
	stderrPipe, err := installCmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := installCmd.Start(); err != nil {
		return err
	}
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
				return err2
			}
		} else {
			return err
		}
	}
	return nil
}

// writeESLintConfig writes the eslint.config.mjs file.
func writeESLintConfig() error {
	log.Info("Writing eslint.config.mjs...")
	eslintConfig := `import echoConfig, { echoGlobalsOverride, echoJestGlobalsOverride } from 'echo-eslint-config';

export default [
  ...echoConfig,
  echoGlobalsOverride,
  echoJestGlobalsOverride,
];
`
	return os.WriteFile("eslint.config.mjs", []byte(eslintConfig), 0644)
}

// updatePackageJSON updates scripts, lint-staged, and prettier config in package.json with standard field order.
func updatePackageJSON() error {
	log.Info("Adding scripts and config to package.json...")
	pkgData, err := os.ReadFile("package.json")
	if err != nil {
		return err
	}
	var pkg map[string]interface{}
	if err := json.Unmarshal(pkgData, &pkg); err != nil {
		return err
	}
	// Ensure scripts exists
	scripts, ok := pkg["scripts"].(map[string]interface{})
	if !ok {
		scripts = make(map[string]interface{})
	}
	scripts["format"] = "prettier --write ."
	scripts["format:ci"] = "prettier --check ."
	scripts["lint"] = "eslint . --cache --ext .js,.jsx,.ts,.tsx,.mjs,.cjs"
	scripts["lint:fix"] = "eslint . --cache --fix --ext .js,.jsx,.ts,.tsx,.mjs,.cjs"
	scripts["lint:report"] = "eslint . --cache --ext .js,.jsx,.ts,.tsx,.mjs,.cjs -o ./eslintReport.html -f html"
	scripts["prepare"] = "husky || echo 'Husky not installed, probably in ci'"
	pkg["scripts"] = scripts
	// Add lint-staged config
	pkg["lint-staged"] = map[string]interface{}{
		"*.{js,ts,md,jsx,tsx,mdx}": []string{"prettier --write", "eslint --cache --fix"},
	}
	// Add prettier config
	pkg["prettier"] = "@eng618/prettier-config"
	// Write back to package.json with standard field order
	standardOrder := []string{
		"name", "version", "description", "keywords", "homepage", "bugs", "license", "author", "contributors", "funding", "main", "module", "types", "exports", "files", "bin", "directories", "repository", "scripts", "dependencies", "devDependencies", "peerDependencies", "optionalDependencies", "engines", "os", "cpu", "private", "publishConfig", "lint-staged", "prettier",
	}
	allKeys := make(map[string]struct{})
	for k := range pkg {
		allKeys[k] = struct{}{}
	}
	var extraKeys []string
	for k := range allKeys {
		found := false
		for _, std := range standardOrder {
			if k == std {
				found = true
				break
			}
		}
		if !found {
			extraKeys = append(extraKeys, k)
		}
	}
	sort.Strings(extraKeys)
	ordered := make([]byte, 0, 4096)
	indent := "  "
	first := true
	writeField := func(key string) {
		if val, ok := pkg[key]; ok {
			if !first {
				ordered = append(ordered, ',', '\n')
			}
			first = false
			ordered = append(ordered, []byte(indent+"\""+key+"\": ")...)
			valBytes, _ := json.MarshalIndent(val, indent, indent)
			ordered = append(ordered, valBytes...)
		}
	}
	ordered = append(ordered, '{', '\n')
	for _, k := range standardOrder {
		writeField(k)
	}
	for _, k := range extraKeys {
		writeField(k)
	}
	ordered = append(ordered, '\n', '}')
	return os.WriteFile("package.json", ordered, 0644)
}

// setupHusky runs husky init and overwrites the pre-commit hook.
func setupHusky() error {
	log.Info("Setting up Husky and pre-commit hook...")
	huskyInit := execCommand("npx", "husky", "init")
	huskyInit.Stdout = log.Writer()
	huskyInit.Stderr = log.ErrorWriter()
	if err := huskyInit.Run(); err != nil {
		return err
	}
	preCommitPath := ".husky/pre-commit"
	hookContent := "npx lint-staged\n"
	if err := os.WriteFile(preCommitPath, []byte(hookContent), 0644); err != nil {
		return err
	}
	return nil
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

		if err := installLintDependencies(); err != nil {
			log.Error("npm install failed: %v", err)
			return
		}

		if err := writeESLintConfig(); err != nil {
			log.Error("Failed to write eslint.config.mjs: %v", err)
			return
		}

		if err := updatePackageJSON(); err != nil {
			log.Error("Failed to update package.json: %v", err)
			return
		}

		if err := setupHusky(); err != nil {
			log.Error("Failed to set up Husky: %v", err)
			return
		}

		log.Success("Linting, formatting, and pre-commit hooks are set up!")
	},
}

func init() {
	CodemodCmd.AddCommand(LintSetupCmd)
}
