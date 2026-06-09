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

	var buf bytes.Buffer
	FilesCmd.SetOut(&buf)

	// Execute the Run function directly
	FilesCmd.Run(FilesCmd, []string{})

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
