package dotfiles

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/dotfiles"
)

var InstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install dotfiles from a bare git repository",
	Long: `Install dotfiles from a bare git repository. This command will:
	- Check and install prerequisites (Homebrew, Git, Bash)
	- Setup SSH keys for GitHub when required by the repository URL
  - Clone your dotfiles repository as a bare repository
  - Backup any conflicting files
  - Checkout dotfiles to your home directory
  - Initialize git submodules
  - Configure git to hide untracked files`,
	RunE: func(cmd *cobra.Command, args []string) error {
		repoURL, branch, bareRepoPath, worktreePath, err := config.VerifyDotfilesConfig()
		if err != nil {
			return fmt.Errorf("configuration verification failed: %w", err)
		}

		opts := dotfiles.InstallOptions{
			RepoURL:      repoURL,
			Branch:       branch,
			BareRepoPath: bareRepoPath,
			WorktreePath: worktreePath,
			Verbose:      cmdutil.IsVerbose(cmd),
		}

		if err := dotfiles.Install(cmd.Context(), opts); err != nil {
			return fmt.Errorf("dotfiles installation failed: %w", err)
		}
		return nil
	},
}
