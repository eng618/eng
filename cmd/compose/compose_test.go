package compose

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

func TestCompleteStackNames(t *testing.T) {
	base := t.TempDir()
	for _, stack := range []string{"media", "arrsenal"} {
		dir := filepath.Join(base, "stacks", stack)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(
			filepath.Join(dir, "docker-compose.yml"),
			[]byte("services:\n  app:\n    image: example/app:latest\n"),
			0o644,
		); err != nil {
			t.Fatal(err)
		}
	}
	viper.Reset()
	viper.Set("containers.path", base)
	defer viper.Reset()

	names, directive := completeStackNames(nil, nil, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("Expected NoFileComp directive, got %v", directive)
	}
	found := map[string]bool{}
	for _, n := range names {
		found[n] = true
	}
	if !found["media"] || !found["arrsenal"] || len(names) != 2 {
		t.Errorf("Expected [media arrsenal] in any order, got %v", names)
	}

	names, _ = completeStackNames(nil, nil, "med")
	if len(names) != 1 || names[0] != "media" {
		t.Errorf("Expected [media] for prefix 'med', got %v", names)
	}

	// Missing containers path → no candidates, no error.
	viper.Set("containers.path", filepath.Join(base, "does-not-exist"))
	if names, _ := completeStackNames(nil, nil, ""); len(names) != 0 {
		t.Errorf("Expected no candidates without stacks, got %v", names)
	}
}
