package compose

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/eng618/eng/internal/config"
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
	expectedSubcommands := []string{"list", "add", "remove", "up", "down", "pull", "status", "logs", "clean"}
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

// setupIsolatedConfig points global viper at a fresh temp config file.
func setupIsolatedConfig(t *testing.T) {
	t.Helper()
	viper.Reset()
	path := filepath.Join(t.TempDir(), ".eng.yaml")
	require.NoError(t, os.WriteFile(path, []byte(""), 0o600))
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")
	t.Cleanup(viper.Reset)
}

func writeStackDir(t *testing.T, dir string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "docker-compose.yml"),
		[]byte("services:\n  app:\n    image: example/app:latest\n"),
		0o644,
	))
}

func TestAddCmd_DirectArgs(t *testing.T) {
	setupIsolatedConfig(t)
	stackDir := filepath.Join(t.TempDir(), "media")
	writeStackDir(t, stackDir)

	require.NoError(t, addCmd.RunE(addCmd, []string{"media", stackDir}))
	stacks := config.GetComposeStacks()
	require.Len(t, stacks, 1)
	require.Equal(t, "media", stacks[0].Name)
}

func TestAddCmd_WizardPrompt(t *testing.T) {
	setupIsolatedConfig(t)
	stackDir := filepath.Join(t.TempDir(), "homelab")
	writeStackDir(t, stackDir)

	oldPrompt := config.PromptComposeStackValues
	defer func() { config.PromptComposeStackValues = oldPrompt }()
	config.PromptComposeStackValues = func(initialName, initialPath string) (string, string, error) {
		require.Equal(t, "home", initialName)
		require.Equal(t, "", initialPath)
		return "homelab", stackDir, nil
	}

	// Partial args prefill the wizard.
	require.NoError(t, addCmd.RunE(addCmd, []string{"home"}))
	stacks := config.GetComposeStacks()
	require.Len(t, stacks, 1)
	require.Equal(t, "homelab", stacks[0].Name)
}

func TestRemoveCmd_InteractivePicker(t *testing.T) {
	setupIsolatedConfig(t)
	require.NoError(t, config.AddComposeStack(config.ComposeStackEntry{Name: "media", Path: "/tmp/media"}))

	oldSelect := config.SelectPrompt
	defer func() { config.SelectPrompt = oldSelect }()
	config.SelectPrompt = func(_ string, _ []string, _ string) (string, error) {
		return "media", nil
	}

	require.NoError(t, removeCmd.RunE(removeCmd, []string{}))
	require.Empty(t, config.GetComposeStacks())
}
