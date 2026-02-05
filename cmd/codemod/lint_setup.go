// Package codemod provides helpers for codemods and project automation.
// This file contains the lint-setup command and related functionality.
package codemod

import (
	_ "embed"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/utils/log"
)

//go:embed eslint.config.standard.tmpl
var standardConfigTmpl []byte

//go:embed eslint.config.echo.tmpl
var echoConfigTmpl []byte

//go:embed eslint.config.js-only.tmpl
var jsOnlyConfigTmpl []byte

// detectTypeScriptUsage checks if the project uses TypeScript by looking for .ts/.tsx files or typescript dependency.
func detectTypeScriptUsage() bool {
	// Check for TypeScript files
	hasTSFiles := false
	err := filepath.Walk(".", func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // ignore errors
		}
		if !info.IsDir() && (strings.HasSuffix(info.Name(), ".ts") || strings.HasSuffix(info.Name(), ".tsx")) {
			hasTSFiles = true
			return filepath.SkipAll // found, no need to continue
		}
		return nil
	})
	if err != nil {
		// If walk fails, assume no TS files
		hasTSFiles = false
	}

	// Check package.json for typescript dependency
	if !hasTSFiles {
		if pkgData, err := os.ReadFile("package.json"); err == nil {
			var pkg map[string]interface{}
			if json.Unmarshal(pkgData, &pkg) == nil {
				// Check dependencies and devDependencies for typescript
				checkDeps := func(deps interface{}) bool {
					if depsMap, ok := deps.(map[string]interface{}); ok {
						for pkgName := range depsMap {
							if strings.HasPrefix(pkgName, "typescript") {
								return true
							}
						}
					}
					return false
				}
				if checkDeps(pkg["dependencies"]) || checkDeps(pkg["devDependencies"]) {
					hasTSFiles = true
				}
			}
		}
	}

	return hasTSFiles
}

// LintSetupCmd sets up linting, formatting, and pre-commit hooks for a Node.js project.
var LintSetupCmd = &cobra.Command{
	Use:   "lint-setup",
	Short: "Setup linting and formatting for a Node.js project",
	Long:  `Install and configure linting, formatting, and pre-commit hooks for a Node.js project (eslint, prettier, husky, etc).`,
	Run: func(cmd *cobra.Command, _args []string) {
		if _, err := os.Stat("package.json"); errors.Is(err, os.ErrNotExist) {
			log.Error("package.json not found in current directory")
			return
		}

		if err := installLintDependencies(echo); err != nil {
			log.Error("npm install failed: %v", err)
			return
		}

		if err := writeESLintConfig(echo); err != nil {
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

// installLintDependencies installs lint/format dependencies via npm or yarn and handles peer dep errors for npm.
func installLintDependencies(echo bool) error {
	log.Info("Installing lint/format dependencies via npm or yarn...")
	var installCmd *exec.Cmd
	var installArgs []string
	usingNpm := true

	baseDeps := []string{
		"eslint@latest", "@eslint/js@latest",
		"eslint-config-prettier@latest", "eslint-plugin-prettier@latest",
		"@eng618/prettier-config@latest", "globals@latest",
		"husky@latest", "lint-staged@latest", "prettier@latest",
		"prettier-plugin-organize-imports@latest",
	}

	// Add TypeScript ESLint dependencies only if TypeScript is detected
	if detectTypeScriptUsage() {
		baseDeps = append(baseDeps, "@typescript-eslint/eslint-plugin@latest", "@typescript-eslint/parser@latest")
	}

	if echo {
		baseDeps = append(baseDeps, "echo-eslint-config@latest")
	}

	if _, err := os.Stat("yarn.lock"); err == nil {
		usingNpm = false
		installArgs = append([]string{"add", "--dev"}, baseDeps...)
		installCmd = execCommand("yarn", installArgs...)
	} else {
		installArgs = append([]string{"install", "--save-dev"}, baseDeps...)
		installCmd = execCommand("npm", installArgs...)
	}

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
	if err != nil && usingNpm {
		stderrStr := string(stderrBytes)
		if strings.Contains(stderrStr, "--legacy-peer-deps") ||
			strings.Contains(stderrStr, "could not resolve dependency") {
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
	} else if err != nil {
		return err
	}
	return nil
}

// writeESLintConfig writes the eslint.config.mjs file.
func writeESLintConfig(echo bool) error {
	log.Info("Writing eslint.config.mjs...")
	var data []byte
	if echo {
		data = echoConfigTmpl
	} else if detectTypeScriptUsage() {
		data = standardConfigTmpl
	} else {
		data = jsOnlyConfigTmpl
	}
	return os.WriteFile("eslint.config.mjs", data, 0o644)
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
	scripts["lint"] = "eslint . --cache"
	scripts["lint:fix"] = "eslint . --cache --fix"
	scripts["lint:report"] = "eslint . --cache -o ./eslintReport.html -f html"
	scripts["prepare"] = "husky || echo 'Husky not installed, probably in ci'"
	pkg["scripts"] = scripts
	// Add lint-staged config
	pkg["lint-staged"] = map[string]interface{}{
		"*.(md)?(x)":            []string{"prettier --write"},
		"*.(js|ts|mjs|cjs)?(x)": []string{"prettier --write", "eslint --cache --fix"},
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
	return os.WriteFile("package.json", ordered, 0o644)
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
	if err := os.WriteFile(preCommitPath, []byte(hookContent), 0o644); err != nil {
		return err
	}
	return nil
}
