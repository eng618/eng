package setup

import (
	"github.com/eng618/eng/internal/dotfiles"
)

// Wire the workstation-setup hooks consumed by internal/dotfiles.
//
// internal/ packages must never import cmd/ packages (commands compose
// services, never the reverse), so the dotfiles workflow declares these
// steps as hook variables with unwired defaults. This init assigns the
// prerequisites implementation owned by this package. SSH-related hooks
// (FindGitHubSSHKey, SetupSSHForGitHub) are wired by cmd/ssh, which owns
// those flows. These inits run for every eng invocation since both packages
// are always linked via the root command.
func init() {
	dotfiles.EnsurePrerequisites = EnsurePrerequisites
}
