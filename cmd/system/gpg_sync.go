package system

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
)

var (
	syncKeyID     string
	syncKeyserver string
	syncGitHubUser string
)

var SyncGPGCmd = &cobra.Command{
	Use:     "sync",
	Aliases: []string{"pull", "fetch", "refresh"},
	Short:   "Sync updated GPG public key and expiration dates from keyservers",
	Long: `Pull and sync updated GPG public keys and signatures from OpenPGP keyservers
and GitHub on secondary devices without needing access to your master key.`,
	RunE: func(cmd *cobra.Command, _args []string) error {
		return syncGPG(cmdutil.IsVerbose(cmd))
	},
}

func init() {
	SyncGPGCmd.Flags().StringVarP(&syncKeyID, "key-id", "k", "", "GPG key ID (defaults to git user.signingkey)")
	SyncGPGCmd.Flags().StringVar(&syncKeyserver, "keyserver", "hkps://keys.openpgp.org", "Keyserver URL")
	SyncGPGCmd.Flags().StringVar(&syncGitHubUser, "github-user", "", "GitHub username to fetch public key from (e.g. eng618)")
}

// syncGPG pulls the latest public key version from keyservers and GitHub to update local keyring.
func syncGPG(verbose bool) error {
	log.Verbose(verbose, "Starting GPG key sync process...")

	// 1. Ensure GPG dependencies exist
	if err := ensureGPGDependencies(verbose); err != nil {
		return err
	}

	// 2. Resolve Key ID
	keyID := strings.TrimSpace(syncKeyID)
	if keyID == "" {
		keyID = getDefaultGitSigningKey()
	}

	if keyID == "" {
		var err error
		keyID, err = ui.Input("Enter GPG key ID (long format or fingerprint)", "")
		if err != nil {
			return fmt.Errorf("canceled: %w", err)
		}
		keyID = strings.TrimSpace(keyID)
	}

	if keyID == "" {
		return fmt.Errorf("key ID is required")
	}

	// Validate hex key ID format (16 to 40 hex chars)
	validKeyID := regexp.MustCompile(`^[0-9A-Fa-f]{16,40}$`)
	if !validKeyID.MatchString(keyID) {
		return fmt.Errorf("invalid GPG key ID format: must be 16 to 40 hexadecimal characters")
	}

	log.Start("Syncing GPG public key for: %s", keyID)

	// 3. Fetch from OpenPGP keyserver
	log.Message("Fetching latest public key from keyserver (%s)...", syncKeyserver)
	ksCmd := execCommand("gpg", "--keyserver", syncKeyserver, "--recv-keys", keyID)
	ksCmd.Stdout = log.Writer()
	ksCmd.Stderr = log.ErrorWriter()
	ksErr := ksCmd.Run()
	if ksErr != nil {
		log.Warn("Keyserver fetch notice: %v", ksErr)
	} else {
		log.Success("Public key updated from keyserver: %s", syncKeyserver)
	}

	// 4. Fetch from GitHub if keyserver didn't succeed or if user specified username
	ghUser := strings.TrimSpace(syncGitHubUser)
	if ghUser == "" {
		// Attempt to detect GitHub username from gh CLI or git email
		if ghPath := findGitHubCLI(); ghPath != "" {
			cmd := execCommand(ghPath, "api", "user", "-q", ".login")
			if out, err := cmd.Output(); err == nil {
				ghUser = strings.TrimSpace(string(out))
			}
		}
	}

	if ghUser != "" {
		log.Message("Checking GitHub public keys for user '%s'...", ghUser)
		ghKeyURL := fmt.Sprintf("https://github.com/%s.gpg", ghUser)
		if err := fetchAndImportGPGURL(ghKeyURL, verbose); err != nil {
			log.Verbose(verbose, "GitHub key import notice: %v", err)
		} else {
			log.Success("Public keys imported from GitHub (%s)", ghKeyURL)
		}
	}

	// 5. Ensure Ultimate Trust Level for Key
	log.Message("")
	log.Start("Updating key trust setting...")
	if err := setGPGTrust(keyID, verbose); err != nil {
		log.Warn("Trust update notice: %v", err)
	}

	// 6. Inspect local key status after sync
	log.Message("")
	log.Start("Inspecting updated GPG key status...")
	primaryFpr, subkeyFprs, err := inspectGPGKey(keyID, verbose)
	if err != nil {
		log.Warn("Key status inspection notice: %v", err)
	} else {
		log.Success("Key ID %s (Fingerprint %s) synced cleanly.", keyID, primaryFpr)
		if len(subkeyFprs) > 0 {
			log.Message("Active subkeys: %d", len(subkeyFprs))
		}
	}

	// 7. Verify signing functionality
	log.Message("")
	log.Start("Verifying GPG signing function...")
	testCmd := execCommand("gpg", "--clearsign")
	testCmd.Stdin = strings.NewReader("eng GPG sync verification test\n")
	var testOut bytes.Buffer
	testCmd.Stdout = &testOut
	testCmd.Stderr = log.ErrorWriter()
	if err := testCmd.Run(); err != nil {
		log.Warn("Signing test notice: %v", err)
		log.Message("Run 'gpg-test' to test signing manually.")
	} else {
		log.Success("GPG signing test passed!")
	}

	log.Success("GPG key sync completed successfully!")
	return nil
}

// fetchAndImportGPGURL downloads a public key block from a URL and imports it into gpg.
func fetchAndImportGPGURL(url string, verbose bool) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download key from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP status %d fetching %s", resp.StatusCode, url)
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if buf.Len() == 0 {
		return fmt.Errorf("received empty key response from %s", url)
	}

	tempFile := filepath.Join(os.TempDir(), "eng_github_key.asc")
	if err := os.WriteFile(tempFile, buf.Bytes(), 0o600); err != nil {
		return fmt.Errorf("failed to write temp key file: %w", err)
	}
	defer os.Remove(tempFile)

	cmd := execCommand("gpg", "--import", tempFile)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to import key from %s: %w", url, err)
	}

	return nil
}
