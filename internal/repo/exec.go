package repo

import (
	"context"
	"os"
	"os/exec"
	"time"
)

// gitTimeout is the maximum duration a git network operation can take.
const gitTimeout = 2 * time.Minute

// execGitCommand prepares an exec.Cmd for a git operation with timeouts and non-interactive environment variables.
func execGitCommand(ctx context.Context, dir string, args ...string) (*exec.Cmd, context.CancelFunc) {
	cmdCtx, cancel := context.WithTimeout(ctx, gitTimeout)

	cmd := exec.CommandContext(cmdCtx, "git", args...)
	if dir != "" {
		cmd.Dir = dir
	}

	// Make sure git never prompts for anything on stdin
	env := os.Environ()
	env = append(env, "GIT_TERMINAL_PROMPT=0")
	env = append(env, "GIT_SSH_COMMAND=ssh -o BatchMode=yes")
	cmd.Env = env

	return cmd, cancel
}
