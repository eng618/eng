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
