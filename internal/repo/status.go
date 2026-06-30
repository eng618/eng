package repo

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// StatusInfo holds detailed git repository status information.
type StatusInfo struct {
	Branch         string
	IsDetached     bool
	AheadCount     int
	BehindCount    int
	HasUpstream    bool
	UnstagedCount  int
	StagedCount    int
	UntrackedCount int
	ConflictCount  int
	OngoingOp      string // "rebase", "merge", "cherry-pick", "bisect", or ""
}

// GetDetailedStatus retrieves rich git status details by running the standard Git CLI.
func GetDetailedStatus(ctx context.Context, repoPath string) (StatusInfo, error) {
	var info StatusInfo

	// 1. Get current branch and detached HEAD state
	cmd, cancel := execGitCommand(ctx, repoPath, "branch", "--show-current")
	out, err := cmd.Output()
	cancel()
	if err != nil {
		return info, fmt.Errorf("failed to get current branch: %w", err)
	}

	branch := strings.TrimSpace(string(out))
	if branch == "" {
		info.IsDetached = true
		// Get commit hash as label
		cmdHash, cancelHash := execGitCommand(ctx, repoPath, "rev-parse", "--short", "HEAD")
		hashOut, errHash := cmdHash.Output()
		cancelHash()
		if errHash == nil {
			info.Branch = fmt.Sprintf("(detached HEAD at %s)", strings.TrimSpace(string(hashOut)))
		} else {
			info.Branch = "(detached HEAD)"
		}
	} else {
		info.Branch = branch
	}

	// 2. Get git status --porcelain
	cmdStatus, cancelStatus := execGitCommand(ctx, repoPath, "status", "--porcelain")
	statusOut, errStatus := cmdStatus.Output()
	cancelStatus()
	if errStatus != nil {
		return info, fmt.Errorf("failed to get git status: %w", errStatus)
	}

	unstaged, staged, untracked, conflicts := parsePorcelainStatus(string(statusOut))
	info.UnstagedCount = unstaged
	info.StagedCount = staged
	info.UntrackedCount = untracked
	info.ConflictCount = conflicts

	// 3. Get ahead/behind counts if has upstream
	if !info.IsDetached {
		cmdAB, cancelAB := execGitCommand(ctx, repoPath, "rev-list", "--left-right", "--count", "HEAD...@{upstream}")
		abOut, errAB := cmdAB.Output()
		cancelAB()
		if errAB == nil {
			parts := strings.Fields(string(abOut))
			if len(parts) == 2 {
				info.AheadCount, _ = strconv.Atoi(parts[0])
				info.BehindCount, _ = strconv.Atoi(parts[1])
				info.HasUpstream = true
			}
		} else {
			info.HasUpstream = false
		}
	}

	// 4. Resolve git dir and check ongoing operations
	cmdGitDir, cancelGitDir := execGitCommand(ctx, repoPath, "rev-parse", "--git-dir")
	gitDirOut, errGitDir := cmdGitDir.Output()
	cancelGitDir()
	if errGitDir == nil {
		gitDir := strings.TrimSpace(string(gitDirOut))
		if !filepath.IsAbs(gitDir) {
			gitDir = filepath.Join(repoPath, gitDir)
		}

		if _, err = os.Stat(filepath.Join(gitDir, "rebase-merge")); err == nil {
			info.OngoingOp = "rebase"
		} else if _, err = os.Stat(filepath.Join(gitDir, "rebase-apply")); err == nil {
			info.OngoingOp = "rebase"
		} else if _, err = os.Stat(filepath.Join(gitDir, "MERGE_HEAD")); err == nil {
			info.OngoingOp = "merge"
		} else if _, err = os.Stat(filepath.Join(gitDir, "CHERRY_PICK_HEAD")); err == nil {
			info.OngoingOp = "cherry-pick"
		} else if _, err = os.Stat(filepath.Join(gitDir, "BISECT_LOG")); err == nil {
			info.OngoingOp = "bisect"
		}
	}

	return info, nil
}

func parsePorcelainStatus(output string) (unstaged, staged, untracked, conflicts int) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if len(line) < 3 {
			continue
		}
		x := line[0]
		y := line[1]

		// Check for conflicts (unmerged)
		isConflict := false
		switch {
		case x == 'U' || y == 'U':
			isConflict = true
		case x == 'A' && y == 'A':
			isConflict = true
		case x == 'D' && y == 'D':
			isConflict = true
		}

		if isConflict {
			conflicts++
			continue
		}

		if x == '?' && y == '?' {
			untracked++
			continue
		}

		if x != ' ' {
			staged++
		}
		if y != ' ' {
			unstaged++
		}
	}
	return unstaged, staged, untracked, conflicts
}
