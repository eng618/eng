package system

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
)

var (
	renewKeyDir     string
	renewDuration   string
	renewKeepMaster bool
)

var RenewGPGCmd = &cobra.Command{
	Use:     "renew",
	Aliases: []string{"update", "extend"},
	Short:   "Renew or extend GPG key and subkey expiration",
	Long: `Interactively inspect and extend expiration dates for your GPG key and subkeys.
Imports the master key from a local/cloud backup directory, updates expiration dates,
archives old key files to archive/<date-expired>/, re-exports updated secret and public
key files back to your backup folder with standard names (for cloud sync), publishes
public keys, and strips the master key locally (keeping subkeys only for security).`,
	RunE: func(cmd *cobra.Command, _args []string) error {
		return renewGPG(cmdutil.IsVerbose(cmd))
	},
}

func init() {
	// NOTE: the default stays empty on purpose. A home-derived default here
	// would bake the local username into --help output, making generated
	// docs differ per machine (and CI would flag it as drift). The runtime
	// default (~/Downloads/gpg) resolves in renewGPG instead.
	RenewGPGCmd.Flags().
		StringVarP(&renewKeyDir, "key-dir", "d", "", "Directory containing GPG key backups (default: ~/Downloads/gpg)")
	RenewGPGCmd.Flags().StringVar(&renewDuration, "duration", "1y", "Expiration duration (e.g., 1y, 2y, 6m)")
	RenewGPGCmd.Flags().
		BoolVar(&renewKeepMaster, "keep-master", false, "Keep master key in local keyring (do not strip to subkeys-only)")
}

// defaultRenewKeyDir resolves the runtime default key directory.
func defaultRenewKeyDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil || homeDir == "" {
		return ""
	}
	return filepath.Join(homeDir, "Downloads", "gpg")
}

