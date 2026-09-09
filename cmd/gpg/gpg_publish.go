package gpg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
)

// setGPGTrust sets a GPG key to ultimate trust level non-interactively.
func setGPGTrust(keyID string, verbose bool) error {
	log.Verbose(verbose, "Setting key trust level to ultimate for %s...", keyID)

	primaryFpr, err := resolvePrimaryFingerprint(keyID, verbose)
	if err != nil || primaryFpr == "" {
		primaryFpr = keyID
	}

	trustInput := fmt.Sprintf("%s:6:\n", strings.ToUpper(primaryFpr))
	cmd := execCommand("gpg", "--import-ownertrust")
	cmd.Stdin = strings.NewReader(trustInput)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()
	if err := cmd.Run(); err != nil {
		log.Warn("Could not set ownertrust automatically via --import-ownertrust: %v", err)
		log.Message("You can manually set trust by running: gpg --edit-key %s", keyID)
		return nil // Non-fatal - user can set manually
	}

	log.Success("Key trust set to ultimate")
	return nil
}

// configureGitSigning configures Git to use the GPG key for signing commits.
func configureGitSigning(keyID string, verbose bool) error {
	log.Message("")
	log.Start("Configuring Git signing...")

	// Set signing key
	cmd := execCommand("git", "config", "--global", "user.signingkey", keyID)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set signing key: %w", err)
	}

	// Enable auto-signing
	cmd = execCommand("git", "config", "--global", "commit.gpgsign", "true")
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to enable auto-signing: %w", err)
	}

	log.Success("Git configured to sign commits with key: %s", keyID)
	return nil
}

// refreshGPGPublicKey updates the public key from the keyserver to get the latest version.
func refreshGPGPublicKey(keyID string, verbose bool) error {
	log.Message("")
	log.Start("Refreshing Public Key from Keyserver")
	log.Message("Checking for latest version of your public key on the keyserver...")

	targetKey := keyID
	if fpr, err := resolvePrimaryFingerprint(keyID, verbose); err == nil && fpr != "" {
		targetKey = fpr
	}

	// Refresh the key from the keyserver
	cmd := execCommand("gpg", "--keyserver", "hkps://keys.openpgp.org", "--recv-keys", targetKey)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()
	if err := cmd.Run(); err != nil {
		log.Warn("Failed to refresh key from keyserver: %v", err)
		log.Message("Your local key may be outdated. You can manually refresh with:")
		log.Message("  gpg --keyserver hkps://keys.openpgp.org --recv-keys %s", targetKey)
		return nil // Non-fatal
	}

	log.Success("Public key refreshed from keyserver")
	log.Verbose(verbose, "Key is now up-to-date with the latest version from the keyserver")
	return nil
}

