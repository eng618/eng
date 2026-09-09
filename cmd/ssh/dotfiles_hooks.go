package ssh

import (
	"github.com/eng618/eng/internal/dotfiles"
)

// Wire the SSH-related workstation-setup hooks consumed by internal/dotfiles.
//
// internal/ packages must never import cmd/ packages, so the dotfiles
// workflow declares these steps as hook variables. This init assigns the
// SSH-owned implementations. Prerequisites wiring lives in cmd/setup.
func init() {
	dotfiles.FindGitHubSSHKey = FindGitHubSSHKey
	dotfiles.SetupSSHForGitHub = SetupSSHForGitHub
}
