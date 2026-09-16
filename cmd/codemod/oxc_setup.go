// Package codemod provides helpers for codemods and project automation.
// This file contains the oxc-setup command and related functionality.
package codemod

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

//go:embed oxlint.config.recommended.tmpl
var oxlintRecommendedTmpl []byte

//go:embed oxlint.config.next.tmpl
var oxlintNextTmpl []byte

//go:embed oxlint.config.vite.tmpl
var oxlintViteTmpl []byte

//go:embed oxlint.config.react.tmpl
var oxlintReactTmpl []byte

//go:embed oxlint.config.typescript.tmpl
var oxlintTypeScriptTmpl []byte

//go:embed oxlint.config.base.tmpl
var oxlintBaseTmpl []byte

//go:embed oxfmt.config.tmpl
var oxfmtConfigTmpl []byte

// oxcConfirmPrompt asks a yes/no question. Defaults to ui.Confirm so tests can override it.
var oxcConfirmPrompt = ui.Confirm

// Flags for the oxc-setup command.
var (
	oxcPreset       string
	oxcTypeAware    bool
	oxcRemoveEslint bool
	oxcYes          bool
)

// eslintPrettierPackages lists the ESLint/Prettier stack replaced by @gv-tech/oxc-config.
var eslintPrettierPackages = []string{ //nolint:gochecknoglobals
	"eslint",
	"prettier",
	"@gv-tech/eslint-config",
	"@eng618/prettier-config",
	"echo-eslint-config",
	"eslint-config-prettier",
	"eslint-plugin-prettier",
	"@eslint/js",
	"globals",
	"typescript-eslint",
	"@typescript-eslint/eslint-plugin",
	"@typescript-eslint/parser",
	"@next/eslint-plugin-next",
}

// eslintPrettierFiles lists legacy config files removed during migration.
var eslintPrettierFiles = []string{ //nolint:gochecknoglobals
	"eslint.config.mjs",
	"eslint.config.cjs",
	"eslint.config.js",
	".eslintrc",
	".eslintrc.js",
	".eslintrc.cjs",
	".eslintrc.json",
	".eslintrc.yaml",
	".eslintrc.yml",
	".prettierrc",
	".prettierrc.js",
	".prettierrc.cjs",
	".prettierrc.json",
	".prettierrc.yaml",
	".prettierrc.yml",
	"prettier.config.js",
	"prettier.config.cjs",
	"prettier.config.mjs",
}

// OxcSetupCmd installs and configures Oxlint + Oxfmt via @gv-tech/oxc-config.
var OxcSetupCmd = &cobra.Command{
	Use:   "oxc-setup",
	Short: "Setup Oxlint + Oxfmt via @gv-tech/oxc-config",
	Long: `Install and configure Oxlint + Oxfmt for a Node.js project using @gv-tech/oxc-config.

Writes oxlint.config.ts and oxfmt.config.ts, updates package.json scripts
(lint, lint:fix, format, format:ci), configures lint-staged + Husky, and can
migrate off ESLint/Prettier.

Migration: pass --remove-eslint to uninstall the ESLint/Prettier stack and
delete legacy configs, or run without the flag to be prompted when an
ESLint/Prettier setup is detected.`,
	Example: `  eng codemod oxc-setup
  eng codemod oxc-setup --preset next
  eng codemod oxc-setup --preset vite --type-aware
  eng codemod oxc-setup --remove-eslint --yes`,
	Run: func(_ *cobra.Command, _ []string) {
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary).
			MarginBottom(1)
		if !ui.DisableProgress {
			fmt.Fprintln(log.Out, headerStyle.Render("⚡ Oxlint + Oxfmt Setup"))
		}

		if _, err := os.Stat("package.json"); errors.Is(err, os.ErrNotExist) {
			log.Error("package.json not found in current directory")
			return
		}

		checkNodeVersion()

		preset, err := resolveOxcPreset(oxcPreset)
		if err != nil {
			log.Error("%v", err)
			return
		}
		log.Info("Using oxc preset: %s", preset)

		if err := installOxcDependencies(preset, oxcTypeAware); err != nil {
			log.Error("package manager install failed: %v", err)
			return
		}

		if oxcRemoveEslint {
			if err := removeEslintPrettier(); err != nil {
				log.Error("Failed to remove ESLint/Prettier: %v", err)
				return
			}
		} else if detectEslintPrettier() {
			proceed := oxcYes
			if !oxcYes {
				var promptErr error
				proceed, promptErr = oxcConfirmPrompt(
					"ESLint/Prettier setup detected. Remove it and migrate to Oxlint + Oxfmt?",
					true,
				)
				if promptErr != nil {
					log.Error("Prompt failed: %v", promptErr)
					return
				}
			}
			if proceed {
				if err := removeEslintPrettier(); err != nil {
					log.Error("Failed to remove ESLint/Prettier: %v", err)
					return
				}
			} else {
				log.Info("Keeping existing ESLint/Prettier setup alongside Oxlint + Oxfmt.")
			}
		}

		if err := writeOxcConfigs(preset, oxcTypeAware); err != nil {
			log.Error("Failed to write oxlint.config.ts / oxfmt.config.ts: %v", err)
			return
		}

		if err := updatePackageJSONForOxc(); err != nil {
			log.Error("Failed to update package.json: %v", err)
			return
		}

		if err := setupHusky(); err != nil {
			log.Error("Failed to set up Husky: %v", err)
			return
		}

		log.Success("Oxlint + Oxfmt are set up! Run 'oxlint .' and 'oxfmt --check .' to verify.")
	},
}

