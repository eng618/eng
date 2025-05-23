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
