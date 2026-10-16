package gpg

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/log"
)

var SetupGPGCmd = &cobra.Command{
	Use:     "setup",
	Aliases: []string{"init"},
	Short:   "Setup GPG keys for signing and encryption",
	Long: `Setup GPG keys for signing commits and encryption. This command will:
  - Prompt you for GPG key files to import
  - Import master key and subkeys
  - Set ultimate trust on the key
  - Configure Git to use your GPG key for signing
  - Optionally remove the master key (keeping only subkeys for security)
  - Configure ~/.gnupg and wrapper permissions and restart gpg-agent`,
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

	// Step 8: Ensure GPG directory permissions and agent configuration
	ensureGPGConfiguration(verbose)

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