// renewGPG guides the user through extending GPG key expiration and managing master key export/removal.
func renewGPG(verbose bool) error {
	log.Verbose(verbose, "Starting GPG key renewal process...")

	// 1. Ensure GPG dependencies exist
	if err := ensureGPGDependencies(verbose); err != nil {
		return err
	}

	// 2. Resolve Key Directory
	flagKeyDir := renewKeyDir
	if flagKeyDir == "" {
		flagKeyDir = defaultRenewKeyDir()
	}
	keyDir, err := ui.Input("Directory containing GPG key backups", flagKeyDir)
	if err != nil {
		return fmt.Errorf("canceled: %w", err)
	}
	keyDir = strings.TrimSpace(keyDir)
	if keyDir == "" {
		keyDir = flagKeyDir
	}

	if err := os.MkdirAll(keyDir, 0o700); err != nil {
		return fmt.Errorf("failed to access or create key directory %s: %w", keyDir, err)
	}

	// 3. Find and Import Master Secret Key
	secretKeyPath, err := findAndImportMasterKey(keyDir, verbose)
	if err != nil {
		log.Warn("Master key import step: %v", err)
		log.Message("Proceeding with existing keyring keys...")
	} else {
		log.Success("Master secret key imported from: %s", secretKeyPath)
	}

	// 4. Determine default Key ID from Git configuration
	defaultKeyID := getDefaultGitSigningKey()

	// 5. Prompt for Key ID
	promptMsg := "Enter GPG key ID (long format or fingerprint)"
	keyID, err := ui.Input(promptMsg, defaultKeyID)
	if err != nil {
		return fmt.Errorf("canceled: %w", err)
	}

	keyID = strings.TrimSpace(keyID)
	if keyID == "" {
		return fmt.Errorf("key ID is required")
	}

	// Validate hex key ID format (16 to 40 hex chars)
	validKeyID := regexp.MustCompile(`^[0-9A-Fa-f]{16,40}$`)
	if !validKeyID.MatchString(keyID) {
		return fmt.Errorf("invalid GPG key ID format: must be 16 to 40 hexadecimal characters")
	}

	// 6. Inspect GPG Key & Subkeys
	primaryFpr, subkeyFprs, err := inspectGPGKey(keyID, verbose)
	if err != nil {
		return fmt.Errorf("failed to inspect GPG key: %w", err)
	}

	// 7. Prompt for Expiration Duration
	duration, err := ui.Input("Enter expiration duration (e.g., 1y, 2y, 6m, 365d)", renewDuration)
	if err != nil {
		return fmt.Errorf("canceled: %w", err)
	}
	duration = strings.TrimSpace(duration)
	if duration == "" {
		duration = "1y"
	}

	// 8. Extend Primary Key Expiry
	log.Message("")
	log.Start("Updating primary key expiration date...")
	log.Message("Note: pinentry-mac may prompt for your GPG passphrase in a desktop dialog.")

	cmd := execCommand("gpg", "--quick-set-expire", primaryFpr, duration)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()
	if err := cmd.Run(); err != nil {
		log.Warn("Failed to set primary key expiry automatically: %v", err)
		log.Message("You can set it manually by running: gpg --edit-key %s", primaryFpr)
	} else {
		log.Success("Primary key expiry extended by %s", duration)
	}

	// 9. Extend Subkey Expiry (if any subkeys found)
	if len(subkeyFprs) > 0 {
		log.Message("")
		log.Start("Updating subkey expiration dates...")
		for _, subFpr := range subkeyFprs {
			log.Verbose(verbose, "Updating subkey: %s", subFpr)
			subCmd := execCommand("gpg", "--quick-set-expire", primaryFpr, duration, subFpr)
			subCmd.Stdout = log.Writer()
			subCmd.Stderr = log.ErrorWriter()
			if err := subCmd.Run(); err != nil {
				log.Warn("Failed to update subkey %s: %v", subFpr, err)
			} else {
				log.Success("Subkey %s expiry extended by %s", subFpr, duration)
			}
		}
	}

	// 10. Archive Existing Old Key Files Before Exporting New Ones
	log.Message("")
	log.Start("Archiving previous key files...")
	if archivePath, err := archiveExistingKeys(keyDir, verbose); err != nil {
		log.Warn("Failed to archive old keys: %v", err)
	} else {
		log.Verbose(verbose, "Previous keys archived to: %s", archivePath)
	}

	// 11. Export Updated Secret Keys & Public Keys back to key directory (for Box.com / cloud backup)
	log.Start("Exporting updated GPG key files to backup folder...")

	// Full master secret key export
	masterExportPath := filepath.Join(keyDir, "eng618.secret.gpg")
	cmd = execCommand("gpg", "--export-secret-keys", primaryFpr)
	masterBytes, err := cmd.Output()
	if err == nil && len(masterBytes) > 0 {
		if err := os.WriteFile(masterExportPath, masterBytes, 0o600); err == nil {
			log.Success("Master secret key exported to: %s", masterExportPath)
		}
	}

	// Subkeys-only secret export
	subkeysExportPath := filepath.Join(keyDir, "subkeys-only.gpg")
	cmd = execCommand("gpg", "--export-secret-subkeys", primaryFpr)
	subkeysBytes, err := cmd.Output()
	if err == nil && len(subkeysBytes) > 0 {
		if err := os.WriteFile(subkeysExportPath, subkeysBytes, 0o600); err == nil {
			log.Success("Subkeys-only file exported to: %s", subkeysExportPath)
			// Also update secsub.gpg for compatibility
			secsubPath := filepath.Join(keyDir, "eng618.secsub.gpg")
			_ = os.WriteFile(secsubPath, subkeysBytes, 0o600)
		}
	}

	// Standard public key export
	pubExportPath := filepath.Join(keyDir, "eng618.public.gpg")
	cmd = execCommand("gpg", "--export", primaryFpr)
	pubBytes, err := cmd.Output()
	if err == nil && len(pubBytes) > 0 {
		if err := os.WriteFile(pubExportPath, pubBytes, 0o600); err == nil {
			log.Success("Public key exported to: %s", pubExportPath)
		}
	}

	// Armored public key export
	pubAscExportPath := filepath.Join(keyDir, "eng618-updated-public.asc")
	cmd = execCommand("gpg", "--armor", "--export", primaryFpr)
	pubAscBytes, err := cmd.Output()
	if err == nil && len(pubAscBytes) > 0 {
		_ = os.WriteFile(pubAscExportPath, pubAscBytes, 0o600)
	}

	log.Message("💡 Re-upload updated files in %s to your Box.com / cloud backup.", keyDir)

	// 12. Send updated key to OpenPGP keyserver
	sendKeyserver, err := ui.Confirm("Publish updated public key to OpenPGP keyserver (keys.openpgp.org)?", true)
	if err == nil && sendKeyserver {
		log.Start("Publishing to OpenPGP keyserver...")
		ksCmd := execCommand("gpg", "--keyserver", "hkps://keys.openpgp.org", "--send-keys", primaryFpr)
		ksCmd.Stdout = log.Writer()
		ksCmd.Stderr = log.ErrorWriter()
		if err := ksCmd.Run(); err != nil {
			log.Warn("Failed to publish to keyserver: %v", err)
		} else {
			log.Success("Public key published to keys.openpgp.org")
		}
	}

	// 13. Upload updated key to GitHub via gh CLI
	uploadGH, err := ui.Confirm("Upload updated public key to GitHub via gh CLI?", true)
	if err == nil && uploadGH && len(pubAscBytes) > 0 {
		ghPath := findGitHubCLI()
		if ghPath != "" {
			log.Start("Uploading public key to GitHub...")
			ghCmd := execCommand(ghPath, "gpg-key", "add", pubAscExportPath)
			ghCmd.Stdout = log.Writer()
			ghCmd.Stderr = log.ErrorWriter()
			if err := ghCmd.Run(); err != nil {
				log.Warn("GitHub upload output: %v (key may already exist on your profile)", err)
			} else {
				log.Success("Public key added to GitHub profile")
			}
		} else {
			log.Message(
				"GitHub CLI (gh) not found or not functional. Update manually under GitHub Settings → SSH and GPG keys.",
			)
		}
	}

	// 14. Remove Master Key from Local Keyring (Subkey-only Mode)
	if !renewKeepMaster {
		removeMaster, err := ui.Confirm(
			"Remove master key from local keyring (keep subkeys only for commit signing)?",
			true,
		)
		if err == nil && removeMaster {
			log.Message("")
			log.Start("Removing master key from local keyring...")

			delFpr, err := resolvePrimaryFingerprint(primaryFpr, verbose)
			if err != nil || delFpr == "" {
				delFpr = primaryFpr
			}
			// Delete entire secret key from local GPG
			delCmd := execCommand("gpg", "--batch", "--yes", "--delete-secret-keys", delFpr)
			delCmd.Stdout = log.Writer()
			delCmd.Stderr = log.ErrorWriter()
			if err := delCmd.Run(); err != nil {
				log.Warn("Failed to delete local master key: %v", err)
			} else {
				// Re-import subkeys only
				impCmd := execCommand("gpg", "--import", subkeysExportPath)
				impCmd.Stdout = log.Writer()
				impCmd.Stderr = log.ErrorWriter()
				if err := impCmd.Run(); err != nil {
					log.Warn("Failed to re-import subkeys: %v", err)
				} else {
					log.Success("Master key removed from local keyring. Subkeys active (sec#).")
				}
			}
		}
	}

	// 15. Test clear-signing
	log.Message("")
	log.Start("Verifying GPG signing function...")
	testCmd := execCommand("gpg", "--clearsign")
	testCmd.Stdin = strings.NewReader("eng GPG key verification test\n")
	var testOut bytes.Buffer
	testCmd.Stdout = &testOut
	testCmd.Stderr = log.ErrorWriter()
	if err := testCmd.Run(); err != nil {
		log.Warn("Signing test failed: %v", err)
		log.Message("Run 'gpg-test' to test signing manually.")
	} else {
		log.Success("GPG signing test passed!")
	}

	log.Success("GPG key renewal completed successfully!")
	return nil
}

