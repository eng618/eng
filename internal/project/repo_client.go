package project

import (
	"context"
	"errors"

	"github.com/eng618/eng/internal/repo"
	"github.com/eng618/eng/internal/ui"
)

// RepoClient defines the interface for repository operations to allow testing.
type RepoClient interface {
	Clone(ctx context.Context, url, path string) error
	IsDirty(ctx context.Context, repoPath string) (bool, error)
	PullLatestCode(ctx context.Context, repoPath string) error
	FetchAllPrune(ctx context.Context, repoPath string) error
	FetchWithOptions(ctx context.Context, repoPath string, force bool) error
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

func (d *defaultRepoClient) FetchWithOptions(ctx context.Context, repoPath string, force bool) error {
	if force {
		return repo.FetchAllPruneWithForce(ctx, repoPath)
	}
	err := repo.FetchAllPrune(ctx, repoPath)
	if err == nil {
		return nil
	}
	var clobberErr *repo.TagClobberError
	if !errors.As(err, &clobberErr) {
		return err
	}
	// Tag clobber needs a human decision: prompt here (this package
	// already owns presentation) rather than inside internal/repo,
	// which must stay free of UI imports.
	proceed, promptErr := ui.Confirm(
		"Tag clobber detected. Overwrite local tags with remote tags?",
		false,
	)
	if promptErr != nil || !proceed {
		return clobberErr
	}
	return repo.FetchAllPruneWithForce(ctx, repoPath)
}
