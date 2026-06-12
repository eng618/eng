package system

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestSystemCmdProperties(t *testing.T) {
	assert.Equal(t, "system", SystemCmd.Use)
	assert.Equal(t, "A command for managing the system", SystemCmd.Short)
	assert.Equal(t, "This command will help manage various aspects of MacOS and Linux systems.", SystemCmd.Long)
}

func TestSystemCmdSubcommands(t *testing.T) {
	// Expected subcommands based on the Use field
	expectedSubcommands := []string{
		"killPort",
		"killProcess",
		"update",
		"proxy",
		"compauditFix",
		"setup",
	}

	// Get actual subcommands (first word of Use)
	var actualSubcommands []string
	for _, cmd := range SystemCmd.Commands() {
		actualSubcommands = append(actualSubcommands, strings.Split(cmd.Use, " ")[0])
	}

	// Verify all expected subcommands are present
	for _, expected := range expectedSubcommands {
		assert.Contains(t, actualSubcommands, expected)
	}
}

func TestSystemCmdRun(t *testing.T) {
	SystemCmd.SetOut(new(bytes.Buffer))
	SystemCmd.SetErr(new(bytes.Buffer))
	defer func() {
		SystemCmd.SetOut(nil)
		SystemCmd.SetErr(nil)
	}()

	// Ensure Run function doesn't panic
	assert.NotPanics(t, func() {
		SystemCmd.Run(SystemCmd, []string{})
	})
}
