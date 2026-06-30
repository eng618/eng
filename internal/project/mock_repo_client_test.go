package project

import (
	"context"
)

type MockRepoClient struct {
	CloneFunc          func(ctx context.Context, url, path string) error
	IsDirtyFunc        func(ctx context.Context, repoPath string) (bool, error)
	PullLatestCodeFunc func(ctx context.Context, repoPath string) error
	FetchAllPruneFunc  func(ctx context.Context, repoPath string) error
}

func (m *MockRepoClient) Clone(ctx context.Context, url, path string) error {
	if m.CloneFunc != nil {
		return m.CloneFunc(ctx, url, path)
	}
	return nil
}

func (m *MockRepoClient) IsDirty(ctx context.Context, repoPath string) (bool, error) {
	if m.IsDirtyFunc != nil {
		return m.IsDirtyFunc(ctx, repoPath)
	}
	return false, nil
}

func (m *MockRepoClient) PullLatestCode(ctx context.Context, repoPath string) error {
	if m.PullLatestCodeFunc != nil {
		return m.PullLatestCodeFunc(ctx, repoPath)
	}
	return nil
}

func (m *MockRepoClient) FetchAllPrune(ctx context.Context, repoPath string) error {
	if m.FetchAllPruneFunc != nil {
		return m.FetchAllPruneFunc(ctx, repoPath)
	}
	return nil
}
