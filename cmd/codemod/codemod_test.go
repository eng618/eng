package codemod

import (
	"os/exec"
	"testing"
)

// TestCodemodCmd_Integration tests the main codemod command structure
func TestCodemodCmd_Integration(t *testing.T) {
	// Test that the main command has the expected subcommands
	subcommands := CodemodCmd.Commands()

	expectedCommands := map[string]bool{
		"lint-setup": false,
		"copilot":    false,
	}

	for _, cmd := range subcommands {
		if _, exists := expectedCommands[cmd.Use]; exists {
			expectedCommands[cmd.Use] = true
		}
	}

	for cmdName, found := range expectedCommands {
		if !found {
			t.Errorf("Expected subcommand %s not found", cmdName)
		}
	}
}

// TestExecCommand_MockOverride tests that execCommand can be overridden for testing
func TestExecCommand_MockOverride(t *testing.T) {
	original := execCommand
	defer func() { execCommand = original }()

	// Override execCommand
	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("echo", "mocked")
	}

	// Test that the override works
	cmd := execCommand("test", "args")
	if cmd.Args[0] != "echo" {
		t.Error("execCommand override did not work as expected")
	}
}
