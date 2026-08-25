package system

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eng618/eng/internal/bitwarden"
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
		log.Message("You can configure SSH keys later with: eng system setup ssh")
		return nil
	default:
		return setupSSHGenerateFlow(defaultKeyTarget, verbose)
	}

	// Resolve active key after migration/retrieval
	activeKey := FindGitHubSSHKey(sshDir)
	_ = EnsureSSHConfig(activeKey)
	return ValidateGitHubSSHAuth(activeKey, verbose)
}

// setupSSHFromBackupFlow guides the user through selecting a backup directory and migrating keys.
func setupSSHFromBackupFlow(sshDir string, verbose bool) error {
	candidates := DetectSSHBackupCandidates()
	var sourceDir string

	if len(candidates) > 0 {
		options := append(candidates, "Enter a custom backup path...")
		selected, err := ui.Select("Select SSH backup directory to import from:", options, options[0])
		if err == nil && selected != "Enter a custom backup path..." {
			sourceDir = selected
		}
	}

	if sourceDir == "" {
		homeDir, _ := userHomeDir()
		defaultPrompt := filepath.Join(homeDir, "Downloads", "ssh")
		input, err := ui.Input("Enter path to your SSH backup directory:", defaultPrompt)
		if err != nil {
			return fmt.Errorf("backup import canceled: %w", err)
		}
		sourceDir = strings.TrimSpace(input)
	}

	sourceDir = os.ExpandEnv(sourceDir)
	if _, err := stat(sourceDir); err != nil {
		return fmt.Errorf("backup directory not found at: %s", sourceDir)
	}

	return MigrateSSHFromBackup(sourceDir, sshDir, verbose)
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

// DetectSSHBackupCandidates scans common local locations for SSH backup folders containing key files.
func DetectSSHBackupCandidates() []string {
	homeDir, err := userHomeDir()
	if err != nil {
		return nil
	}

	searchPaths := []string{
		filepath.Join(homeDir, "Downloads", "ssh"),
		filepath.Join(homeDir, "Downloads", ".ssh"),
		filepath.Join(homeDir, "Downloads", "Box", "ssh"),
		filepath.Join(homeDir, "Box", "ssh"),
		filepath.Join(homeDir, "Box Sync", "ssh"),
		filepath.Join(homeDir, "Dropbox", "ssh"),
		filepath.Join(homeDir, "Documents", "ssh"),
		filepath.Join(homeDir, "Downloads"),
	}

	var candidates []string
	for _, p := range searchPaths {
		info, err := stat(p)
		if err != nil || !info.IsDir() {
			continue
		}

		if containsSSHKeysOrConfig(p) {
			candidates = append(candidates, p)
		}
	}

	return candidates
}

// containsSSHKeysOrConfig checks if a directory contains SSH private keys, public keys, or config file.
func containsSSHKeysOrConfig(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if name == "config" || name == "known_hosts" || strings.HasSuffix(name, ".pub") {
			return true
		}
		if name == "github" || name == "id_ed25519" || name == "id_rsa" || name == "id_ecdsa" || strings.HasSuffix(name, ".pem") || strings.HasSuffix(name, ".key") {
			return true
		}
		// Check file content for SSH private key header
		filePath := filepath.Join(dir, name)
		if isPrivateKeyFile(filePath) {
			return true
		}
	}

	return false
}

// isPrivateKeyFile inspects the start of a file to check for standard OpenSSH/PEM private key headers.
func isPrivateKeyFile(filePath string) bool {
	info, err := stat(filePath)
	if err != nil || info.IsDir() || info.Size() > 1024*64 {
		return false
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}

	content := string(data)
	return strings.Contains(content, "-----BEGIN") && strings.Contains(content, "PRIVATE KEY-----")
}

