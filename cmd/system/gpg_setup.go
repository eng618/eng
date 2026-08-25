package system

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := setupGPG(cmdutil.IsVerbose(cmd)); err != nil {
			return fmt.Errorf("gpg setup failed: %w", err)
		}
		return nil
	},
}

// GPGKeyInfo represents a GPG secret key found in the local keyring.
type GPGKeyInfo struct {
	KeyID       string   // 16-character Long Key ID (e.g. BE363376C8A71C92)
	Fingerprint string   // 40-character Primary Fingerprint
	UID         string   // e.g. "Eric N. Garcia <eng618@garciaericn.com>"
	HasMaster   bool     // true if master secret key is present (sec vs sec#)
	Subkeys     []string // Subkey IDs / fingerprints
}

// listLocalSecretGPGKeys lists all secret keys currently present in the GPG keyring.
func listLocalSecretGPGKeys(verbose bool) ([]GPGKeyInfo, error) {
	cmd := execCommand("gpg", "--list-secret-keys", "--with-colons", "--keyid-format", "LONG")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list secret keys: %w", err)
	}

	lines := strings.Split(string(out), "\n")
	var keys []GPGKeyInfo
	var currentKey *GPGKeyInfo
	var lastRecordType string

	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) == 0 {
			continue
		}
		recType := fields[0]
		switch recType {
		case "sec":
			if currentKey != nil {
				keys = append(keys, *currentKey)
			}
			keyID := ""
			if len(fields) > 4 {
				keyID = fields[4]
			}
			hasMaster := true
			// In gpg with-colons output, dummy secret keys or keys without master secret have '#' in validity/flags
			if len(fields) > 14 && strings.Contains(fields[14], "#") {
				hasMaster = false
			}
			currentKey = &GPGKeyInfo{
				KeyID:     keyID,
				HasMaster: hasMaster,
			}
			lastRecordType = "sec"
		case "ssb", "sub":
			lastRecordType = recType
			if currentKey != nil && len(fields) > 4 && fields[4] != "" {
				currentKey.Subkeys = append(currentKey.Subkeys, fields[4])
			}
		case "fpr":
			if len(fields) > 9 && fields[9] != "" {
				fpr := fields[9]
				if lastRecordType == "sec" && currentKey != nil {
					if currentKey.Fingerprint == "" {
						currentKey.Fingerprint = fpr
					}
				} else if (lastRecordType == "ssb" || lastRecordType == "sub") && currentKey != nil {
					currentKey.Subkeys = append(currentKey.Subkeys, fpr)
				}
			}
		case "uid":
			if currentKey != nil && currentKey.UID == "" && len(fields) > 9 && fields[9] != "" {
				currentKey.UID = fields[9]
			}
		}
	}

	if currentKey != nil {
		keys = append(keys, *currentKey)
	}

	return keys, nil
}

