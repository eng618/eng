package project

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultRepoClient(t *testing.T) {
	client := &defaultRepoClient{}
	ctx := context.Background()
	path := t.TempDir()

	err := client.Clone(ctx, "invalid-url", path)
	assert.Error(t, err)

	_, err = client.IsDirty(ctx, path)
	assert.Error(t, err)

	err = client.PullLatestCode(ctx, path)
	assert.Error(t, err)

	err = client.FetchAllPrune(ctx, path)
	assert.Error(t, err)
}

func TestFetchWithOptions_Passthrough(t *testing.T) {
	oldConfirm := ConfirmPrompt
	defer func() { ConfirmPrompt = oldConfirm }()
	ConfirmPrompt = func(string, bool) (bool, error) {
		t.Error("ConfirmPrompt must not fire for non-clobber errors")
		return false, nil
	}

	client := &defaultRepoClient{}
	ctx := context.Background()

	// Force flag delegates straight to force-fetch (errors on non-repo path).
	assert.Error(t, client.FetchWithOptions(ctx, t.TempDir(), true))
	// Non-clobber failures pass through without prompting.
	assert.Error(t, client.FetchWithOptions(ctx, t.TempDir(), false))
}