// MigrateSSHFromBackup copies all SSH keys, public keys, and config from sourceDir into destDir, enforcing secure permissions.
func MigrateSSHFromBackup(sourceDir, destDir string, verbose bool) error {
	log.Start("Migrating SSH keys and config from %s to %s", sourceDir, destDir)

	if err := os.MkdirAll(destDir, 0o700); err != nil {
		return fmt.Errorf("failed to create SSH directory %s: %w", destDir, err)
	}
	_ = os.Chmod(destDir, 0o700)

	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return fmt.Errorf("failed to read backup directory: %w", err)
	}

	copiedCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		srcPath := filepath.Join(sourceDir, name)
		dstPath := filepath.Join(destDir, name)

		isPub := strings.HasSuffix(name, ".pub")
		isCfg := name == "config"
		isPriv := isPrivateKeyFile(srcPath) || (!isPub && (name == "github" || name == "id_ed25519" || name == "id_rsa" || name == "id_ecdsa" || strings.HasSuffix(name, ".pem") || strings.HasSuffix(name, ".key")))
		isKnownHosts := name == "known_hosts"

		if !isPub && !isCfg && !isPriv && !isKnownHosts {
			log.Verbose(verbose, "Skipping non-SSH file: %s", name)
			continue
		}

		data, err := os.ReadFile(srcPath)
		if err != nil {
			log.Warn("Could not read backup file %s: %v", srcPath, err)
			continue
		}

		// Config file merging
		if isCfg {
			if err := mergeOrWriteSSHConfig(dstPath, string(data)); err != nil {
				log.Warn("Failed to merge SSH config: %v", err)
			} else {
				log.Message("  • Migrated SSH config: %s (mode 0600)", dstPath)
				copiedCount++
			}
			continue
		}

		// File permissions: 0600 for private keys and known_hosts, 0644 for public keys
		mode := os.FileMode(0o600)
		if isPub {
			mode = 0o644
		}

		if err := os.WriteFile(dstPath, data, mode); err != nil {
			log.Warn("Failed to copy %s to %s: %v", name, dstPath, err)
			continue
		}
		_ = os.Chmod(dstPath, mode)

		typeLabel := "private key"
		if isPub {
			typeLabel = "public key"
		} else if isKnownHosts {
			typeLabel = "known hosts"
		}
		log.Message("  • Migrated %s: %s (mode %04o)", typeLabel, name, mode)
		copiedCount++
	}

	if copiedCount == 0 {
		return fmt.Errorf("no valid SSH keys or config files found to migrate in %s", sourceDir)
	}

	log.Success("Successfully migrated %d SSH file(s) from backup", copiedCount)
	return nil
}

// mergeOrWriteSSHConfig writes or merges an SSH config file, preserving existing host blocks.
func mergeOrWriteSSHConfig(destConfigPath, newContent string) error {
	existingBytes, err := os.ReadFile(destConfigPath)
	if err != nil || len(existingBytes) == 0 {
		// New file
		return os.WriteFile(destConfigPath, []byte(strings.TrimSpace(newContent)+"\n"), 0o600)
	}

	existingStr := string(existingBytes)
	// If existing already has all content, do nothing
	if strings.Contains(existingStr, strings.TrimSpace(newContent)) {
		return nil
	}

	// Append missing blocks separated by newline
	merged := strings.TrimRight(existingStr, "\n") + "\n\n" + strings.TrimSpace(newContent) + "\n"
	return os.WriteFile(destConfigPath, []byte(merged), 0o600)
}

// FindGitHubSSHKey locates the primary GitHub SSH private key within sshDir.
func FindGitHubSSHKey(sshDir string) string {
	// 1. Check ~/.ssh/github
	githubKey := filepath.Join(sshDir, "github")
	if _, err := stat(githubKey); err == nil {
		return githubKey
	}

	// 2. Check ~/.ssh/config for Host github.com -> IdentityFile
	cfgPath := filepath.Join(sshDir, "config")
	if cfgData, err := os.ReadFile(cfgPath); err == nil {
		lines := strings.Split(string(cfgData), "\n")
		inGitHubHost := false
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "#") {
				continue
			}
			lower := strings.ToLower(trimmed)
			if strings.HasPrefix(lower, "host ") {
				inGitHubHost = strings.Contains(lower, "github.com")
			} else if inGitHubHost && strings.HasPrefix(lower, "identityfile ") {
				rawPath := strings.TrimSpace(trimmed[len("identityfile "):])
				rawPath = os.ExpandEnv(rawPath)
				if strings.HasPrefix(rawPath, "~/") {
					homeDir, _ := userHomeDir()
					rawPath = filepath.Join(homeDir, rawPath[2:])
				}
				if _, err := stat(rawPath); err == nil {
					return rawPath
				}
			}
		}
	}

	// 3. Check id_ed25519
	ed25519Key := filepath.Join(sshDir, "id_ed25519")
	if _, err := stat(ed25519Key); err == nil {
		return ed25519Key
	}

	// 4. Check id_rsa
	rsaKey := filepath.Join(sshDir, "id_rsa")
	if _, err := stat(rsaKey); err == nil {
		return rsaKey
	}

	// 5. Check id_ecdsa
	ecdsaKey := filepath.Join(sshDir, "id_ecdsa")
	if _, err := stat(ecdsaKey); err == nil {
		return ecdsaKey
	}

	return githubKey
}

