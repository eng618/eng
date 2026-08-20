package immich

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTopLevelImmichCmd_Structure(t *testing.T) {
	assert.Equal(t, "immich", ImmichCmd.Use)
	assert.Contains(t, ImmichCmd.Aliases, "photos")
	assert.NotEmpty(t, ImmichCmd.Commands())
}