// archiveExistingKeys moves existing key files in keyDir into an archive directory.
func archiveExistingKeys(keyDir string, verbose bool) (string, error) {
	nowStr := time.Now().Format("2006-01-02")
	archiveDir := filepath.Join(keyDir, "archive", fmt.Sprintf("%s-expired", nowStr))
	if err := os.MkdirAll(archiveDir, 0o700); err != nil {
		return "", fmt.Errorf("failed to create archive directory: %w", err)
	}

	filesToArchive := []string{
		"eng618.public.gpg",
		"eng618.public.asc",
		"eng618.secret.gpg",
		"subkeys-only.gpg",
		"eng618-updated-public.asc",
		"eng618.public.updated.gpg",
	}

	archivedCount := 0
	for _, fname := range filesToArchive {
		src := filepath.Join(keyDir, fname)
		if _, err := os.Stat(src); err == nil {
			dst := filepath.Join(archiveDir, fname)
			if err := os.Rename(src, dst); err == nil {
				archivedCount++
				log.Verbose(verbose, "Archived %s to %s", fname, archiveDir)
			}
		}
	}

	if archivedCount > 0 {
		log.Success("Archived %d previous key file(s) to: %s", archivedCount, archiveDir)
	}

	return archiveDir, nil
}

// findAndImportMasterKey looks for master secret key files in keyDir and imports them.
func findAndImportMasterKey(keyDir string, verbose bool) (string, error) {
	candidates := []string{
		filepath.Join(keyDir, "eng618.secret.gpg"),
		filepath.Join(keyDir, "secret.gpg"),
		filepath.Join(keyDir, "master.gpg"),
	}

	// Also check any .secret.gpg files in keyDir
	entries, err := os.ReadDir(keyDir)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() &&
				(strings.HasSuffix(entry.Name(), ".secret.gpg") || strings.HasSuffix(entry.Name(), ".sec.gpg")) {
				path := filepath.Join(keyDir, entry.Name())
				candidates = append([]string{path}, candidates...)
			}
		}
	}

	var foundPath string
	for _, cand := range candidates {
		if _, err := os.Stat(cand); err == nil {
			foundPath = cand
			break
		}
	}

	if foundPath == "" {
		// Prompt user for custom path
		inputPath, err := ui.Input("Path to master secret key file", filepath.Join(keyDir, "eng618.secret.gpg"))
		if err != nil || strings.TrimSpace(inputPath) == "" {
			return "", fmt.Errorf("no master secret key file provided")
		}
		foundPath = strings.TrimSpace(inputPath)
	}

	if _, err := os.Stat(foundPath); err != nil {
		return "", fmt.Errorf("master secret key file not found: %s", foundPath)
	}

	confirmImport, err := ui.Confirm(fmt.Sprintf("Import master secret key from %s?", foundPath), true)
	if err != nil || !confirmImport {
		return "", fmt.Errorf("skipped master key import")
	}

	log.Start("Importing master key: %s...", foundPath)
	cmd := execCommand("gpg", "--import", foundPath)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to import master key: %w", err)
	}

	return foundPath, nil
}

