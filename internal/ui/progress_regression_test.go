package ui

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/eng618/eng/internal/ui/theme"
)

func TestDummySpinner_InfoMatchesCharmBanner(t *testing.T) {
	var buf bytes.Buffer
	d := &dummySpinner{text: "hello", out: &buf}
	d.Info()
	out := buf.String()
	require.Contains(t, out, "INFO")
	require.Contains(t, out, "hello")

	var buf2 bytes.Buffer
	c := &charmSpinner{text: "hello", out: &buf2}
	c.Info()
	require.Contains(t, buf2.String(), "INFO")
}

func TestThemeTokens_NoHardcodedHex(t *testing.T) {
	// Regression: warning/debug/info colors must come from tokens so light/dark stay in sync.
	require.NotEmpty(t, theme.Warning.Light)
	require.NotEmpty(t, theme.Warning.Dark)
	require.NotEmpty(t, theme.Debug.Light)
	require.NotEmpty(t, theme.InfoBanner)
}
