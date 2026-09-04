package system

import (
	"github.com/eng618/eng/internal/dotfiles"
)

// Wire the workstation-setup hooks consumed by internal/dotfiles.
//
// internal/ packages must never import cmd/ packages (commands compose
// services, never the reverse), so the dotfiles workflow declares these
// steps as hook variables with unwired defaults. This init assigns the real
// implementations from this package, which owns interactive workstation
// setup. It runs for every eng invocation since cmd/system is always linked
// via the root command.
func init() {
	dotfiles.EnsurePrerequisites = EnsurePrerequisites
	dotfiles.FindGitHubSSHKey = FindGitHubSSHKey
	dotfiles.SetupSSHForGitHub = SetupSSHForGitHub
}
