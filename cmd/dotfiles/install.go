package dotfiles

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/cmdutil"
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
		if err := dotfiles.Install(cmd.Context(), cmdutil.IsVerbose(cmd)); err != nil {
			return fmt.Errorf("dotfiles installation failed: %w", err)
		}
		return nil
	},
}
