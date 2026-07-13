package project

import (
	"context"

	"github.com/eng618/eng/internal/repo"
)

// RepoClient defines the interface for repository operations to allow testing.
type RepoClient interface {
	Clone(ctx context.Context, url, path string) error
	IsDirty(ctx context.Context, repoPath string) (bool, error)
	PullLatestCode(ctx context.Context, repoPath string) error
	FetchAllPrune(ctx context.Context, repoPath string) error
}

// defaultRepoClient provides the standard implementation using internal/repo.
type defaultRepoClient struct{}

func (d *defaultRepoClient) Clone(ctx context.Context, url, path string) error {
	return repo.Clone(ctx, url, path)
}

func (d *defaultRepoClient) IsDirty(ctx context.Context, repoPath string) (bool, error) {
	return repo.IsDirty(ctx, repoPath)
}

func (d *defaultRepoClient) PullLatestCode(ctx context.Context, repoPath string) error {
	return repo.PullLatestCode(ctx, repoPath)
}

func (d *defaultRepoClient) FetchAllPrune(ctx context.Context, repoPath string) error {
	return repo.FetchAllPrune(ctx, repoPath)
}
