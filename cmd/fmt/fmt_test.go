package fmtcmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eng618/eng/internal/oxfmt"
)

func TestFmtCmdDefinition(t *testing.T) {
	assert.Equal(t, "fmt [path]", FmtCmd.Use)
	assert.NotEmpty(t, FmtCmd.Short)
	assert.NotEmpty(t, FmtCmd.Long)
	assert.NotEmpty(t, FmtCmd.Example)
	require.NoError(t, FmtCmd.Args(FmtCmd, nil))
	require.NoError(t, FmtCmd.Args(FmtCmd, []string{"web"}))
	assert.Error(t, FmtCmd.Args(FmtCmd, []string{"a", "b"}))
}

func TestFmtCmdFlagDefaults(t *testing.T) {
	flags := FmtCmd.Flags()
	write, err := flags.GetBool("write")
	require.NoError(t, err)
	assert.False(t, write)
	check, err := flags.GetBool("check")
	require.NoError(t, err)
	assert.False(t, check)
	diff, err := flags.GetBool("diff")
	require.NoError(t, err)
	assert.False(t, diff)
	runner, err := flags.GetString("runner")
	require.NoError(t, err)
	assert.Equal(t, oxfmt.RunnerAuto, runner)
	config, err := flags.GetString("config")
	require.NoError(t, err)
	assert.Empty(t, config)
}