// resolveOxcPreset validates an explicit --preset flag or auto-detects one.
func resolveOxcPreset(flag string) (string, error) {
	valid := map[string]bool{
		"": true, "auto": true,
		"recommended": true, "next": true, "vite": true,
		"react": true, "typescript": true, "base": true,
	}
	if !valid[flag] {
		return "", fmt.Errorf(
			"invalid --preset %q: must be one of auto, recommended, next, vite, react, typescript, base",
			flag,
		)
	}
	if flag != "" && flag != "auto" {
		return flag, nil
	}
	switch {
	case detectNextJsUsage():
		return "next", nil
	case detectViteUsage() || detectReactUsage():
		// React apps default to the vite preset per oxc-config guidance
		// (Next.js projects are handled above).
		if detectViteUsage() {
			return "vite", nil
		}
		return "react", nil
	case detectTypeScriptUsage():
		return "typescript", nil
	default:
		return "recommended", nil
	}
}

// detectReactUsage checks for React usage via package.json deps or JSX/TSX files.
func detectReactUsage() bool {
	if pkgData, err := os.ReadFile("package.json"); err == nil {
		var pkg map[string]interface{}
		if json.Unmarshal(pkgData, &pkg) == nil {
			for _, key := range []string{"dependencies", "devDependencies"} {
				if depsMap, ok := pkg[key].(map[string]interface{}); ok {
					if _, ok := depsMap["react"]; ok {
						return true
					}
				}
			}
		}
	}
	found := false
	_ = filepath.Walk(".", func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() &&
			(strings.HasSuffix(info.Name(), ".jsx") || strings.HasSuffix(info.Name(), ".tsx")) {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

// detectViteUsage checks for Vite usage via package.json deps or vite config files.
func detectViteUsage() bool {
	if pkgData, err := os.ReadFile("package.json"); err == nil {
		var pkg map[string]interface{}
		if json.Unmarshal(pkgData, &pkg) == nil {
			for _, key := range []string{"dependencies", "devDependencies"} {
				if depsMap, ok := pkg[key].(map[string]interface{}); ok {
					if _, ok := depsMap["vite"]; ok {
						return true
					}
				}
			}
		}
	}
	for _, name := range []string{
		"vite.config.ts", "vite.config.mts", "vite.config.js", "vite.config.mjs",
	} {
		if _, err := os.Stat(name); err == nil {
			return true
		}
	}
	return false
}

// checkNodeVersion warns when node is older than the 22.18 minimum required for oxlint.config.ts.
func checkNodeVersion() {
	cmd := execCommand("node", "--version")
	out, err := cmd.Output()
	if err != nil {
		log.Warn("Could not detect node version; @gv-tech/oxc-config requires node >= 22.18.0")
		return
	}
	ver := strings.TrimSpace(string(out))
	ver = strings.TrimPrefix(ver, "v")
	parts := strings.Split(ver, ".")
	if len(parts) < 2 {
		return
	}
	major, err1 := strconv.Atoi(parts[0])
	minorStr := strings.SplitN(parts[1], "-", 2)[0]
	minor, err2 := strconv.Atoi(minorStr)
	if err1 != nil || err2 != nil {
		return
	}
	if major < 22 || (major == 22 && minor < 18) {
		log.Warn("Detected node %s; @gv-tech/oxc-config requires node >= 22.18.0", ver)
	}
}

// installOxcDependencies installs Oxlint + Oxfmt dependencies via npm, yarn, or bun.
func installOxcDependencies(preset string, typeAware bool) error {
	log.Info("Installing Oxlint/Oxfmt dependencies via npm, yarn, or bun...")
	var installCmd *exec.Cmd
	var installArgs []string
	var packageManager string

	baseDeps := []string{
		"@gv-tech/oxc-config@latest",
		"oxlint@latest",
		"oxfmt@latest",
		"husky@latest", "lint-staged@latest",
	}

	// TypeScript is a required peer dep for TS presets.
	if preset != "base" && detectTypeScriptUsage() {
		baseDeps = append(baseDeps, "typescript@latest")
	}

	if typeAware {
		baseDeps = append(baseDeps, "oxlint-tsgolint@latest")
	}

	// Detect package manager: bun > yarn > npm
	if _, err := os.Stat("bun.lock"); err == nil {
		packageManager = "bun"
		installArgs = append([]string{"add", "--dev"}, baseDeps...)
		installCmd = execCommand("bun", installArgs...)
	} else if _, err := os.Stat("yarn.lock"); err == nil {
		packageManager = "yarn"
		installArgs = append([]string{"add", "--dev"}, baseDeps...)
		installCmd = execCommand("yarn", installArgs...)
	} else {
		packageManager = "npm"
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
	if err != nil && packageManager == "npm" {
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

// detectEslintPrettier reports whether an ESLint/Prettier setup is present.
func detectEslintPrettier() bool {
	if pkgData, err := os.ReadFile("package.json"); err == nil {
		var pkg map[string]interface{}
		if json.Unmarshal(pkgData, &pkg) == nil {
			for _, key := range []string{"dependencies", "devDependencies"} {
				if depsMap, ok := pkg[key].(map[string]interface{}); ok {
					for _, dep := range eslintPrettierPackages {
						if _, ok := depsMap[dep]; ok {
							return true
						}
					}
				}
			}
			if _, ok := pkg["prettier"]; ok {
				return true
			}
			if _, ok := pkg["eslintConfig"]; ok {
				return true
			}
		}
	}
	for _, name := range eslintPrettierFiles {
		if _, err := os.Stat(name); err == nil {
			return true
		}
	}
	return false
}

// removeEslintPrettier uninstalls the ESLint/Prettier stack and deletes legacy config files.
func removeEslintPrettier() error {
	log.Info("Removing ESLint/Prettier dependencies...")

	var installed []string
	if pkgData, err := os.ReadFile("package.json"); err == nil {
		var pkg map[string]interface{}
		if json.Unmarshal(pkgData, &pkg) == nil {
			seen := map[string]bool{}
			for _, key := range []string{"dependencies", "devDependencies"} {
				if depsMap, ok := pkg[key].(map[string]interface{}); ok {
					for _, dep := range eslintPrettierPackages {
						if _, ok := depsMap[dep]; ok && !seen[dep] {
							seen[dep] = true
							installed = append(installed, dep)
						}
					}
				}
			}
		}
	}

	if len(installed) > 0 {
		var removeCmd *exec.Cmd
		if _, err := os.Stat("bun.lock"); err == nil {
			removeCmd = execCommand("bun", append([]string{"remove"}, installed...)...)
		} else if _, err := os.Stat("yarn.lock"); err == nil {
			removeCmd = execCommand("yarn", append([]string{"remove"}, installed...)...)
		} else {
			removeCmd = execCommand("npm", append([]string{"uninstall"}, installed...)...)
		}
		removeCmd.Stdout = log.Writer()
		removeCmd.Stderr = log.ErrorWriter()
		if err := removeCmd.Run(); err != nil {
			return err
		}
	} else {
		log.Info("No ESLint/Prettier dependencies found in package.json.")
	}

	for _, name := range eslintPrettierFiles {
		if _, err := os.Stat(name); err == nil {
			log.Info("Removing %s...", name)
			if err := os.Remove(name); err != nil {
				return err
			}
		}
	}
	return nil
}

// writeOxcConfigs writes oxlint.config.ts and oxfmt.config.ts files.
func writeOxcConfigs(preset string, typeAware bool) error {
	log.Info("Writing oxlint.config.ts and oxfmt.config.ts...")
	var data []byte
	switch preset {
	case "next":
		data = oxlintNextTmpl
	case "vite":
		data = oxlintViteTmpl
	case "react":
		data = oxlintReactTmpl
	case "typescript":
		data = oxlintTypeScriptTmpl
	case "base":
		data = oxlintBaseTmpl
	default:
		data = oxlintRecommendedTmpl
	}

	if typeAware {
		data = []byte(injectTypeAware(string(data)))
	}

	if err := os.WriteFile("oxlint.config.ts", data, 0o644); err != nil {
		return err
	}
	return os.WriteFile("oxfmt.config.ts", oxfmtConfigTmpl, 0o644)
}

// injectTypeAware adds the typeAware preset to an oxlint config's extends list.
func injectTypeAware(config string) string {
	if strings.Contains(config, "typeAware") {
		return config
	}
	config = strings.Replace(
		config,
		"import { defineConfig } from 'oxlint';",
		"import { defineConfig } from 'oxlint';\nimport { typeAware } from '@gv-tech/oxc-config/type-aware';",
		1,
	)
	// Handle `extends: [x]` with any spacing.
	if idx := strings.Index(config, "extends: ["); idx != -1 {
		end := strings.Index(config[idx:], "]")
		if end != -1 {
			inner := strings.TrimSpace(config[idx+len("extends: [") : idx+end])
			inner = strings.TrimSuffix(inner, ",")
			config = config[:idx] + "extends: [" + inner + ", typeAware]" + config[idx+end+1:]
		}
	}
	return config
}

// updatePackageJSONForOxc updates scripts and lint-staged for Oxlint + Oxfmt.
func updatePackageJSONForOxc() error {
	log.Info("Adding Oxlint/Oxfmt scripts and config to package.json...")
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
	scripts["format"] = "oxfmt --write ."
	scripts["format:ci"] = "oxfmt --check ."
	scripts["lint"] = "oxlint ."
	scripts["lint:fix"] = "oxlint --fix"
	delete(scripts, "lint:report")
	scripts["prepare"] = "husky || echo 'Husky not installed, probably in ci'"
	pkg["scripts"] = scripts
	// Add lint-staged config (oxc style)
	pkg["lint-staged"] = map[string]interface{}{
		"*.{js,jsx,ts,tsx,mjs,cjs,mts,cts}": []string{"oxfmt --write", "oxlint --fix"},
		"*.{json,jsonc,md,yml,yaml}":        []string{"oxfmt --write"},
	}
	// Remove prettier key; formatting moves to Oxfmt.
	delete(pkg, "prettier")
	delete(pkg, "eslintConfig")
	return writeOrderedPackageJSON(pkg)
}

// writeOrderedPackageJSON writes package.json with the standard field order.
func writeOrderedPackageJSON(pkg map[string]interface{}) error {
	standardOrder := []string{
		"name",
		"version",
		"description",
		"keywords",
		"homepage",
		"bugs",
		"license",
		"author",
		"type",
		"exports",
		"main",
		"module",
		"types",
		"files",
		"bin",
		"directories",
		"repository",
		"scripts",
		"dependencies",
		"devDependencies",
		"peerDependencies",
		"peerDependenciesMeta",
		"optionalDependencies",
		"overrides",
		"resolutions",
		"engines",
		"packageManager",
		"os",
		"cpu",
		"private",
		"publishConfig",
		"lint-staged",
		"prettier",
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
