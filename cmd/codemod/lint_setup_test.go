package codemod

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestLintSetupCmd_ModifiesPackageJson(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	}()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change to temp dir: %v", err)
	}
	// Setup: create a minimal package.json
	pkg := map[string]interface{}{
		"name": "testpkg",
	}
	data, _ := json.Marshal(pkg)
	if err := os.WriteFile("package.json", data, 0o644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}
	// Mock exec.Command to avoid running npm/husky
	oldCommand := execCommand
	defer func() { execCommand = oldCommand }()
	execCommand = func(name string, arg ...string) *exec.Cmd {
		if name == "npx" && len(arg) > 0 && arg[0] == "husky" {
			_ = os.MkdirAll(".husky", 0o755)
		}
		return exec.Command("echo", append([]string{name}, arg...)...)
	}
	// Run
	cmd := LintSetupCmd
	output := &strings.Builder{}
	cmd.SetOut(output)
	cmd.SetArgs([]string{})
	cmd.Run(cmd, []string{})
	// Check package.json was updated
	newData, err := os.ReadFile("package.json")
	if err != nil {
		t.Fatalf("failed to read package.json: %v", err)
	}
	var newPkg map[string]interface{}
	if err := json.Unmarshal(newData, &newPkg); err != nil {
		t.Fatalf("failed to parse updated package.json: %v", err)
	}
	if _, ok := newPkg["scripts"]; !ok {
		t.Error("expected scripts in package.json")
	}
	if _, ok := newPkg["lint-staged"]; !ok {
		t.Error("expected lint-staged in package.json")
	}
	if _, ok := newPkg["prettier"]; !ok {
		t.Error("expected prettier in package.json")
	}
}

func TestLintSetupCmd_Echo(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	}()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change to temp dir: %v", err)
	}
	// Setup: create a minimal package.json
	pkg := map[string]interface{}{
		"name": "testpkg",
	}
	data, _ := json.Marshal(pkg)
	if err := os.WriteFile("package.json", data, 0o644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}
	// Mock exec.Command to avoid running npm/husky
	oldCommand := execCommand
	defer func() { execCommand = oldCommand }()
	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("echo", append([]string{name}, arg...)...)
	}
	// Set echo flag manually (since global)
	oldEcho := echo
	echo = true
	defer func() { echo = oldEcho }()
	// Run
	cmd := LintSetupCmd
	output := &strings.Builder{}
	cmd.SetOut(output)
	cmd.SetArgs([]string{})
	cmd.Run(cmd, []string{})
	// Check eslint.config.mjs was created for echo
	eslintData, err := os.ReadFile("eslint.config.mjs")
	if err != nil {
		t.Fatalf("failed to read eslint.config.mjs: %v", err)
	}
	if !strings.Contains(string(eslintData), "echo-eslint-config") {
		t.Error("expected echo-eslint-config in eslint.config.mjs")
	}
}

func TestWriteESLintConfig(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tempDir)

	err := writeESLintConfig(false)
	if err != nil {
		t.Fatalf("writeESLintConfig failed: %v", err)
	}
	data, err := os.ReadFile("eslint.config.mjs")
	if err != nil {
		t.Fatalf("eslint.config.mjs not written: %v", err)
	}
	if strings.Contains(string(data), "echo-eslint-config") {
		t.Error("eslint.config.mjs should not contain echo-eslint-config for standard setup")
	}
}

func TestWriteESLintConfig_EchoMode(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tempDir)

	err := writeESLintConfig(true)
	if err != nil {
		t.Fatalf("writeESLintConfig failed: %v", err)
	}
	data, err := os.ReadFile("eslint.config.mjs")
	if err != nil {
		t.Fatalf("eslint.config.mjs not written: %v", err)
	}
	if !strings.Contains(string(data), "echo-eslint-config") {
		t.Error("eslint.config.mjs missing echo-eslint-config for echo mode")
	}
}

func TestDetectNextJsUsage(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tempDir)

	// Case 1: Next.js in dependencies
	pkgDeps := map[string]interface{}{
		"dependencies": map[string]interface{}{
			"next": "latest",
		},
	}
	data, _ := json.Marshal(pkgDeps)
	_ = os.WriteFile("package.json", data, 0o644)
	if !detectNextJsUsage() {
		t.Error("detectNextJsUsage should return true when next is in dependencies")
	}

	// Case 2: Next.js in devDependencies
	pkgDevDeps := map[string]interface{}{
		"devDependencies": map[string]interface{}{
			"next": "latest",
		},
	}
	data, _ = json.Marshal(pkgDevDeps)
	_ = os.WriteFile("package.json", data, 0o644)
	if !detectNextJsUsage() {
		t.Error("detectNextJsUsage should return true when next is in devDependencies")
	}

	// Case 3: No Next.js
	pkgNoNext := map[string]interface{}{
		"dependencies": map[string]interface{}{
			"react": "latest",
		},
	}
	data, _ = json.Marshal(pkgNoNext)
	_ = os.WriteFile("package.json", data, 0o644)
	if detectNextJsUsage() {
		t.Error("detectNextJsUsage should return false when next is not present")
	}
}

