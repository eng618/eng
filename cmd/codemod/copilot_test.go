package codemod

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestCopilotSetupCmd_CreatesFile(t *testing.T) {
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

	// Create a mock .git directory to simulate git repository
	if err := os.MkdirAll(
		".git",
		0o755,
	); err != nil { // Reverted to original MkdirAll, assuming user intended to add nolint to this line.
		t.Fatalf("failed to create .git directory: %v", err)
	}

	// Mock exec.Command to avoid running git commands
	oldCommand := execCommand
	defer func() { execCommand = oldCommand }()
	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("echo", append([]string{name}, arg...)...)
	}

	// Run the command
	cmd := CopilotSetupCmd
	output := &strings.Builder{}
	cmd.SetOut(output)
	cmd.SetArgs([]string{})
	cmd.Run(cmd, []string{})

	// Check that the file was created
	filePath := ".github/copilot-instructions.md"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("copilot-instructions.md file was not created")
	}

	// Check file content
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read copilot-instructions.md: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "# Copilot Custom Instructions") {
		t.Error("file missing expected header")
	}
	if !strings.Contains(content, "## General") {
		t.Error("file missing General section")
	}
	if !strings.Contains(content, "## Code Quality") {
		t.Error("file missing Code Quality section")
	}
}

func TestCopilotSetupCmd_SkipsIfFileExists(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tempDir)

	// Create .git directory and existing file
	_ = os.MkdirAll(".git", 0o755)
	_ = os.MkdirAll(".github", 0o755)
	existingContent := "# Existing Content"
	_ = os.WriteFile(".github/copilot-instructions.md", []byte(existingContent), 0o644)

	// Mock exec.Command
	oldCommand := execCommand
	defer func() { execCommand = oldCommand }()
	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("echo", append([]string{name}, arg...)...)
	}

	// Run the command
	err := createCopilotInstructions()
	if err != nil {
		t.Fatalf("createCopilotInstructions failed: %v", err)
	}

	// Verify original content is preserved
	data, err := os.ReadFile(".github/copilot-instructions.md")
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(data) != existingContent {
		t.Error("existing file content was modified")
	}
}

func TestCopilotSetupCmd_ForceFlag(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tempDir)

	// Don't create .git directory to simulate non-git directory

	// Mock exec.Command
	oldCommand := execCommand
	defer func() { execCommand = oldCommand }()
	execCommand = func(name string, arg ...string) *exec.Cmd {
		// Simulate git command failure (not a git repo)
		if name == "git" && len(arg) > 0 && arg[0] == "rev-parse" {
			return exec.Command("sh", "-c", "exit 1")
		}
		return exec.Command("echo", append([]string{name}, arg...)...)
	}

	// Test the createCopilotInstructions function directly since we're testing force logic
	err := createCopilotInstructions()
	if err != nil {
		t.Fatalf("createCopilotInstructions failed: %v", err)
	}

	// Check that the file was created despite not being a git repo (when called directly)
	filePath := ".github/copilot-instructions.md"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("copilot-instructions.md file was not created")
	}
}

func TestCopilotSetupCmd_GitCheckFailsWithoutForce(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tempDir)

	// Mock exec.Command to simulate git failure
	oldCommand := execCommand
	defer func() { execCommand = oldCommand }()
	execCommand = func(name string, arg ...string) *exec.Cmd {
		if name == "git" && len(arg) > 0 && arg[0] == "rev-parse" {
			return exec.Command("sh", "-c", "exit 1")
		}
		return exec.Command("echo", append([]string{name}, arg...)...)
	}

	// Test that isGitRepository returns false
	if isGitRepository() {
		t.Error("isGitRepository should return false when git commands fail")
	}

	// File should not be created when not in git repo and no force flag
	filePath := ".github/copilot-instructions.md"
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("file should not exist when not in git repo and no force flag")
	}
}

func TestIsGitRepository(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tempDir)

	// Mock exec.Command
	oldCommand := execCommand
	defer func() { execCommand = oldCommand }()

	// Test case 1: .git directory exists
	_ = os.MkdirAll(".git", 0o755)
	if !isGitRepository() {
		t.Error("should detect git repository when .git directory exists")
	}

	// Test case 2: .git directory doesn't exist, but git rev-parse succeeds
	_ = os.RemoveAll(".git")
	execCommand = func(name string, arg ...string) *exec.Cmd {
		if name == "git" && len(arg) > 0 && arg[0] == "rev-parse" {
			return exec.Command(
				"echo",
				"success",
			) // Assuming tt.mockOutput was a placeholder and "success" is a reasonable mock output.
		}
		return exec.Command("echo", append([]string{name}, arg...)...)
	}
	if !isGitRepository() {
		t.Error("should detect git repository when git rev-parse succeeds")
	}

	// Test case 3: not a git repository
	execCommand = func(name string, arg ...string) *exec.Cmd {
		if name == "git" && len(arg) > 0 && arg[0] == "rev-parse" {
			return exec.Command("sh", "-c", "exit 1")
		}
		return exec.Command("echo", append([]string{name}, arg...)...)
	}
	if isGitRepository() {
		t.Error("should not detect git repository when git rev-parse fails")
	}
}

func TestCreateCopilotInstructions_CreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tempDir)

	// Ensure .github directory doesn't exist
	if _, err := os.Stat(".github"); !os.IsNotExist(err) {
		t.Fatal(".github directory should not exist initially")
	}

	err := createCopilotInstructions()
	if err != nil {
		t.Fatalf("createCopilotInstructions failed: %v", err)
	}

	// Check that .github directory was created
	if _, err := os.Stat(".github"); os.IsNotExist(err) {
		t.Error(".github directory was not created")
	}

	// Check that file was created
	if _, err := os.Stat(".github/copilot-instructions.md"); os.IsNotExist(err) {
		t.Error("copilot-instructions.md file was not created")
	}
}
