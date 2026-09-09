package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/log"
)

var SetupDotfilesCmd = &cobra.Command{
	Use:   "dotfiles",
	Short: "Setup dotfiles from your git repository",
	Long: `Setup dotfiles from your git repository. This command will:
	- Check and install prerequisites (Homebrew, Git, Bash)
	- Setup SSH keys for GitHub when required by the repository URL
  - Clone your dotfiles repository as a bare repository
  - Backup any conflicting files
  - Checkout dotfiles to your home directory
  - Initialize git submodules
	- Configure git to hide untracked files
	- Restore dotfiles secrets when manifest and BWS token are available`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := setupDotfiles(cmdutil.IsVerbose(cmd)); err != nil {
			return fmt.Errorf("dotfiles setup failed: %w", err)
		}
		return nil
	},
}

// setupDotfiles sets up dotfiles by checking prerequisites and running the install command.
func setupDotfiles(verbose bool) error {
	log.Verbose(verbose, "Starting dotfiles setup...")

	// Get the path to the current executable
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	log.Start("Running dotfiles install...")
	// Run dependencies install command
	args := []string{"dotfiles", "install"}
	if verbose {
		args = append(args, "-v")
	}
	cmd := execCommand(exe, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("dotfiles install failed: %w", err)
	}

	if err := maybeRestoreDotfilesSecrets(exe, verbose); err != nil {
		return err
	}

	return nil
}

func maybeRestoreDotfilesSecrets(exe string, verbose bool) error {
	manifestPath := filepath.Join(resolveDotfilesWorktreePath(), "bin", "secrets", "server.manifest")
	if _, err := stat(manifestPath); err != nil {
		log.Verbose(verbose, "Skipping dotfiles secrets restore, manifest not found: %s", manifestPath)
		return nil
	}

	if strings.TrimSpace(os.Getenv("BWS_ACCESS_TOKEN")) == "" {
		log.Warn("Skipping dotfiles secrets restore: BWS_ACCESS_TOKEN is not set")
		log.Message("Run manually after exporting BWS_ACCESS_TOKEN: eng dotfiles secrets restore")
		return nil
	}

	log.Start("Restoring dotfiles secrets...")
	args := []string{"dotfiles", "secrets", "restore", "--manifest", manifestPath}
	if verbose {
		args = append(args, "-v")
	}

	cmd := execCommand(exe, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("dotfiles secrets restore failed: %w", err)
	}

	return nil
}

func resolveDotfilesWorktreePath() string {
	worktreePath := strings.TrimSpace(viper.GetString("dotfiles.worktree_path"))
	if worktreePath == "" {
		worktreePath = strings.TrimSpace(os.Getenv("HOME"))
	}
	if worktreePath == "" {
		return "."
	}
	return os.ExpandEnv(worktreePath)
}
