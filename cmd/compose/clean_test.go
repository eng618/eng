package compose

import (
	"bytes"
	"testing"
)

func TestCleanCommandHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	ComposeCmd.SetOut(buf)
	ComposeCmd.SetArgs([]string{"clean", "--help"})

	err := ComposeCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error executing compose clean --help: %v", err)
	}

	output := buf.String()
	expectedFlags := []string{"--all", "--older-than", "--build-cache", "--volumes", "--dry-run", "--yes"}
	for _, flag := range expectedFlags {
		if !bytes.Contains([]byte(output), []byte(flag)) {
			t.Errorf("expected flag %q in clean help output", flag)
		}
	}
}
