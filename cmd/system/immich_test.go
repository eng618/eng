package system

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImmichCmd_Structure(t *testing.T) {
	assert.Equal(t, "immich", ImmichCmd.Use)
	assert.Contains(t, ImmichCmd.Aliases, "photos")

	subCmds := ImmichCmd.Commands()
	subNames := make([]string, 0, len(subCmds))
	for _, c := range subCmds {
		subNames = append(subNames, c.Name())
	}

	assert.Contains(t, subNames, "status")
	assert.Contains(t, subNames, "backup")
	assert.Contains(t, subNames, "restore")
	assert.Contains(t, subNames, "start")
	assert.Contains(t, subNames, "stop")
	assert.Contains(t, subNames, "restart")
	assert.Contains(t, subNames, "logs")
}

func TestImmichCmd_Help(t *testing.T) {
	assert.NotEmpty(t, ImmichCmd.Short)
	assert.NotEmpty(t, ImmichCmd.Long)
	assert.Contains(t, ImmichCmd.Short, "Immich")
	assert.Contains(t, ImmichCmd.Long, "Immich")

	// Test subcommands have short descriptions
	for _, c := range ImmichCmd.Commands() {
		assert.NotEmpty(t, c.Short, "subcommand %s should have short description", c.Name())
	}
}