// uploadPublicKeyOption prompts the user to upload their public key to a keyserver.
func uploadPublicKeyOption(keyID string, verbose bool) error {
	log.Message("")
	log.Start("Public Key Distribution")
	log.Message("Your public key has been refreshed and is ready to share.")
	log.Message("Upload it to keyservers so others can verify your signatures.")
	log.Message("")

	uploadKey, err := ui.Confirm("Upload public key to keyserver?", true)
	if err != nil {
		return nil // Non-fatal if user cancels
	}

	targetKey := keyID
	if fpr, err := resolvePrimaryFingerprint(keyID, verbose); err == nil && fpr != "" {
		targetKey = fpr
	}

	if !uploadKey {
		log.Message("You can upload your public key manually later:")
		log.Message("  gpg --keyserver hkps://keys.openpgp.org --send-keys %s", targetKey)
		return nil
	}

	log.Start("Uploading public key to keyserver...")

	// Export public key
	cmd := execCommand("gpg", "--armor", "--export", targetKey)
	publicKeyBytes, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to export public key: %w", err)
	}

	// Upload to keyserver
	cmd = execCommand("gpg", "--keyserver", "hkps://keys.openpgp.org", "--send-keys", targetKey)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()
	if err := cmd.Run(); err != nil {
		log.Warn("Failed to upload to OpenPGP keyserver: %v", err)
		log.Message("You can try uploading manually or to a different keyserver.")
	} else {
		log.Success("Public key uploaded to keyserver")
	}

	// Optional: Upload to GitHub
	uploadGitHub, err := ui.Confirm("Also upload public key to GitHub?", true)
	if err == nil && uploadGitHub {
		// Check if gh is available
		if _, err := lookPath("gh"); err == nil {
			log.Start("Uploading public key to GitHub...")

			// Create temp file for public key
			homeDir, _ := userHomeDir()
			keyFile := filepath.Join(homeDir, ".gpg_temp.asc")
			if err := os.WriteFile(keyFile, publicKeyBytes, 0o600); err != nil {
				log.Warn("Failed to create temp key file: %v", err)
				return nil
			}
			defer func() {
				if err := os.Remove(keyFile); err != nil {
					log.Verbose(verbose, "Failed to remove temp key file: %v", err)
				}
			}()

			// Upload using gh CLI
			cmd := execCommand("gh", "gpg-key", "add", keyFile)
			cmd.Stdout = log.Writer()
			cmd.Stderr = log.ErrorWriter()
			if err := cmd.Run(); err != nil {
				log.Warn("Failed to upload to GitHub: %v", err)
				log.Message("You can upload manually via GitHub Settings → SSH and GPG keys")
			} else {
				log.Success("Public key uploaded to GitHub")
			}
		} else {
			log.Message("GitHub CLI (gh) not found. You can upload manually via GitHub Settings → SSH and GPG keys")
		}
	}

	return nil
}

// ensureGPGConfiguration ensures ~/.gnupg has 0700 permissions, config files have 0600 permissions,
// ~/.local/bin/pinentry wrapper is executable if present, and restarts gpg-agent to apply changes.
func ensureGPGConfiguration(verbose bool) {
	log.Message("")
	log.Start("Configuring GPG environment and permissions...")

	homeDir, err := userHomeDir()
	if err != nil {
		log.Error("Could not determine home directory: %v", err)
		return
	}

	gpgDir := filepath.Join(homeDir, ".gnupg")
	if err := os.MkdirAll(gpgDir, 0o700); err != nil {
		log.Warn("Failed to ensure .gnupg directory permissions: %v", err)
	} else {
		_ = os.Chmod(gpgDir, 0o700)
	}

	// Ensure config files inside ~/.gnupg have 0600 permissions
	for _, confFile := range []string{"gpg-agent.conf", "dirmngr.conf", "gpg.conf"} {
		confPath := filepath.Join(gpgDir, confFile)
		if _, err := stat(confPath); err == nil {
			_ = os.Chmod(confPath, 0o600)
			log.Verbose(verbose, "Set permissions 0600 on %s", confPath)
		}
	}

	// Ensure ~/.local/bin/pinentry is executable if present
	pinentryWrapper := filepath.Join(homeDir, ".local", "bin", "pinentry")
	if _, err := stat(pinentryWrapper); err == nil {
		_ = os.Chmod(pinentryWrapper, 0o755)
		log.Verbose(verbose, "Set permissions 0755 on %s", pinentryWrapper)
	}

	// Restart gpg-agent to pick up any configuration / pinentry changes
	log.Verbose(verbose, "Restarting gpg-agent to apply configuration...")
	killCmd := execCommand("gpgconf", "--kill", "gpg-agent")
	_ = killCmd.Run()

	launchCmd := execCommand("gpgconf", "--launch", "gpg-agent")
	if err := launchCmd.Run(); err != nil {
		log.Verbose(verbose, "gpgconf --launch gpg-agent output: %v", err)
	} else {
		log.Success("GPG agent restarted with updated configuration")
	}

	log.Success("GPG environment and directory permissions configured")
}
