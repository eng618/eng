package ssh

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/cmdutil"
)

// SshCmd manages SSH keys for GitHub and general access.
var SshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "Manage SSH keys for GitHub access",
	Long:  `Check for existing SSH keys, retrieve from Bitwarden, generate new ed25519 keys, and configure SSH for GitHub.`,
	RunE: func(cmd *cobra.Command, _args []string) error {
		return cmd.Help()
	},
}

// SetupCmd runs the SSH setup flow.
var SetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup SSH keys for GitHub access",
	Long: `Setup SSH keys for GitHub access. This command will:
  - Check for existing SSH keys
  - Attempt to retrieve SSH keys from Bitwarden vault
  - Generate new SSH keys if none found
  - Configure SSH config for GitHub`,
	RunE: func(cmd *cobra.Command, _args []string) error {
		if err := SetupSSH(cmdutil.IsVerbose(cmd)); err != nil {
			return fmt.Errorf("ssh setup failed: %w", err)
		}
		return nil
	},
}

// SetupSSHForGitHub exposes the SSH setup flow for other commands that need GitHub access.
func SetupSSHForGitHub(verbose bool) error {
	return SetupSSH(verbose)
}

func init() {
	SshCmd.AddCommand(SetupCmd)
}
