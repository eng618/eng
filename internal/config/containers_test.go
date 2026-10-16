package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateComposeStackName(t *testing.T) {
	for _, name := range []string{"default", "media", "arr-1", "a_b"} {
		require.NoError(t, ValidateComposeStackName(name))
	}
	for _, name := range []string{"", "-bad", "has space", "bad!"} {
		require.Error(t, ValidateComposeStackName(name))
	}
}

func TestComposeStacks_AddRemoveRoundtrip(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	require.NoError(t, AddComposeStack(ComposeStackEntry{Name: "Media", Path: "/tmp/media"}))
	// Case-insensitive update, not duplicate.
	require.NoError(t, AddComposeStack(ComposeStackEntry{Name: "media", Path: "/tmp/media2"}))
	stacks := GetComposeStacks()
	require.Len(t, stacks, 1)
	require.Equal(t, "/tmp/media2", stacks[0].Path)

	require.Error(t, AddComposeStack(ComposeStackEntry{Name: "bad name!", Path: "/tmp/x"}))
	require.Error(t, AddComposeStack(ComposeStackEntry{Name: "ok", Path: "  "}))

	require.NoError(t, RemoveComposeStack("MEDIA"))
	require.Empty(t, GetComposeStacks())
	require.Error(t, RemoveComposeStack("missing"))
}

func TestValidate_ComposeStacks(t *testing.T) {
	tmp := t.TempDir()
	c := &ResolvedConfig{
		DotfilesBranch: "main",
		GitDevPath:     tmp,
		ContainersStacks: []ComposeStackEntry{
			{Name: "good", Path: tmp},
			{Name: "good", Path: tmp},
			{Name: "bad name", Path: ""},
		},
	}
	errs := c.Validate()
	fields := make([]string, 0, len(errs))
	for _, e := range errs {
		fields = append(fields, e.Field)
	}
	require.Contains(t, fields, "containers.stacks[1].name")
	require.Contains(t, fields, "containers.stacks[2].name")
	require.Contains(t, fields, "containers.stacks[2].path")
}

func TestSelectComposeStack(t *testing.T) {
	_, err := SelectComposeStack(nil)
	require.Error(t, err)

	oldSelect := SelectPrompt
	defer func() { SelectPrompt = oldSelect }()
	SelectPrompt = func(_ string, options []string, _ string) (string, error) {
		require.Equal(t, []string{"media", "homelab"}, options)
		return "homelab", nil
	}
	idx, err := SelectComposeStack([]ComposeStackEntry{
		{Name: "media", Path: "/tmp/media"},
		{Name: "homelab", Path: "/tmp/homelab"},
	})
	require.NoError(t, err)
	require.Equal(t, 1, idx)
}

func TestValidateComposeStackPath(t *testing.T) {
	require.Error(t, validateComposeStackPath("  "))
	require.Error(t, validateComposeStackPath(t.TempDir()+"/does-not-exist"))
	require.NoError(t, validateComposeStackPath(t.TempDir()))
}
