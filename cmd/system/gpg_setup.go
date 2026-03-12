package system

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/utils"
	"github.com/eng618/eng/internal/utils/log"
)

var SetupGPGCmd = &cobra.Command{
	Use:   "gpg",
	Short: "Setup GPG keys for signing and encryption",
	Long: `Setup GPG keys for signing commits and encryption. This command will:
  - Prompt you for GPG key files to import
  - Import master key and subkeys
  - Set ultimate trust on the key
  - Configure Git to use your GPG key for signing
  - Optionally remove the master key (keeping only subkeys for security)`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := setupGPG(utils.IsVerbose(cmd)); err != nil {
			log.Fatal("GPG setup failed: %v", err)
		}
	},
}

// setupGPG runs the interactive GPG setup flow.
func setupGPG(verbose bool) error {
	log.Verbose(verbose, "Starting GPG setup...")

	// Step 1: Ensure gnupg and pinentry are installed
	if err := ensureGPGDependencies(verbose); err != nil {
		return err
	}

	// Step 2: Prompt for key files and import them
	keyID, err := importGPGKeys(verbose)
	if err != nil {
		return fmt.Errorf("failed to import GPG keys: %w", err)
	}

	// Step 3: Set trust to ultimate
	if err := setGPGTrust(keyID, verbose); err != nil {
		return fmt.Errorf("failed to set trust level: %w", err)
	}

	// Step 4: Configure Git signing
	if err := configureGitSigning(keyID, verbose); err != nil {
		return fmt.Errorf("failed to configure git signing: %w", err)
	}

	// Step 5: Optional - Remove master key (subkey-only workflow)
	removeKey := false
	prompt := &survey.Confirm{
		Message: "Remove master key and keep only subkeys for enhanced security?",
		Default: true,
	}
	if err := askOne(prompt, &removeKey); err != nil {
		log.Warn("Could not prompt for master key removal: %v", err)
	}

	if removeKey {
		if err := removeGPGMasterKey(keyID, verbose); err != nil {
			log.Error("Failed to remove master key: %v", err)
			log.Message("You can manually remove it later by running: gpg --delete-secret-keys <keyid>")
		} else {
			log.Success("Master key removed - only subkeys remain for local signing and encryption")
		}
	}

	// Step 6: Refresh public key from keyserver
	if err := refreshGPGPublicKey(keyID, verbose); err != nil {
		log.Error("Failed to refresh public key: %v", err)
	}

	// Step 7: Optional - Upload public key to keyserver
	if err := uploadPublicKeyOption(keyID, verbose); err != nil {
		log.Error("Failed to upload public key: %v", err)
	}

	log.Success("GPG setup completed successfully!")
	log.Message("")
	log.Message("Your GPG key is now configured for:")
	log.Message("  • Signing commits")
	log.Message("  • Encrypting files and messages")
	if removeKey {
		log.Message("  • Enhanced security (subkeys only, master key offline)")
	}

	return nil
}

// ensureGPGDependencies checks for gnupg and pinentry installations.
func ensureGPGDependencies(verbose bool) error {
	log.Verbose(verbose, "Checking for GPG dependencies...")

	// Check for gnupg
	if _, err := lookPath("gpg"); err != nil {
		return fmt.Errorf("gpg is not installed - please install it via: brew install gnupg")
	}
	log.Verbose(verbose, "gnupg is installed")

	// Check for pinentry
	if _, err := lookPath("pinentry-mac"); err != nil {
		if _, err := lookPath("pinentry"); err != nil {
			return fmt.Errorf("pinentry is not installed - please install it via: brew install pinentry-mac")
		}
	}
	log.Verbose(verbose, "pinentry is installed")

	return nil
}

// importGPGKeys prompts user for key files and imports them, returning the key ID.
func importGPGKeys(verbose bool) (string, error) {
	log.Message("")
	log.Start("GPG Key Import")
	log.Message("You need to provide GPG key files to import.")
	log.Message("Typically you have:")
	log.Message("  • A master secret key file (e.g., eng618.secret.gpg)")
	log.Message("  • Subkeys file (e.g., eng618.secsub.gpg)")
	log.Message("")

	var secretKeyPath string
	secretPrompt := &survey.Input{
		Message: "Path to secret key file",
		Default: filepath.Join(os.Getenv("HOME"), "Downloads", "gpg", "eng618.secret.gpg"),
	}
	if err := askOne(secretPrompt, &secretKeyPath); err != nil {
		return "", fmt.Errorf("canceled: %w", err)
	}

	// Verify file exists
	if _, err := os.Stat(secretKeyPath); err != nil {
		return "", fmt.Errorf("file not found: %s", secretKeyPath)
	}

	// Import secret key
	log.Start("Importing secret key...")
	cmd := execCommand("gpg", "--import", secretKeyPath)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to import secret key: %w", err)
	}
	log.Success("Secret key imported")

	// Optional: Import subkeys
	importSubkeys := false
	subkeysPrompt := &survey.Confirm{
		Message: "Import subkeys file?",
		Default: true,
	}
	if err := askOne(subkeysPrompt, &importSubkeys); err == nil && importSubkeys {
		var subkeysPath string
		subkeysPathPrompt := &survey.Input{
			Message: "Path to subkeys file",
			Default: filepath.Join(os.Getenv("HOME"), "Downloads", "gpg", "eng618.secsub.gpg"),
		}
		if err := askOne(subkeysPathPrompt, &subkeysPath); err == nil {
			if _, err := os.Stat(subkeysPath); err == nil {
				log.Start("Importing subkeys...")
				cmd := execCommand("gpg", "--import", subkeysPath)
				cmd.Stdout = log.Writer()
				cmd.Stderr = log.ErrorWriter()
				if err := cmd.Run(); err != nil {
					log.Warn("Failed to import subkeys: %v", err)
				} else {
					log.Success("Subkeys imported")
				}
			}
		}
	}

	// Get key ID from user or list keys
	var keyID string
	keyIDPrompt := &survey.Input{
		Message: "Enter your GPG key ID (long format, e.g., 7C180F0FCB31441B)",
	}
	if err := askOne(keyIDPrompt, &keyID); err != nil {
		return "", fmt.Errorf("canceled: %w", err)
	}

	keyID = strings.TrimSpace(keyID)
	if keyID == "" {
		return "", fmt.Errorf("key ID is required")
	}

	// Validate key ID format to prevent argument injection
	validKeyID := regexp.MustCompile(`^[0-9A-Fa-f]{16}$`)
	if !validKeyID.MatchString(keyID) {
		return "", fmt.Errorf("invalid GPG key ID format: must be a 16-character hexadecimal string")
	}

	// Verify the key exists
	listCmd := execCommand("gpg", "--list-secret-keys", "--keyid-format", "LONG", keyID)
	if err := listCmd.Run(); err != nil {
		return "", fmt.Errorf("key not found: %s", keyID)
	}

	log.Success("Key ID verified: %s", keyID)
	return keyID, nil
}

