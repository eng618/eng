package repo

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestParsePorcelainStatus(t *testing.T) {
	testCases := []struct {
		name      string
		output    string
		unstaged  int
		staged    int
		untracked int
		conflicts int
	}{
		{
			name:      "clean status",
			output:    "",
			unstaged:  0,
			staged:    0,
			untracked: 0,
			conflicts: 0,
		},
		{
			name:      "unstaged modifications",
			output:    " M file1.txt\n D file2.txt\n",
			unstaged:  2,
			staged:    0,
			untracked: 0,
			conflicts: 0,
		},
		{
			name:      "staged modifications",
			output:    "M  file1.txt\nA  file2.txt\nD  file3.txt\n",
			unstaged:  0,
			staged:    3,
			untracked: 0,
			conflicts: 0,
		},
		{
			name:      "untracked files",
			output:    "?? file1.txt\n?? file2.txt\n",
			unstaged:  0,
			staged:    0,
			untracked: 2,
			conflicts: 0,
		},
		{
			name:      "conflicts",
			output:    "UU file1.txt\nAA file2.txt\nUD file3.txt\n",
			unstaged:  0,
			staged:    0,
			untracked: 0,
			conflicts: 3,
		},
		{
			name:      "mixed status",
			output:    " M file1.txt\nM  file2.txt\n?? file3.txt\nUU file4.txt\nAM file5.txt\n",
			unstaged:  2, // file1.txt and file5.txt
			staged:    2, // file2.txt and file5.txt
			untracked: 1, // file3.txt
			conflicts: 1, // file4.txt
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			unstaged, staged, untracked, conflicts := parsePorcelainStatus(tc.output)
			if unstaged != tc.unstaged {
				t.Errorf("expected unstaged %d, got %d", tc.unstaged, unstaged)
			}
			if staged != tc.staged {
				t.Errorf("expected staged %d, got %d", tc.staged, staged)
			}
			if untracked != tc.untracked {
				t.Errorf("expected untracked %d, got %d", tc.untracked, untracked)
			}
			if conflicts != tc.conflicts {
				t.Errorf("expected conflicts %d, got %d", tc.conflicts, conflicts)
			}
		})
	}
}

func TestGetDetailedStatus(t *testing.T) {
	ctx := context.Background()
	repoPath := setupTestRepo(t, "main")
	defer os.RemoveAll(repoPath)

	// 1. Check clean status
	info, err := GetDetailedStatus(ctx, repoPath)
	if err != nil {
		t.Fatalf("failed to get detailed status: %v", err)
	}
	if info.Branch != "main" {
		t.Errorf("expected branch 'main', got %q", info.Branch)
	}
	if info.UnstagedCount != 0 || info.StagedCount != 0 || info.UntrackedCount != 0 || info.ConflictCount != 0 {
		t.Errorf("expected clean working tree, got %+v", info)
	}

	// 2. Add an untracked file
	untrackedFile := filepath.Join(repoPath, "untracked.txt")
	if err := os.WriteFile(untrackedFile, []byte("untracked"), 0o644); err != nil {
		t.Fatalf("failed to write untracked file: %v", err)
	}

	info, err = GetDetailedStatus(ctx, repoPath)
	if err != nil {
		t.Fatalf("failed to get detailed status: %v", err)
	}
	if info.UntrackedCount != 1 {
		t.Errorf("expected 1 untracked file, got %d", info.UntrackedCount)
	}
	if info.UnstagedCount != 0 || info.StagedCount != 0 {
		t.Errorf("expected 0 unstaged/staged changes, got %+v", info)
	}

	// 3. Make unstaged change to tracked file
	testFile := filepath.Join(repoPath, "test.txt")
	if err := os.WriteFile(testFile, []byte("modified content"), 0o644); err != nil {
		t.Fatalf("failed to modify test file: %v", err)
	}

	info, err = GetDetailedStatus(ctx, repoPath)
	if err != nil {
		t.Fatalf("failed to get detailed status: %v", err)
	}
	if info.UnstagedCount != 1 {
		t.Errorf("expected 1 unstaged change, got %d", info.UnstagedCount)
	}

	// Cleanup untracked file so other tests running in parallel won't be affected (though this has its own temp dir)
}