// getDefaultGitSigningKey returns the user's configured git signing key if available.
func getDefaultGitSigningKey() string {
	cmd := execCommand("git", "config", "--global", "user.signingkey")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// inspectGPGKey parses key information from gpg --list-keys --with-colons (falling back to --list-secret-keys).
// Returns primary key fingerprint and slice of subkey fingerprints.
func inspectGPGKey(keyID string, verbose bool) (string, []string, error) {
	cmd := execCommand("gpg", "--list-keys", "--with-colons", keyID)
	out, err := cmd.Output()
	if err != nil {
		cmd = execCommand("gpg", "--list-secret-keys", "--with-colons", keyID)
		out, err = cmd.Output()
		if err != nil {
			return "", nil, fmt.Errorf("key not found in local keyring: %s", keyID)
		}
	}

	lines := strings.Split(string(out), "\n")
	var primaryFpr string
	var subkeyFprs []string
	var lastRecordType string

	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) == 0 {
			continue
		}
		recType := fields[0]
		switch recType {
		case "pub", "sub", "ssb", "sec":
			lastRecordType = recType
		case "fpr":
			if len(fields) > 9 && fields[9] != "" {
				fpr := fields[9]
				switch lastRecordType {
				case "pub", "sec":
					if primaryFpr == "" {
						primaryFpr = fpr
					}
				case "sub", "ssb":
					subkeyFprs = append(subkeyFprs, fpr)
				}
			}
		}
	}

	if primaryFpr == "" {
		primaryFpr = keyID
	}

	log.Message("Primary Key Fingerprint: %s", primaryFpr)
	if len(subkeyFprs) > 0 {
		log.Message("Found %d subkey(s):", len(subkeyFprs))
		for _, sub := range subkeyFprs {
			log.Message("  • %s", sub)
		}
	} else {
		log.Message("No subkeys found.")
	}

	return primaryFpr, subkeyFprs, nil
}

// resolvePrimaryFingerprint resolves the full 40-character primary key fingerprint for a given key ID,
// which is strictly required by GnuPG when deleting keys non-interactively in batch mode (--batch --yes).
func resolvePrimaryFingerprint(keyID string, verbose bool) (string, error) {
	primaryFpr, _, err := inspectGPGKey(keyID, verbose)
	if err == nil && len(primaryFpr) >= 32 {
		return primaryFpr, nil
	}

	// Try querying secret keys explicitly with colons output
	cmd := execCommand("gpg", "--list-secret-keys", "--with-colons", keyID)
	out, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		var lastType string
		for _, line := range lines {
			fields := strings.Split(line, ":")
			if len(fields) == 0 {
				continue
			}
			if fields[0] == "sec" || fields[0] == "pub" {
				lastType = fields[0]
			} else if fields[0] == "fpr" && len(fields) > 9 && fields[9] != "" {
				if lastType == "sec" || lastType == "pub" {
					return fields[9], nil
				}
			}
		}
	}

	cleaned := strings.TrimSpace(keyID)
	if len(cleaned) == 40 {
		return cleaned, nil
	}

	if primaryFpr != "" && primaryFpr != keyID {
		return primaryFpr, nil
	}

	return "", fmt.Errorf(
		"unable to resolve 40-character fingerprint for key %s (gpg requires full fingerprint in batch mode)",
		keyID,
	)
}

// findGitHubCLI searches for a working gh executable.
func findGitHubCLI() string {
	if path, err := lookPath("gh"); err == nil {
		cmd := execCommand(path, "--version")
		if err := cmd.Run(); err == nil {
			return path
		}
	}
	for _, candidate := range []string{"/opt/homebrew/bin/gh", "/usr/local/bin/gh", "/usr/bin/gh"} {
		if _, err := stat(candidate); err == nil {
			cmd := execCommand(candidate, "--version")
			if err := cmd.Run(); err == nil {
				return candidate
			}
		}
	}
	return ""
}
