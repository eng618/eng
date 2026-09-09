package ssh

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
)

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
		if name == "github" || name == "id_ed25519" || name == "id_rsa" || name == "id_ecdsa" ||
			strings.HasSuffix(name, ".pem") ||
			strings.HasSuffix(name, ".key") {
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
		isPriv := isPrivateKeyFile(srcPath) ||
			(!isPub && (name == "github" || name == "id_ed25519" || name == "id_rsa" || name == "id_ecdsa" || strings.HasSuffix(name, ".pem") || strings.HasSuffix(name, ".key")))
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
