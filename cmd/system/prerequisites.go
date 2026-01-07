package system

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/log"
)

// EnsurePrerequisites checks and installs all prerequisites needed for dotfiles installation.
// Returns an error if any critical prerequisite cannot be satisfied.
func EnsurePrerequisites(verbose bool) error {
	log.Verbose(verbose, "Checking prerequisites for dotfiles installation")

	// Sequential checks with progress logging
	if err := ensureHomebrew(verbose); err != nil {
		return err
	}

	if err := ensureGit(verbose); err != nil {
		return err
	}

	if err := ensureBash(verbose); err != nil {
		return err
	}

	if err := ensureGitHubSSH(verbose); err != nil {
		return err
	}

	log.Verbose(verbose, "All prerequisites satisfied")
	return nil
}

// ensureHomebrew checks if Homebrew is installed, and if not, prompts to install it.
func ensureHomebrew(verbose bool) error {
	log.Verbose(verbose, "Checking for Homebrew")

	_, err := lookPath("brew")
	if err == nil {
		log.Verbose(verbose, "Homebrew is installed")
		return nil
	}

	log.Warn("Homebrew is not installed")
	log.Message("Homebrew is required to install Git and Bash")

	var confirm bool
	prompt := &survey.Confirm{
		Message: "Would you like to install Homebrew now?",
		Default: true,
	}
	err = askOne(prompt, &confirm)
	cobra.CheckErr(err)

	if !confirm {
		return fmt.Errorf("homebrew installation declined - cannot proceed without homebrew")
	}

	log.Start("Installing Homebrew (this may take a few minutes)")

	// Default to system-wide installation
	log.Message("Installing Homebrew system-wide (may require sudo)...")

	// Check for bash
	bashPath, err := lookPath("bash")
	if err != nil {
		return fmt.Errorf("bash is required for homebrew installation but was not found: %w", err)
	}

	// Download the install script to a temporary file to ensure we can run it
	// without piping (which breaks sudo password prompts).
	tmpDir := os.TempDir()
	installScript := filepath.Join(tmpDir, "install_homebrew.sh")

	downloadCmd := execCommand("curl", "-fsSL", "https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh", "-o", installScript)
	downloadCmd.Stdout = log.Writer()
	downloadCmd.Stderr = log.ErrorWriter()
	if err := downloadCmd.Run(); err != nil {
		return fmt.Errorf("failed to download homebrew install script: %w", err)
	}
	defer func() {
		if err := os.Remove(installScript); err != nil {
			log.Warn("Failed to remove temporary install script: %v", err)
		}
	}()

	// Run the script using bash
	installCmd := execCommand(bashPath, installScript)
	installCmd.Stdin = os.Stdin
	installCmd.Stdout = log.Writer()
	installCmd.Stderr = log.ErrorWriter()

	if err := installCmd.Run(); err != nil {
		log.Error("Failed to install Homebrew: %v", err)
		return fmt.Errorf("homebrew installation failed: %w", err)
	}

	log.Success("Homebrew installed successfully")
	return nil
}

// ensureGit checks if Git is installed, and if not, installs it via Homebrew.
func ensureGit(verbose bool) error {
	log.Verbose(verbose, "Checking for Git")

	_, err := lookPath("git")
	if err == nil {
		log.Verbose(verbose, "Git is installed")
		return nil
	}

	log.Warn("Git is not installed")
	log.Start("Installing Git via Homebrew")

	cmd := execCommand("brew", "install", "git")
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()

	if err := cmd.Run(); err != nil {
		log.Error("Failed to install Git: %v", err)
		return fmt.Errorf("git installation failed: %w", err)
	}

	log.Success("Git installed successfully")
	return nil
}

// ensureBash checks if Bash is installed, and if not, installs it via Homebrew.
func ensureBash(verbose bool) error {
	log.Verbose(verbose, "Checking for Bash")

	_, err := lookPath("bash")
	if err == nil {
		log.Verbose(verbose, "Bash is installed")
		return nil
	}

	log.Warn("Bash is not installed")
	log.Start("Installing Bash via Homebrew")

	cmd := execCommand("brew", "install", "bash")
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()

	if err := cmd.Run(); err != nil {
		log.Error("Failed to install Bash: %v", err)
		return fmt.Errorf("bash installation failed: %w", err)
	}

	log.Success("Bash installed successfully")
	return nil
}