// setGPGTrust sets a GPG key to ultimate trust level.
func setGPGTrust(keyID string, verbose bool) error {
	log.Message("")
	log.Start("Setting key trust level...")
	log.Message("You will be prompted to set the trust level for this key to 'ultimate'.")
	log.Message("At the 'gpg>' prompt, type: trust")
	log.Message("Then select 5 for 'ultimate' trust")
	log.Message("Then type 'save' to confirm")
	log.Message("")

	cmd := execCommand("gpg", "--edit-key", keyID)
	cmd.Stdin = strings.NewReader("trust\n5\ny\nsave\n")
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()
	if err := cmd.Run(); err != nil {
		log.Warn("Could not automatically set trust - please run: gpg --edit-key %s", keyID)
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

// removeGPGMasterKey exports subkeys, removes the entire key, and re-imports subkeys only.
// This implements the subkey-only workflow for enhanced security.
func removeGPGMasterKey(keyID string, verbose bool) error {
	log.Message("")
	log.Start("Removing master key (keeping subkeys only)...")

	homeDir, err := userHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}

	gpgDir := filepath.Join(homeDir, ".gnupg")
	subkeysExportPath := filepath.Join(gpgDir, "subkeys-only.gpg")

	// Step 1: Export subkeys
	log.Verbose(verbose, "Exporting subkeys...")
	cmd := execCommand("gpg", "--export-secret-subkeys", keyID)
	subkeysOutput, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to export subkeys: %w", err)
	}

	// Write subkeys to file
	if err := os.WriteFile(subkeysExportPath, subkeysOutput, 0o600); err != nil {
		return fmt.Errorf("failed to write subkeys file: %w", err)
	}
	log.Verbose(verbose, "Subkeys exported to: "+subkeysExportPath)

	// Step 2: Delete the entire secret key
	log.Verbose(verbose, "Removing master key from local keyring...")
	cmd = execCommand("gpg", "--batch", "--yes", "--delete-secret-keys", keyID)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete master key: %w", err)
	}

	// Step 3: Re-import subkeys only
	log.Verbose(verbose, "Re-importing subkeys only...")
	cmd = execCommand("gpg", "--import", subkeysExportPath)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to re-import subkeys: %w", err)
	}

	log.Success("Master key removed - subkeys available for signing/encryption")
	log.Message("Subkeys backup saved to: %s", subkeysExportPath)
	return nil
}

// refreshGPGPublicKey updates the public key from the keyserver to get the latest version.
func refreshGPGPublicKey(keyID string, verbose bool) error {
	log.Message("")
	log.Start("Refreshing Public Key from Keyserver")
	log.Message("Checking for latest version of your public key on the keyserver...")

	// Refresh the key from the keyserver
	cmd := execCommand("gpg", "--keyserver", "hkps://keys.openpgp.org", "--recv-keys", keyID)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()
	if err := cmd.Run(); err != nil {
		log.Warn("Failed to refresh key from keyserver: %v", err)
		log.Message("Your local key may be outdated. You can manually refresh with:")
		log.Message("  gpg --keyserver hkps://keys.openpgp.org --recv-keys %s", keyID)
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

	uploadKey := false
	prompt := &survey.Confirm{
		Message: "Upload public key to keyserver?",
		Default: true,
	}
	if err := askOne(prompt, &uploadKey); err != nil {
		return nil // Non-fatal if user cancels
	}

	if !uploadKey {
		log.Message("You can upload your public key manually later:")
		log.Message("  gpg --keyserver hkps://keys.openpgp.org --send-keys %s", keyID)
		return nil
	}

	log.Start("Uploading public key to keyserver...")

	// Export public key
	cmd := execCommand("gpg", "--armor", "--export", keyID)
	publicKeyBytes, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to export public key: %w", err)
	}

	// Upload to keyserver
	cmd = execCommand("gpg", "--keyserver", "hkps://keys.openpgp.org", "--send-keys", keyID)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()
	if err := cmd.Run(); err != nil {
		log.Warn("Failed to upload to OpenPGP keyserver: %v", err)
		log.Message("You can try uploading manually or to a different keyserver.")
	} else {
		log.Success("Public key uploaded to keyserver")
	}

	// Optional: Upload to GitHub
	uploadGitHub := false
	githubPrompt := &survey.Confirm{
		Message: "Also upload public key to GitHub?",
		Default: true,
	}
	if err := askOne(githubPrompt, &uploadGitHub); err == nil && uploadGitHub {
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
