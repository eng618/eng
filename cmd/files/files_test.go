package files

import (
	"bytes"
	"strings"
	"testing"
)

func TestFilesCmdProperties(t *testing.T) {
	if FilesCmd.Use != "files" {
		t.Errorf("Expected Use to be 'files', got '%s'", FilesCmd.Use)
	}

	if FilesCmd.Short == "" {
		t.Error("Expected Short description to be set, but it was empty")
	}

	if FilesCmd.Long == "" {
		t.Error("Expected Long description to be set, but it was empty")
	}

	// Test Run function
	if FilesCmd.Run == nil {
		t.Error("Expected Run function to be set")
	}

	// Execute with empty args to test Run behavior
	var buf bytes.Buffer
	FilesCmd.SetOut(&buf)
	FilesCmd.SetArgs([]string{})
	if err := FilesCmd.Execute(); err != nil {
		t.Errorf("Expected execute with empty args to not return error, got %v", err)
	}

	// Check that Help was called (which writes to the output buffer by default)
	output := buf.String()
	if !strings.Contains(output, "Usage:") {
		t.Errorf("Expected help output to be written (containing 'Usage:'), but got: %s", output)
	}
}

func TestFilesCmdSubCommands(t *testing.T) {
	commands := FilesCmd.Commands()

	hasFindAndDelete := false
	hasFindNonMovieFolders := false

	for _, cmd := range commands {
		if cmd.Use == FindAndDeleteCmd.Use {
			hasFindAndDelete = true
		}
		if cmd.Use == FindNonMovieFoldersCmd.Use {
			hasFindNonMovieFolders = true
		}
	}

	if !hasFindAndDelete {
		t.Error("Expected FindAndDeleteCmd to be a subcommand of FilesCmd")
	}

	if !hasFindNonMovieFolders {
		t.Error("Expected FindNonMovieFoldersCmd to be a subcommand of FilesCmd")
	}
}

func TestFilesCmdFlags(t *testing.T) {
	// Test FindAndDeleteCmd flags
	expectedFADFlags := []string{"glob", "ext", "filename", "list-extensions"}
	for _, flagName := range expectedFADFlags {
		flag := FindAndDeleteCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag '%s' on FindAndDeleteCmd", flagName)
		}
	}

	// Test FindNonMovieFoldersCmd flags
	expectedFNMFlags := []string{"dry-run"}
	for _, flagName := range expectedFNMFlags {
		flag := FindNonMovieFoldersCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag '%s' on FindNonMovieFoldersCmd", flagName)
		} else if flag.DefValue != "true" {
			t.Errorf("Expected default value for flag '%s' to be 'true', got '%s'", flagName, flag.DefValue)
		}
	}
}
