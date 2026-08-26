package system

import (
	"bytes"
	"testing"
)

func TestCleanCommandHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	SystemCmd.SetOut(buf)
	SystemCmd.SetArgs([]string{"clean", "--help"})

	err := SystemCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error executing system clean --help: %v", err)
	}

	output := buf.String()
	expectedFlags := []string{"--docker", "--all-images", "--older-than", "--journal", "--journal-size", "--packages", "--asdf", "--brew", "--dry-run", "--yes"}
	for _, flag := range expectedFlags {
		if !bytes.Contains([]byte(output), []byte(flag)) {
			t.Errorf("expected flag %q in clean help output", flag)
		}
	}
}