func TestDetectTypeScriptUsage_TypeScriptFiles(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tempDir)

	// Create a TypeScript file
	_ = os.WriteFile("index.ts", []byte("console.log('hello');"), 0o644)

	if !detectTypeScriptUsage() {
		t.Error("detectTypeScriptUsage should return true for projects with .ts files")
	}
}

func TestDetectTypeScriptUsage_TypeScriptDependency(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tempDir)

	// Create package.json with typescript dependency
	pkg := map[string]interface{}{
		"devDependencies": map[string]interface{}{
			"typescript": "^4.0.0",
		},
	}
	pkgData, _ := json.Marshal(pkg)
	_ = os.WriteFile("package.json", pkgData, 0o644)

	if !detectTypeScriptUsage() {
		t.Error("detectTypeScriptUsage should return true for projects with typescript dependency")
	}
}

func TestDetectTypeScriptUsage_JavaScriptOnly(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tempDir)

	// Create a JavaScript file and package.json without typescript
	pkg := map[string]interface{}{
		"name": "test",
	}
	pkgData, _ := json.Marshal(pkg)
	_ = os.WriteFile("package.json", pkgData, 0o644)
	_ = os.WriteFile("index.js", []byte("console.log('hello');"), 0o644)

	if detectTypeScriptUsage() {
		t.Error("detectTypeScriptUsage should return false for pure JavaScript projects")
	}
}

func TestWriteESLintConfig_TypeScriptProject(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tempDir)

	// Create TypeScript project
	pkg := map[string]interface{}{
		"devDependencies": map[string]interface{}{
			"typescript": "^4.0.0",
		},
	}
	pkgData, _ := json.Marshal(pkg)
	_ = os.WriteFile("package.json", pkgData, 0o644)

	err := writeESLintConfig(false)
	if err != nil {
		t.Fatalf("writeESLintConfig failed: %v", err)
	}
	data, err := os.ReadFile("eslint.config.mjs")
	if err != nil {
		t.Fatalf("eslint.config.mjs not written: %v", err)
	}
	if !strings.Contains(string(data), "@gv-tech/eslint-config") {
		t.Error("@gv-tech/eslint-config should be imported")
	}
	if !strings.Contains(string(data), "recommended") {
		t.Error("recommended config should be used for TypeScript projects")
	}
}

func TestWriteESLintConfig_JavaScriptProject(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tempDir)

	// Create JavaScript-only project
	pkg := map[string]interface{}{
		"name": "test",
	}
	pkgData, _ := json.Marshal(pkg)
	_ = os.WriteFile("package.json", pkgData, 0o644)
	_ = os.WriteFile("index.js", []byte("console.log('hello');"), 0o644)

	err := writeESLintConfig(false)
	if err != nil {
		t.Fatalf("writeESLintConfig failed: %v", err)
	}
	data, err := os.ReadFile("eslint.config.mjs")
	if err != nil {
		t.Fatalf("eslint.config.mjs not written: %v", err)
	}
	if strings.Contains(string(data), "...typescript") {
		t.Error("typescript config should not be used for JavaScript projects")
	}
	// Should contain the new javascriptRecommended preset
	if !strings.Contains(string(data), "javascriptRecommended") {
		t.Error("javascriptRecommended config should be included for JavaScript projects")
	}
}

func TestUpdatePackageJSON_OrderAndValues(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tempDir)
	pkg := map[string]interface{}{
		"name": "testpkg",
		"foo":  "bar",
	}
	data, _ := json.Marshal(pkg)
	_ = os.WriteFile("package.json", data, 0o644)
	if err := updatePackageJSON(); err != nil {
		t.Fatalf("updatePackageJSON failed: %v", err)
	}
	out, _ := os.ReadFile("package.json")
	if !strings.Contains(string(out), "\"scripts\":") {
		t.Error("scripts not present in package.json")
	}
	if !strings.Contains(string(out), "\"lint-staged\":") {
		t.Error("lint-staged not present in package.json")
	}
	if !strings.Contains(string(out), "\"prettier\":") {
		t.Error("prettier not present in package.json")
	}
	// Check order: name should come before scripts
	nameIdx := strings.Index(string(out), "\"name\"")
	scriptsIdx := strings.Index(string(out), "\"scripts\"")
	if nameIdx == -1 || scriptsIdx == -1 || nameIdx > scriptsIdx {
		t.Error("fields not in expected order")
	}
	// Extra key should be present
	if !strings.Contains(string(out), "foo") {
		t.Error("extra key not preserved")
	}
}

