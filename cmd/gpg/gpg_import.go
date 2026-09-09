package gpg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
)

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
		log.Verbose(
			verbose,
			"No secret master keys detected in local keyring to remove (keys are already subkey-only).",
		)
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
