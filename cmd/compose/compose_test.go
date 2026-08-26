package compose

import (
	"bytes"
	"testing"
)

func TestComposeCommandStructure(t *testing.T) {
	buf := new(bytes.Buffer)
	ComposeCmd.SetOut(buf)
	ComposeCmd.SetArgs([]string{"--help"})

	err := ComposeCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error executing compose --help: %v", err)
	}

	output := buf.String()
	expectedSubcommands := []string{"list", "up", "down", "pull", "status", "logs", "clean"}
	for _, sub := range expectedSubcommands {
		if !bytes.Contains([]byte(output), []byte(sub)) {
			t.Errorf("expected subcommand %q in help output", sub)
		}
	}
}