// setupGPG runs the interactive GPG setup flow.
func setupGPG(verbose bool) error {
	log.Verbose(verbose, "Starting GPG setup...")

	// Step 1: Ensure gnupg and pinentry are installed
	if err := ensureGPGDependencies(verbose); err != nil {
		return err
	}

	// Step 2: Prompt for key files and import them (or select from existing keys)
	keyID, keyInfo, err := importGPGKeys(verbose)
	if err != nil {
		return fmt.Errorf("failed to import/select GPG keys: %w", err)
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
	removedAnyMaster := false
	if err := promptAndRemoveMasterKeys(keyID, keyInfo, verbose, &removedAnyMaster); err != nil {
		log.Warn("Master key removal encountered an issue: %v", err)
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
	if removedAnyMaster {
		log.Message("  • Enhanced security (subkeys only, master key offline)")
	}

	return nil
}

// promptAndRemoveMasterKeys asks user which master key(s) to remove, supporting single and multi-selection.
func promptAndRemoveMasterKeys(activeKeyID string, activeKeyInfo GPGKeyInfo, verbose bool, removedAny *bool) error {
	availableKeys, _ := listLocalSecretGPGKeys(verbose)
	var keysWithMaster []GPGKeyInfo
	for _, k := range availableKeys {
		if k.HasMaster {
			keysWithMaster = append(keysWithMaster, k)
		}
	}

	if len(keysWithMaster) == 0 {
		log.Verbose(verbose, "No secret master keys detected in local keyring to remove (keys are already subkey-only).")
		return nil
	}

	var keysToRemove []GPGKeyInfo
	if len(keysWithMaster) == 1 {
		k := keysWithMaster[0]
		keyLabel := fmt.Sprintf("%s (Key ID: %s)", k.UID, k.KeyID)
		if k.UID == "" {
			keyLabel = fmt.Sprintf("Key ID: %s", k.KeyID)
		}
		removeKey, err := ui.Confirm(
			fmt.Sprintf("Remove master key for %s and keep only subkeys for enhanced security?", keyLabel),
			true,
		)
		if err != nil {
			log.Warn("Could not prompt for master key removal: %v", err)
			return nil
		}
		if removeKey {
			keysToRemove = append(keysToRemove, k)
		}
	} else {
		// Multiple keys with master secret found: provide multi-select
		options := make([]string, len(keysWithMaster))
		keyMap := make(map[string]GPGKeyInfo)
		for i, k := range keysWithMaster {
			opt := fmt.Sprintf("[%s] %s", k.KeyID, k.UID)
			if k.UID == "" {
				opt = fmt.Sprintf("Key ID: %s", k.KeyID)
			}
			options[i] = opt
			keyMap[opt] = k
		}
		selected, err := ui.MultiSelect(
			"Select GPG key(s) to remove master key for (keeping only subkeys):",
			options,
			[]string{options[0]},
		)
		if err != nil {
			log.Warn("Multi-select canceled: %v", err)
			return nil
		}
		for _, sel := range selected {
			if k, ok := keyMap[sel]; ok {
				keysToRemove = append(keysToRemove, k)
			}
		}
	}

	for _, k := range keysToRemove {
		targetKey := k.Fingerprint
		if targetKey == "" {
			targetKey = k.KeyID
		}
		label := k.KeyID
		if k.UID != "" {
			label = fmt.Sprintf("%s (%s)", k.UID, k.KeyID)
		}
		log.Start("Removing master key for %s...", label)
		if err := removeGPGMasterKey(targetKey, verbose); err != nil {
			log.Error("Failed to remove master key for %s: %v", label, err)
			log.Message("You can manually remove it later by running: gpg --delete-secret-keys %s", targetKey)
		} else {
			*removedAny = true
			log.Success("Master key removed for %s - only subkeys remain for local signing/encryption", label)
		}
	}

	return nil
}

// ensureGPGDependencies checks for gnupg and pinentry installations.
func ensureGPGDependencies(verbose bool) error {
	log.Verbose(verbose, "Checking for GPG dependencies...")

	distro := detectDistro()

	// Check for gnupg
	if _, err := lookPath("gpg"); err != nil {
		if distro.IsFedora() {
			return fmt.Errorf("gpg is not installed - please install it via: sudo dnf install -y gnupg2")
		} else if distro.IsDebianUbuntu() {
			return fmt.Errorf("gpg is not installed - please install it via: sudo apt-get install -y gnupg")
		}
		return fmt.Errorf("gpg is not installed - please install it via: brew install gnupg")
	}
	log.Verbose(verbose, "gnupg is installed")

	// Check for pinentry (pinentry-mac on macOS, pinentry / pinentry-curses / pinentry-gnome3 / pinentry-qt on Linux)
	foundPinentry := false
	pinentryCandidates := []string{"pinentry", "pinentry-curses", "pinentry-gnome3", "pinentry-qt", "pinentry-mac"}
	if distro.IsMacOS() {
		pinentryCandidates = []string{"pinentry-mac", "pinentry"}
	}
	for _, candidate := range pinentryCandidates {
		if _, err := lookPath(candidate); err == nil {
			foundPinentry = true
			log.Verbose(verbose, "found pinentry binary: %s", candidate)
			break
		}
	}

	if !foundPinentry {
		if distro.IsFedora() {
			return fmt.Errorf("pinentry is not installed - please install it via: sudo dnf install -y pinentry")
		} else if distro.IsDebianUbuntu() {
			return fmt.Errorf(
				"pinentry is not installed - please install it via: sudo apt-get install -y pinentry-curses",
			)
		}
		return fmt.Errorf("pinentry is not installed - please install it via: brew install pinentry-mac")
	}
	log.Verbose(verbose, "pinentry is installed")

	return nil
}

// importGPGKeys prompts user for key files, imports them if requested, and allows key selection.
func importGPGKeys(verbose bool) (string, GPGKeyInfo, error) {
	log.Message("")
	log.Start("GPG Key Import")

	// Check if keys already exist in keyring
	existingKeys, _ := listLocalSecretGPGKeys(verbose)
	shouldImportFiles := true

	if len(existingKeys) > 0 {
		log.Message("Found %d existing secret key(s) in local GPG keyring.", len(existingKeys))
		for _, k := range existingKeys {
			status := "sec (master key present)"
			if !k.HasMaster {
				status = "sec# (subkey-only)"
			}
			log.Message("  • [%s] %s - %s", k.KeyID, k.UID, status)
		}
		log.Message("")

		importMore, err := ui.Confirm("Do you want to import additional GPG key files?", false)
		if err == nil {
			shouldImportFiles = importMore
		}
	}

	if shouldImportFiles {
		log.Message("You can provide GPG key files to import:")
		log.Message("  • A master secret key file (e.g., eng618.secret.gpg)")
		log.Message("  • Subkeys file (e.g., eng618.secsub.gpg)")
		log.Message("")

		secretKeyPath, err := ui.Input(
			"Path to secret key file (leave empty to skip)",
			filepath.Join(os.Getenv("HOME"), "Downloads", "gpg", "eng618.secret.gpg"),
		)
		if err != nil {
			return "", GPGKeyInfo{}, fmt.Errorf("canceled: %w", err)
		}

		secretKeyPath = strings.TrimSpace(secretKeyPath)
		if secretKeyPath != "" {
			if _, err := os.Stat(secretKeyPath); err != nil {
				log.Warn("Secret key file not found at %s: %v", secretKeyPath, err)
			} else {
				log.Start("Importing secret key...")
				cmd := execCommand("gpg", "--import", secretKeyPath)
				cmd.Stdout = log.Writer()
				cmd.Stderr = log.ErrorWriter()
				if err := cmd.Run(); err != nil {
					return "", GPGKeyInfo{}, fmt.Errorf("failed to import secret key: %w", err)
				}
				log.Success("Secret key imported")
			}
		}

		// Optional: Import subkeys
		importSubkeys, err := ui.Confirm("Import subkeys file?", true)
		if err == nil && importSubkeys {
			subkeysPath, err := ui.Input(
				"Path to subkeys file",
				filepath.Join(os.Getenv("HOME"), "Downloads", "gpg", "eng618.secsub.gpg"),
			)
			if err == nil {
				subkeysPath = strings.TrimSpace(subkeysPath)
				if subkeysPath != "" {
					if _, err := os.Stat(subkeysPath); err != nil {
						log.Warn("Subkeys file not found at %s: %v", subkeysPath, err)
					} else {
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
		}
	}

	// Re-list secret keys from keyring
	keys, err := listLocalSecretGPGKeys(verbose)
	if err == nil && len(keys) > 0 {
		if len(keys) == 1 {
			k := keys[0]
			keyLabel := fmt.Sprintf("%s (Key ID: %s)", k.UID, k.KeyID)
			if k.UID == "" {
				keyLabel = fmt.Sprintf("Key ID: %s", k.KeyID)
			}
			log.Message("Detected GPG secret key: %s", keyLabel)
			useDetected, err := ui.Confirm(fmt.Sprintf("Configure detected GPG key %s?", keyLabel), true)
			if err == nil && useDetected {
				target := k.KeyID
				if target == "" {
					target = k.Fingerprint
				}
				log.Success("Selected key: %s", keyLabel)
				return target, k, nil
			}
		} else {
			// Multi-key selection
			options := make([]string, 0, len(keys)+1)
			keyMap := make(map[string]GPGKeyInfo)
			for _, k := range keys {
				opt := fmt.Sprintf("[%s] %s", k.KeyID, k.UID)
				if k.UID == "" {
					opt = fmt.Sprintf("Key ID: %s", k.KeyID)
				}
				options = append(options, opt)
				keyMap[opt] = k
			}
			options = append(options, "Enter key ID manually...")

			selected, err := ui.Select("Select GPG key to configure:", options, options[0])
			if err == nil && selected != "Enter key ID manually..." {
				chosen := keyMap[selected]
				target := chosen.KeyID
				if target == "" {
					target = chosen.Fingerprint
				}
				log.Success("Selected key: %s", selected)
				return target, chosen, nil
			}
		}
	}

	// Fallback: manual entry
	keyID, err := ui.Input("Enter your GPG key ID or Fingerprint (e.g., 7C180F0FCB31441B)", "")
	if err != nil {
		return "", GPGKeyInfo{}, fmt.Errorf("canceled: %w", err)
	}

	keyID = strings.TrimSpace(keyID)
	if keyID == "" {
		return "", GPGKeyInfo{}, fmt.Errorf("key ID is required")
	}

	// Verify the key exists
	listCmd := execCommand("gpg", "--list-secret-keys", "--keyid-format", "LONG", keyID)
	if err := listCmd.Run(); err != nil {
		return "", GPGKeyInfo{}, fmt.Errorf("key not found in keyring: %s", keyID)
	}

	log.Success("Key verified: %s", keyID)
	return keyID, GPGKeyInfo{KeyID: keyID}, nil
}

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

// removeGPGMasterKey exports subkeys, removes the entire key, and re-imports subkeys only.
// This implements the subkey-only workflow for enhanced security.
func removeGPGMasterKey(keyID string, verbose bool) error {
	homeDir, err := userHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}

	gpgDir := filepath.Join(homeDir, ".gnupg")
	subkeysExportPath := filepath.Join(gpgDir, "subkeys-only.gpg")

	primaryFpr, err := resolvePrimaryFingerprint(keyID, verbose)
	if err != nil {
		return fmt.Errorf("failed to resolve key fingerprint: %w", err)
	}

	// Step 1: Export subkeys
	log.Verbose(verbose, "Exporting subkeys for key %s...", primaryFpr)
	cmd := execCommand("gpg", "--export-secret-subkeys", primaryFpr)
	subkeysOutput, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to export subkeys: %w", err)
	}

	// Write subkeys to file
	if err := os.WriteFile(subkeysExportPath, subkeysOutput, 0o600); err != nil {
		return fmt.Errorf("failed to write subkeys file: %w", err)
	}
	log.Verbose(verbose, "Subkeys exported to: "+subkeysExportPath)

	// Step 2: Delete the entire secret key using full fingerprint (required by GPG in batch mode)
	log.Verbose(verbose, "Removing master key from local keyring using fingerprint %s...", primaryFpr)
	cmd = execCommand("gpg", "--batch", "--yes", "--delete-secret-keys", primaryFpr)
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

	log.Message("Subkeys backup saved to: %s", subkeysExportPath)
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
