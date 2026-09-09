package setup

import (
	"os"
	"path/filepath"

	"github.com/eng618/eng/internal/log"
)

func setupGPGPermissions(verbose bool) {
	log.Verbose(verbose, "Checking GPG directory permissions and configuration...")

	homeDir, err := userHomeDir()
	if err != nil {
		log.Error("Could not determine home directory: %v", err)
		return
	}

	gpgDir := filepath.Join(homeDir, ".gnupg")
	if _, err := stat(gpgDir); err != nil {
		log.Verbose(verbose, "GPG directory does not exist, skipping permissions fix")
		return
	}

	log.Start("Fixing GPG directory and configuration permissions...")
	cmd := execCommand("chmod", "700", gpgDir)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()
	if err := cmd.Run(); err != nil {
		log.Error("Failed to fix GPG directory permissions: %v", err)
	} else {
		log.Success("GPG directory permissions fixed")
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

	// Restart gpg-agent to apply configuration
	log.Verbose(verbose, "Restarting gpg-agent to apply configuration...")
	killCmd := execCommand("gpgconf", "--kill", "gpg-agent")
	_ = killCmd.Run()

	launchCmd := execCommand("gpgconf", "--launch", "gpg-agent")
	if err := launchCmd.Run(); err != nil {
		log.Verbose(verbose, "gpgconf --launch gpg-agent output: %v", err)
	} else {
		log.Success("GPG agent restarted with updated configuration")
	}
}
