package theme

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigure_NoColorFlag(t *testing.T) {
	Configure(true)
	require.True(t, NoColorEnabled())
	Configure(false)
	t.Setenv("NO_COLOR", "1")
	require.True(t, NoColorEnabled())
}