func TestSetupHusky(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tempDir)
	// Mock execCommand
	oldCommand := execCommand
	defer func() { execCommand = oldCommand }()
	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("echo", append([]string{name}, arg...)...)
	}
	_ = os.MkdirAll(".husky", 0o755)
	// Remove .husky/pre-commit if it exists to simulate fresh state
	_ = os.Remove(".husky/pre-commit")
	if err := setupHusky(); err != nil {
		t.Fatalf("setupHusky failed: %v", err)
	}
	data, err := os.ReadFile(".husky/pre-commit")
	if err != nil {
		t.Fatalf("pre-commit hook not written: %v", err)
	}
	if !strings.Contains(string(data), "npx lint-staged") {
		t.Error("pre-commit hook missing lint-staged command")
	}
}

func TestInstallLintDependencies_PeerDepsRetry(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tempDir)
	_ = os.WriteFile("package.json", []byte(`{"name":"test"}`), 0o644)
	called := 0
	oldCommand := execCommand
	defer func() { execCommand = oldCommand }()
	execCommand = func(name string, arg ...string) *exec.Cmd {
		called++
		if called == 1 {
			// Simulate peer dep error on first call
			return exec.Command("sh", "-c", "echo 'could not resolve dependency --legacy-peer-deps' 1>&2; exit 1")
		}
		// On retry, simulate success
		return exec.Command("sh", "-c", "exit 0")
	}
	err := installLintDependencies(false)
	if err != nil {
		t.Fatalf("installLintDependencies should succeed after retry, got: %v", err)
	}
	if called < 2 {
		t.Error("expected retry with --legacy-peer-deps")
	}
}

func TestInstallLintDependencies_UsesYarnIfPresent(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tempDir)
	// Create a yarn.lock file to trigger yarn usage
	_ = os.WriteFile("yarn.lock", []byte(""), 0o644)
	called := struct{ npm, yarn int }{}
	oldCommand := execCommand
	defer func() { execCommand = oldCommand }()
	execCommand = func(name string, arg ...string) *exec.Cmd {
		switch name {
		case "yarn":
			called.yarn++
		case "npm":
			called.npm++
		}
		return exec.Command("echo", append([]string{name}, arg...)...)
	}
	err := installLintDependencies(false)
	if err != nil {
		t.Fatalf("installLintDependencies with yarn.lock should succeed, got: %v", err)
	}
	if called.yarn == 0 {
		t.Error("expected yarn to be called when yarn.lock is present")
	}
	if called.npm != 0 {
		t.Error("did not expect npm to be called when yarn.lock is present")
	}
}

func TestInstallLintDependencies_UsesNpmIfNoYarnLock(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tempDir)
	called := struct{ npm, yarn int }{}
	oldCommand := execCommand
	defer func() { execCommand = oldCommand }()
	execCommand = func(name string, arg ...string) *exec.Cmd {
		switch name {
		case "yarn":
			called.yarn++
		case "npm":
			called.npm++
		}
		return exec.Command("echo", append([]string{name}, arg...)...)
	}
	err := installLintDependencies(false)
	if err != nil {
		t.Fatalf("installLintDependencies with no yarn.lock should succeed, got: %v", err)
	}
	if called.npm == 0 {
		t.Error("expected npm to be called when yarn.lock is not present")
	}
	if called.yarn != 0 {
		t.Error("did not expect yarn to be called when yarn.lock is not present")
	}
}
func TestCheckRedundantDependencies(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tempDir)

	pkg := map[string]interface{}{
		"devDependencies": map[string]interface{}{
			"eslint-plugin-prettier": "latest",
			"globals":                "latest",
		},
	}
	pkgData, _ := json.Marshal(pkg)
	_ = os.WriteFile("package.json", pkgData, 0o644)

	// Since it just logs, we could try to capture logs if we had a logger mock,
	// but for now we just ensure it doesn't crash and we can visually verify output in -v mode.
	checkRedundantDependencies()
}
