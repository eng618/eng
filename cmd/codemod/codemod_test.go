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
	if err := os.WriteFile("package.json", data, 0644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}
	// Mock exec.Command to avoid running npm/husky
	oldCommand := execCommand
	defer func() { execCommand = oldCommand }()
	execCommand = func(name string, arg ...string) *exec.Cmd {
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

func TestWriteESLintConfig(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tempDir)

	err := writeESLintConfig()
	if err != nil {
		t.Fatalf("writeESLintConfig failed: %v", err)
	}
	data, err := os.ReadFile("eslint.config.mjs")
	if err != nil {
		t.Fatalf("eslint.config.mjs not written: %v", err)
	}
	if !strings.Contains(string(data), "echo-eslint-config") {
		t.Error("eslint.config.mjs missing expected content")
	}
}

func TestUpdatePackageJSON_OrderAndValues(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tempDir)
	pkg := map[string]interface{}{
		"name": "testpkg",
		"foo": "bar",
	}
	data, _ := json.Marshal(pkg)
	_ = os.WriteFile("package.json", data, 0644)
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
	_ = os.MkdirAll(".husky", 0755)
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
	_ = os.WriteFile("package.json", []byte(`{"name":"test"}`), 0644)
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
	err := installLintDependencies()
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
	_ = os.WriteFile("yarn.lock", []byte(""), 0644)
	called := struct{ npm, yarn int }{}
	oldCommand := execCommand
	defer func() { execCommand = oldCommand }()
	execCommand = func(name string, arg ...string) *exec.Cmd {
		if name == "yarn" {
			called.yarn++
		} else if name == "npm" {
			called.npm++
		}
		return exec.Command("echo", append([]string{name}, arg...)...)
	}
	err := installLintDependencies()
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
		if name == "yarn" {
			called.yarn++
		} else if name == "npm" {
			called.npm++
		}
		return exec.Command("echo", append([]string{name}, arg...)...)
	}
	err := installLintDependencies()
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
