package dotfiles

import (
	"context"
	"errors"
	"testing"
)

func TestSyncRepo(t *testing.T) {
	// Keep references to the original functions so we can restore them
	origFetchRepo := FetchRepo
	origPullRebaseRepo := PullRebaseRepo
	defer func() {
		FetchRepo = origFetchRepo
		PullRebaseRepo = origPullRebaseRepo
	}()

	tests := []struct {
		name                 string
		fetchErr             error
		pullRebaseErr        error
		expectedErr          error
		expectPullRebaseCalled bool
	}{
		{
			name:                 "Success case",
			fetchErr:             nil,
			pullRebaseErr:        nil,
			expectedErr:          nil,
			expectPullRebaseCalled: true,
		},
		{
			name:                 "Fetch fails",
			fetchErr:             errors.New("fetch failed"),
			pullRebaseErr:        nil,
			expectedErr:          errors.New("fetch failed"),
			expectPullRebaseCalled: false,
		},
		{
			name:                 "Pull rebase fails",
			fetchErr:             nil,
			pullRebaseErr:        errors.New("pull rebase failed"),
			expectedErr:          errors.New("pull rebase failed"),
			expectPullRebaseCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pullRebaseCalled bool

			// Set up mocks
			FetchRepo = func(ctx context.Context, repoPath, worktreePath string) error {
				return tt.fetchErr
			}
			PullRebaseRepo = func(ctx context.Context, repoPath, worktreePath string) error {
				pullRebaseCalled = true
				return tt.pullRebaseErr
			}

			err := SyncRepo(context.Background(), "/mock/repo", "/mock/worktree", false)

			// Assert errors
			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			}

			// Assert if pull rebase was called when expected
			if pullRebaseCalled != tt.expectPullRebaseCalled {
				t.Errorf("expected pullRebaseCalled to be %v, got %v", tt.expectPullRebaseCalled, pullRebaseCalled)
			}
		})
	}
}
