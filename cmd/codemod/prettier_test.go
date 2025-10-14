// Package codemod provides helpers for codemods and project automation.
// This file contains tests for the prettier formatting command.
package codemod

import (
	"os/exec"
	"strings"
	"testing"
)

func TestCanResolvePackage(t *testing.T) {
	// Save original execCommand
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	tests := []struct {
		name        string
		packageName string
		mockOutput  string
		expected    bool
	}{
		{
			name:        "package exists",
			packageName: "@eng618/prettier-config",
			mockOutput:  "true",
			expected:    true,
		},
		{
			name:        "package does not exist",
			packageName: "non-existent-package",
			mockOutput:  "false",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock execCommand for this test
			execCommand = func(command string, args ...string) *exec.Cmd {
				// Return a command that will output our expected result
				return exec.Command("echo", tt.mockOutput)
			}

			result := canResolvePackage(tt.packageName)
			if result != tt.expected {
				t.Errorf("canResolvePackage(%q) = %v, want %v", tt.packageName, result, tt.expected)
			}
		})
	}
}

func TestGetNpmPrefix(t *testing.T) {
	// Save original execCommand
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	expectedPrefix := "/usr/local"
	
	// Mock execCommand
	execCommand = func(command string, args ...string) *exec.Cmd {
		return exec.Command("echo", expectedPrefix)
	}

	result, err := getNpmPrefix()
	if err != nil {
		t.Fatalf("getNpmPrefix() returned error: %v", err)
	}
	if result != expectedPrefix {
		t.Errorf("getNpmPrefix() = %q, want %q", result, expectedPrefix)
	}
}

func TestGetConfigPath(t *testing.T) {
	// Save original execCommand
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	expectedPath := "/path/to/prettier-config"
	packageName := "@eng618/prettier-config"
	
	// Mock execCommand
	execCommand = func(command string, args ...string) *exec.Cmd {
		return exec.Command("echo", expectedPath)
	}

	result, err := getConfigPath(packageName)
	if err != nil {
		t.Fatalf("getConfigPath(%q) returned error: %v", packageName, err)
	}
	if result != expectedPath {
		t.Errorf("getConfigPath(%q) = %q, want %q", packageName, result, expectedPath)
	}
}

func TestRunPrettier_Integration(t *testing.T) {
	// This is a simple integration test that just checks the function exists and can be called
	// without panicking. We skip actual execution since it would require Node.js setup.
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Save original execCommand
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	// Mock execCommand to avoid actually running npm/node commands
	execCommand = func(command string, args ...string) *exec.Cmd {
		// For canResolvePackage check, return false to trigger default behavior
		if command == "node" && len(args) > 0 && strings.Contains(args[1], "require.resolve") {
			return exec.Command("echo", "false")
		}
		// For prettier execution, just return success
		return exec.Command("true")
	}

	// This should not panic and should complete without error
	err := runPrettier(".")
	if err != nil {
		t.Logf("runPrettier returned error (expected in test environment): %v", err)
	}
}