// SetupSSHFromBitwarden retrieves SSH keys from Bitwarden vault and configures them locally.
func SetupSSHFromBitwarden(sshKeyPath string, verbose bool) error {
	log.Start("Retrieving SSH keys from Bitwarden vault...")

	session, err := bitwarden.EnsureBitwardenSession()
	if err != nil {
		return fmt.Errorf("failed to access Bitwarden: %w", err)
	}
	_ = session

	sshKeys, err := bitwarden.FindSSHKeysInVault()
	if err != nil {
		return fmt.Errorf("failed to search for SSH keys in Bitwarden: %w", err)
	}

	if len(sshKeys) == 0 {
		return fmt.Errorf("no SSH keys found in Bitwarden vault")
	}

	var selectedKey *bitwarden.BitwardenItem
	if len(sshKeys) > 1 {
		log.Message("Multiple SSH keys found in Bitwarden vault:")
		var options []string
		keyMap := make(map[string]*bitwarden.BitwardenItem)

		for _, key := range sshKeys {
			option := fmt.Sprintf("%s (ID: %s)", key.Name, key.ID)
			options = append(options, option)
			keyMap[option] = &key
		}

		selected, err := ui.Select("Select the SSH key to use for GitHub:", options, options[0])
		if err != nil {
			return fmt.Errorf("key selection canceled: %w", err)
		}

		selectedKey = keyMap[selected]
	} else {
		selectedKey = &sshKeys[0]
		log.Verbose(verbose, "Found SSH key: %s", selectedKey.Name)
	}

	sshKey, err := bitwarden.ExtractSSHKeyFromItem(selectedKey)
	if err != nil {
		return fmt.Errorf("failed to extract SSH key: %w", err)
	}

	sshDir := filepath.Dir(sshKeyPath)
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		return fmt.Errorf("failed to create SSH directory: %w", err)
	}

	if err := os.WriteFile(sshKeyPath, []byte(strings.TrimSpace(sshKey)+"\n"), 0o600); err != nil {
		return fmt.Errorf("failed to write SSH key: %w", err)
	}
	_ = os.Chmod(sshKeyPath, 0o600)

	// If public key is available, write .pub file
	if selectedKey.SSHKey != nil && selectedKey.SSHKey.PublicKey != "" {
		pubKeyPath := sshKeyPath + ".pub"
		_ = os.WriteFile(pubKeyPath, []byte(strings.TrimSpace(selectedKey.SSHKey.PublicKey)+"\n"), 0o644)
		_ = os.Chmod(pubKeyPath, 0o644)
	}

	log.Success("SSH key '%s' written to %s", selectedKey.Name, sshKeyPath)
	return nil
}

// GenerateSSHKey generates a new ed25519 SSH key pair and attempts to register it in GitHub.
func GenerateSSHKey(sshKeyPath string, verbose bool) (bool, error) {
	log.Start("Generating new SSH key pair...")

	sshDir := filepath.Dir(sshKeyPath)
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		return false, fmt.Errorf("failed to create SSH directory: %w", err)
	}

	log.Verbose(verbose, "Generating SSH key: ssh-keygen -t ed25519 -f %s -N '' -C github", sshKeyPath)
	cmd := execCommand("ssh-keygen", "-t", "ed25519", "-f", sshKeyPath, "-N", "", "-C", "github")
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()

	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("failed to generate SSH key: %w", err)
	}
	_ = os.Chmod(sshKeyPath, 0o600)

	pubKeyPath := sshKeyPath + ".pub"
	if _, err := stat(pubKeyPath); err == nil {
		_ = os.Chmod(pubKeyPath, 0o644)
	}

	log.Success("SSH key pair generated successfully at %s", sshKeyPath)

	// Attempt automatic registration via gh CLI
	if err := addSSHKeyToGitHub(sshKeyPath); err != nil {
		log.Warn("Could not automatically add SSH key to GitHub: %v", err)
		log.Message("")
		displaySSHKeyInstructions(sshKeyPath)
		return false, nil
	}

	return true, nil
}

