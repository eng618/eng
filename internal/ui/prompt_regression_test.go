package ui

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPromptVars_Wired(t *testing.T) {
	// Regression: all prompt entry points must be non-nil vars so tests can stub them.
	require.NotNil(t, Confirm)
	require.NotNil(t, ConfirmDanger)
	require.NotNil(t, Input)
	require.NotNil(t, Select)
	require.NotNil(t, SelectWithFilter)
	require.NotNil(t, MultiSelect)
	require.NotNil(t, Password)
}
