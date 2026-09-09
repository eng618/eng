package ssh

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
)

// SetupSSH handles SSH key setup for GitHub and general access with multiple sources.
func SetupSSH(verbose bool) error {
	log.Start("Setting up SSH keys for GitHub access")

	homeDir, err := userHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}

	sshDir := filepath.Join(homeDir, ".ssh")
	sshKeyPath := FindGitHubSSHKey(sshDir)

	// Check if a valid SSH key already exists and authenticates
	if _, err := stat(sshKeyPath); err == nil {
		log.Verbose(verbose, "Found existing SSH key at %s", sshKeyPath)
		_ = EnsureSSHConfig(sshKeyPath)
		if err := ValidateGitHubSSHAuth(sshKeyPath, verbose); err == nil {
			log.Success("SSH key at %s is configured and validated for GitHub", sshKeyPath)
			return nil
		}
		log.Warn("Existing SSH key at %s failed GitHub authentication", sshKeyPath)
	}

	// Interactive source selection
	options := []string{
		"backup - Import / migrate from SSH backup folder (Box / Downloads)",
		"bitwarden - Retrieve SSH key from Bitwarden vault",
		"generate - Generate a new SSH key pair (ed25519)",
		"skip - Skip SSH setup (configure manually later)",
	}

	selected, err := ui.Select("How would you like to set up your SSH keys for GitHub?", options, options[0])
	if err != nil {
		// Non-interactive / headless fallback
		log.Verbose(verbose, "Selection non-interactive fallback (%v); checking candidate backup folders...", err)
		candidates := DetectSSHBackupCandidates()
		if len(candidates) > 0 {
			selected = "backup"
		} else {
			selected = "generate"
		}
	}

	choice := strings.Split(selected, " ")[0]
	defaultKeyTarget := filepath.Join(sshDir, "github")

	switch choice {
	case "backup":
		if err := setupSSHFromBackupFlow(sshDir, verbose); err != nil {
			log.Warn("SSH backup import failed: %v", err)
			log.Message("Falling back to key generation...")
			return setupSSHGenerateFlow(defaultKeyTarget, verbose)
		}
	case "bitwarden":
		if err := SetupSSHFromBitwarden(defaultKeyTarget, verbose); err != nil {
			log.Warn("Could not retrieve SSH key from Bitwarden: %v", err)
			log.Message("Falling back to key generation...")
			return setupSSHGenerateFlow(defaultKeyTarget, verbose)
		}
	case "generate":
		return setupSSHGenerateFlow(defaultKeyTarget, verbose)
	case "skip":
		log.Message("SSH setup skipped.")
		log.Message("You can configure SSH keys later with: eng ssh setup")
		return nil
	default:
		return setupSSHGenerateFlow(defaultKeyTarget, verbose)
	}

	// Resolve active key after migration/retrieval
	activeKey := FindGitHubSSHKey(sshDir)
	_ = EnsureSSHConfig(activeKey)
	return ValidateGitHubSSHAuth(activeKey, verbose)
}

// setupSSHGenerateFlow generates a new SSH key pair, attempts GitHub registration, and validates auth.
func setupSSHGenerateFlow(sshKeyPath string, verbose bool) error {
	autoRegistered, err := GenerateSSHKey(sshKeyPath, verbose)
	if err != nil {
		return fmt.Errorf("failed to generate SSH key: %w", err)
	}

	if err := EnsureSSHConfig(sshKeyPath); err != nil {
		return err
	}

	if !autoRegistered {
		if err := waitForManualGitHubKeyRegistration(sshKeyPath); err != nil {
			return err
		}
	}

	return ValidateGitHubSSHAuth(sshKeyPath, verbose)
}