// addSSHKeyToGitHub attempts to add the public key to GitHub via the gh CLI.
func addSSHKeyToGitHub(sshKeyPath string) error {
	publicKeyPath := sshKeyPath + ".pub"

	if _, err := lookPath("gh"); err != nil {
		return fmt.Errorf("gh CLI not found")
	}

	authCmd := execCommand("gh", "auth", "status")
	if err := authCmd.Run(); err != nil {
		return fmt.Errorf("gh CLI not authenticated")
	}

	log.Start("Adding SSH key to GitHub via gh CLI...")
	hostname, _ := os.Hostname()
	title := "eng-github"
	if hostname != "" {
		title = fmt.Sprintf("eng-github (%s)", hostname)
	}

	addCmd := execCommand("gh", "ssh-key", "add", publicKeyPath, "--title", title)
	addCmd.Stdout = log.Writer()
	addCmd.Stderr = log.ErrorWriter()

	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("gh ssh-key add failed: %w", err)
	}

	log.Success("SSH key successfully added to your GitHub account")
	return nil
}

// displaySSHKeyInstructions displays public key and manual registration steps.
func displaySSHKeyInstructions(sshKeyPath string) {
	publicKeyPath := sshKeyPath + ".pub"

	pubKeyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		log.Error("Could not read public key: %v", err)
		return
	}

	log.Message("Your SSH public key:")
	log.Message("---------------------------------------------")
	log.Message("%s", strings.TrimSpace(string(pubKeyBytes)))
	log.Message("---------------------------------------------")
	log.Message("")
	log.Message("Steps to add your SSH key to GitHub:")
	log.Message("  1. Copy the public key shown above")
	log.Message("  2. Open: https://github.com/settings/keys")
	log.Message("  3. Click 'New SSH key' and paste your key")
	log.Message("  4. Click 'Add SSH key'")
	log.Message("")
}

// waitForManualGitHubKeyRegistration prompts user to confirm they added the key on GitHub.
func waitForManualGitHubKeyRegistration(sshKeyPath string) error {
	log.Message("After adding the key to GitHub, press Enter to continue...")
	_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
	return nil
}

// EnsureSSHConfig ensures ~/.ssh/config contains the Host github.com configuration block.
func EnsureSSHConfig(sshKeyPath string) error {
	sshDir := filepath.Dir(sshKeyPath)
	sshConfigPath := filepath.Join(sshDir, "config")

	configBlock := fmt.Sprintf(
		"Host github.com\n    PreferredAuthentications publickey\n    HostName github.com\n    IdentityFile %s\n",
		sshKeyPath,
	)

	if _, err := stat(sshConfigPath); err == nil {
		existingBytes, err := os.ReadFile(sshConfigPath)
		if err != nil {
			log.Warn("Could not read existing SSH config: %v", err)
		} else {
			configStr := string(existingBytes)
			for _, line := range strings.Split(configStr, "\n") {
				trimmed := strings.TrimSpace(line)
				if !strings.HasPrefix(trimmed, "#") && strings.Contains(trimmed, "Host github.com") {
					log.Verbose(true, "GitHub SSH config entry already exists in %s", sshConfigPath)
					return nil
				}
			}
			if !strings.HasSuffix(configStr, "\n") {
				configStr += "\n"
			}
			configBlock = configStr + "\n" + configBlock
		}
	}

	if err := os.WriteFile(sshConfigPath, []byte(configBlock), 0o600); err != nil {
		return fmt.Errorf("failed to write SSH config: %w", err)
	}
	_ = os.Chmod(sshConfigPath, 0o600)

	log.Success("SSH config updated for GitHub access (%s)", sshConfigPath)
	return nil
}

// ValidateGitHubSSHAuth verifies GitHub SSH connectivity using the specified key.
func ValidateGitHubSSHAuth(sshKeyPath string, verbose bool) error {
	log.Start("Validating GitHub SSH authentication...")
	log.Verbose(verbose, "Testing SSH authentication using key: %s", sshKeyPath)

	cmd := execCommand(
		"ssh",
		"-T",
		"-o", "BatchMode=yes",
		"-o", "IdentitiesOnly=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		"-i", sshKeyPath,
		"git@github.com",
	)
	output, err := cmd.CombinedOutput()
	outStr := strings.TrimSpace(string(output))
	if outStr != "" {
		log.Verbose(verbose, "GitHub SSH validation output: %s", outStr)
	}

	if strings.Contains(outStr, "successfully authenticated") || strings.Contains(outStr, "Hi ") {
		log.Success("GitHub SSH authentication validated with %s", sshKeyPath)
		return nil
	}

	if err != nil {
		return fmt.Errorf("github SSH validation failed with %s: %w", sshKeyPath, err)
	}

	return fmt.Errorf("github SSH validation failed with %s", sshKeyPath)
}