// ensureGitHubSSH checks if a valid SSH key for GitHub exists at ~/.ssh/github.
// If not found, attempts to retrieve from Bitwarden vault first, then falls back to manual setup.
func ensureGitHubSSH(verbose bool) error {
	log.Verbose(verbose, "Checking for GitHub SSH key")

	homeDir, err := userHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}

	sshKeyPath := filepath.Join(homeDir, ".ssh", "github")

	// First check if SSH key already exists
	if _, err := stat(sshKeyPath); err == nil {
		log.Verbose(verbose, "GitHub SSH key found at ~/.ssh/github")
		return nil
	}

	// Try to retrieve SSH key from Bitwarden
	log.Verbose(verbose, "SSH key not found locally, checking Bitwarden vault...")
	if err := setupSSHFromBitwarden(sshKeyPath, verbose); err != nil {
		log.Warn("Could not retrieve SSH key from Bitwarden: %v", err)
		log.Message("")
		log.Message("Falling back to manual SSH key setup...")
	} else {
		log.Success("SSH key retrieved from Bitwarden and configured successfully")
		return nil
	}

	// Manual setup instructions
	log.Error("GitHub SSH key not found at ~/.ssh/github")
	log.Message("")
	log.Message("You need a valid SSH key configured for GitHub access.")
	log.Message("The key should be located at: ~/.ssh/github")
	log.Message("")
	log.Message("Your SSH config (~/.ssh/config) should contain:")
	log.Message("")
	log.Message("  Host github.com")
	log.Message("    PreferredAuthentications publickey")
	log.Message("    HostName github.com")
	log.Message("    IdentityFile ~/.ssh/github")
	log.Message("")
	log.Message("For instructions on setting up SSH keys for GitHub, visit:")
	log.Message("https://docs.github.com/en/authentication/connecting-to-github-with-ssh")
	log.Message("")
	log.Message("Alternatively, you can store your SSH key in Bitwarden and this tool")
	log.Message("will automatically retrieve and configure it for you.")
	log.Message("")

	return fmt.Errorf("GitHub SSH key not found - please set up SSH access before continuing")
}

// setupSSHFromBitwarden attempts to retrieve SSH keys from Bitwarden vault and set them up locally.
func setupSSHFromBitwarden(sshKeyPath string, verbose bool) error {
	// Check if Bitwarden is logged in and unlocked
	loggedIn, err := utils.CheckBitwardenLoginStatus()
	if err != nil {
		return fmt.Errorf("failed to check Bitwarden status: %w", err)
	}

	if !loggedIn {
		return utils.UnlockBitwardenVault()
	}

	// Find SSH keys in vault
	sshKeys, err := utils.FindSSHKeysInVault()
	if err != nil {
		return fmt.Errorf("failed to search for SSH keys in Bitwarden: %w", err)
	}

	if len(sshKeys) == 0 {
		return fmt.Errorf("no SSH keys found in Bitwarden vault")
	}

	var selectedKey *utils.BitwardenItem

	// If multiple keys found, let user choose
	if len(sshKeys) > 1 {
		log.Message("Multiple SSH keys found in Bitwarden vault:")
		var options []string
		keyMap := make(map[string]*utils.BitwardenItem)

		for _, key := range sshKeys {
			option := fmt.Sprintf("%s (ID: %s)", key.Name, key.ID)
			options = append(options, option)
			keyMap[option] = &key
		}

		var selected string
		prompt := &survey.Select{
			Message: "Select the SSH key to use for GitHub:",
			Options: options,
		}
		if err := askOne(prompt, &selected); err != nil {
			return fmt.Errorf("key selection canceled: %w", err)
		}

		selectedKey = keyMap[selected]
	} else {
		selectedKey = &sshKeys[0]
		log.Verbose(verbose, "Found SSH key: %s", selectedKey.Name)
	}

	// Extract the SSH key
	sshKey, err := utils.ExtractSSHKeyFromItem(selectedKey)
	if err != nil {
		return fmt.Errorf("failed to extract SSH key: %w", err)
	}

	// Create .ssh directory if it doesn't exist
	sshDir := filepath.Dir(sshKeyPath)
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		return fmt.Errorf("failed to create SSH directory: %w", err)
	}

	// Write SSH key to file
	if err := os.WriteFile(sshKeyPath, []byte(sshKey), 0o600); err != nil {
		return fmt.Errorf("failed to write SSH key: %w", err)
	}

	// Ensure proper permissions
	if err := os.Chmod(sshKeyPath, 0o600); err != nil {
		return fmt.Errorf("failed to set SSH key permissions: %w", err)
	}

	// Create or update SSH config
	sshConfigPath := filepath.Join(sshDir, "config")
	configContent := fmt.Sprintf("Host github.com\n    PreferredAuthentications publickey\n    HostName github.com\n    IdentityFile %s\n", sshKeyPath)

	// Check if config exists and append or create
	if _, err := os.Stat(sshConfigPath); err == nil {
		// Config exists, check if GitHub entry already exists
		existingConfig, err := os.ReadFile(sshConfigPath)
		if err != nil {
			log.Warn("Could not read existing SSH config: %v", err)
		} else {
			configStr := string(existingConfig)
			if !strings.Contains(configStr, "Host github.com") {
				// Append to existing config
				configContent = configStr + "\n" + configContent
			} else {
				// GitHub entry exists, don't modify
				log.Verbose(verbose, "GitHub SSH config entry already exists")
				return nil
			}
		}
	}

	// Write SSH config
	if err := os.WriteFile(sshConfigPath, []byte(configContent), 0o600); err != nil {
		return fmt.Errorf("failed to write SSH config: %w", err)
	}

	log.Verbose(verbose, "SSH key and config configured successfully")
	return nil
}
