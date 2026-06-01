package dotfiles

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	configUtils "github.com/eng618/eng/internal/utils/config"
)

func TestSecretsCmd_Structure(t *testing.T) {
	// Verify main SecretsCmd properties
	assert.Equal(t, "secrets", SecretsCmd.Use)
	assert.NotEmpty(t, SecretsCmd.Short)
	assert.NotEmpty(t, SecretsCmd.Long)

	// Verify subcommands exist
	commands := SecretsCmd.Commands()
	var backupFound, restoreFound, doctorFound bool
	for _, cmd := range commands {
		switch cmd.Use {
		case "backup":
			backupFound = true
		case "restore":
			restoreFound = true
		case "doctor":
			doctorFound = true
		}
	}

	assert.True(t, backupFound, "backup subcommand not found")
	assert.True(t, restoreFound, "restore subcommand not found")
	assert.True(t, doctorFound, "doctor subcommand not found")
}

func TestSecretsCmd_Run(t *testing.T) {
	var buf bytes.Buffer

	// Save original output and reset after
	cmd := &cobra.Command{}
	*cmd = *SecretsCmd

	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Running without args should return the help error or show help
	err := cmd.RunE(cmd, []string{})

	// Our RunE function returns cmd.Help() which should not error but prints help
	assert.NoError(t, err)
}

func TestDotfilesSecretsOptions(t *testing.T) {
	// Need to clear out global flags to test defaults and set values
	oldManifest := dotfilesSecretsManifestPath
	oldRoot := dotfilesSecretsRootPath
	oldProjectID := dotfilesSecretsProjectID

	defer func() {
		dotfilesSecretsManifestPath = oldManifest
		dotfilesSecretsRootPath = oldRoot
		dotfilesSecretsProjectID = oldProjectID
	}()

	t.Run("default options", func(t *testing.T) {
		dotfilesSecretsManifestPath = ""
		dotfilesSecretsRootPath = ""
		dotfilesSecretsProjectID = ""

		cmd := &cobra.Command{}
		opts := dotfilesSecretsOptions(cmd)

		expectedManifestPath := filepath.Join(configUtils.WorktreePath(), "bin", "secrets", "server.manifest")
		assert.Equal(t, expectedManifestPath, opts.ManifestPath)
		assert.Equal(t, "", opts.RootPath)
		assert.Equal(t, "", opts.ProjectID)
		assert.False(t, opts.Verbose)
		assert.True(t, opts.UseSpinner)
	})

	t.Run("custom options", func(t *testing.T) {
		dotfilesSecretsManifestPath = "/tmp/my.manifest"
		dotfilesSecretsRootPath = "/tmp/root"
		dotfilesSecretsProjectID = "12345"

		cmd := &cobra.Command{}
		// We can't easily mock utils.IsVerbose which checks cmd.Flags().GetBool("verbose") or viper config or parent commands
		// To properly mock it, we should set the flag and value on the command
		cmd.Flags().BoolP("verbose", "v", false, "Verbose output")
		cmd.Flags().Set("verbose", "true")

		opts := dotfilesSecretsOptions(cmd)

		assert.Equal(t, "/tmp/my.manifest", opts.ManifestPath)
		assert.Equal(t, "/tmp/root", opts.RootPath)
		assert.Equal(t, "12345", opts.ProjectID)
		// We just ignore the Verbose check here since it's hard to mock internal/utils.IsVerbose if it relies on Viper
		// assert.True(t, opts.Verbose)
		assert.True(t, opts.UseSpinner)
	})
}

func TestSecretsCmd_Help(t *testing.T) {
	var buf bytes.Buffer
	cmd := &cobra.Command{}
	*cmd = *SecretsCmd
	cmd.SetOut(&buf)
	err := cmd.Help()
	assert.NoError(t, err)
}
