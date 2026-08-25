package system

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/bitwarden"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
)

// EnsurePrerequisites checks and installs core prerequisites needed for setup flows.
// SSH is handled contextually by callers that need GitHub access.
func EnsurePrerequisites(verbose bool) error {
	log.Verbose(verbose, "Checking prerequisites for dotfiles and system setup")

	// 1. Homebrew check / offer
	if err := ensureHomebrew(verbose); err != nil {
		return err
	}

	// 2. Git check / installation
	if err := ensureGit(verbose); err != nil {
		return err
	}

	// 3. Bash check / installation
	if err := ensureBash(verbose); err != nil {
		return err
	}

	// 4. Zsh check / installation
	if err := ensureZsh(verbose); err != nil {
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

	// Check Linuxbrew default location if not in PATH yet
	if _, err := stat("/home/linuxbrew/.linuxbrew/bin/brew"); err == nil {
		_ = os.Setenv("PATH", os.Getenv("PATH")+":/home/linuxbrew/.linuxbrew/bin")
		log.Verbose(verbose, "Found Homebrew at /home/linuxbrew/.linuxbrew/bin/brew")
		return nil
	}

	distro := detectDistro()
	isMac := distro.IsMacOS()

	if isMac {
		log.Warn("Homebrew is not installed")
		log.Message("Homebrew is required on macOS to install Git and Bash")
	} else {
		log.Message("Homebrew (Linuxbrew) is not currently installed")
	}

	confirmPrompt := "Would you like to install Homebrew now?"
	if !isMac {
		confirmPrompt = "Would you like to install Homebrew (Linuxbrew) now?"
	}

	confirm, err := ui.Confirm(confirmPrompt, isMac)
	if err != nil {
		if isMac {
			cobra.CheckErr(err)
		}
		return nil
	}

	if !confirm {
		if isMac {
			return fmt.Errorf("homebrew installation declined - cannot proceed on macOS without homebrew")
		}
		log.Verbose(verbose, "Homebrew installation skipped on Linux; continuing with native tools")
		return nil
	}

	log.Start("Installing Homebrew (this may take a few minutes)")
	log.Message("Installing Homebrew system-wide (may require sudo)...")

	// Check for bash
	bashPath, err := lookPath("bash")
	if err != nil {
		return fmt.Errorf("bash is required for homebrew installation but was not found: %w", err)
	}

	// Download the install script to a temporary file
	tmpDir := os.TempDir()
	installScript := filepath.Join(tmpDir, "install_homebrew.sh")

	downloadCmd := execCommand(
		"curl",
		"-fsSL",
		"https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh",
		"-o",
		installScript,
	)
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

	// If on Linux, update PATH for current process
	if _, err := stat("/home/linuxbrew/.linuxbrew/bin/brew"); err == nil {
		_ = os.Setenv("PATH", os.Getenv("PATH")+":/home/linuxbrew/.linuxbrew/bin")
	}

	log.Success("Homebrew installed successfully")
	return nil
}

// ensureGit checks if Git is installed, and if not, installs it via available package manager.
func ensureGit(verbose bool) error {
	log.Verbose(verbose, "Checking for Git")

	_, err := lookPath("git")
	if err == nil {
		log.Verbose(verbose, "Git is installed")
		return nil
	}

	log.Warn("Git is not installed")
	distro := detectDistro()

	if _, err := lookPath("brew"); err == nil {
		log.Start("Installing Git via Homebrew")
		cmd := execCommand("brew", "install", "git")
		cmd.Stdout = log.Writer()
		cmd.Stderr = log.ErrorWriter()
		if err := cmd.Run(); err != nil {
			log.Error("Failed to install Git via Homebrew: %v", err)
			return fmt.Errorf("git installation failed: %w", err)
		}
	} else if distro.IsFedora() {
		log.Start("Installing Git via DNF")
		cmd := execCommand("sudo", "dnf", "install", "-y", "git")
		cmd.Stdout = log.Writer()
		cmd.Stderr = log.ErrorWriter()
		if err := cmd.Run(); err != nil {
			log.Error("Failed to install Git via DNF: %v", err)
			return fmt.Errorf("git installation failed: %w", err)
		}
	} else if distro.IsDebianUbuntu() {
		log.Start("Installing Git via APT")
		cmd := execCommand("sudo", "apt-get", "install", "-y", "git")
		cmd.Stdout = log.Writer()
		cmd.Stderr = log.ErrorWriter()
		if err := cmd.Run(); err != nil {
			log.Error("Failed to install Git via APT: %v", err)
			return fmt.Errorf("git installation failed: %w", err)
		}
	} else {
		return fmt.Errorf("git is not installed and no supported package manager (brew, dnf, apt) was found")
	}

	log.Success("Git installed successfully")
	return nil
}

// ensureBash checks if Bash is installed, and if not, installs it via available package manager.
func ensureBash(verbose bool) error {
	log.Verbose(verbose, "Checking for Bash")

	_, err := lookPath("bash")
	if err == nil {
		log.Verbose(verbose, "Bash is installed")
		return nil
	}

	log.Warn("Bash is not installed")
	distro := detectDistro()

	if _, err := lookPath("brew"); err == nil {
		log.Start("Installing Bash via Homebrew")
		cmd := execCommand("brew", "install", "bash")
		cmd.Stdout = log.Writer()
		cmd.Stderr = log.ErrorWriter()
		if err := cmd.Run(); err != nil {
			log.Error("Failed to install Bash via Homebrew: %v", err)
			return fmt.Errorf("bash installation failed: %w", err)
		}
	} else if distro.IsFedora() {
		log.Start("Installing Bash via DNF")
		cmd := execCommand("sudo", "dnf", "install", "-y", "bash")
		cmd.Stdout = log.Writer()
		cmd.Stderr = log.ErrorWriter()
		if err := cmd.Run(); err != nil {
			log.Error("Failed to install Bash via DNF: %v", err)
			return fmt.Errorf("bash installation failed: %w", err)
		}
	} else if distro.IsDebianUbuntu() {
		log.Start("Installing Bash via APT")
		cmd := execCommand("sudo", "apt-get", "install", "-y", "bash")
		cmd.Stdout = log.Writer()
		cmd.Stderr = log.ErrorWriter()
		if err := cmd.Run(); err != nil {
			log.Error("Failed to install Bash via APT: %v", err)
			return fmt.Errorf("bash installation failed: %w", err)
		}
	} else {
		return fmt.Errorf("bash is not installed and no supported package manager (brew, dnf, apt) was found")
	}

	log.Success("Bash installed successfully")
	return nil
}

// ensureZsh checks if Zsh is installed, and if not, prompts to install it via available package manager.
func ensureZsh(verbose bool) error {
	log.Verbose(verbose, "Checking for Zsh")

	_, err := lookPath("zsh")
	if err == nil {
		log.Verbose(verbose, "Zsh is installed")
		return nil
	}

	log.Warn("Zsh is not installed")
	confirm, err := ui.Confirm("Would you like to install Zsh now?", true)
	if err != nil {
		return err
	}
	if !confirm {
		return fmt.Errorf("zsh installation declined - zsh is required for shell and dotfiles setup")
	}

	distro := detectDistro()

	if _, err := lookPath("brew"); err == nil {
		log.Start("Installing Zsh via Homebrew")
		cmd := execCommand("brew", "install", "zsh")
		cmd.Stdout = log.Writer()
		cmd.Stderr = log.ErrorWriter()
		if err := cmd.Run(); err != nil {
			log.Error("Failed to install Zsh via Homebrew: %v", err)
			return fmt.Errorf("zsh installation failed: %w", err)
		}
	} else if distro.IsFedora() {
		log.Start("Installing Zsh via DNF")
		cmd := execCommand("sudo", "dnf", "install", "-y", "zsh")
		cmd.Stdout = log.Writer()
		cmd.Stderr = log.ErrorWriter()
		if err := cmd.Run(); err != nil {
			log.Error("Failed to install Zsh via DNF: %v", err)
			return fmt.Errorf("zsh installation failed: %w", err)
		}
	} else if distro.IsDebianUbuntu() {
		log.Start("Installing Zsh via APT")
		cmd := execCommand("sudo", "apt-get", "install", "-y", "zsh")
		cmd.Stdout = log.Writer()
		cmd.Stderr = log.ErrorWriter()
		if err := cmd.Run(); err != nil {
			log.Error("Failed to install Zsh via APT: %v", err)
			return fmt.Errorf("zsh installation failed: %w", err)
		}
	} else {
		return fmt.Errorf("zsh is not installed and no supported package manager (brew, dnf, apt) was found")
	}

	log.Success("Zsh installed successfully")
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
	// Ensure Bitwarden session is unlocked (prompts user interactively if needed)
	session, err := bitwarden.EnsureBitwardenSession()
	if err != nil {
		return fmt.Errorf("failed to access Bitwarden: %w", err)
	}
	_ = session // Session is set in environment by EnsureBitwardenSession

	// Find SSH keys in vault
	sshKeys, err := bitwarden.FindSSHKeysInVault()
	if err != nil {
		return fmt.Errorf("failed to search for SSH keys in Bitwarden: %w", err)
	}

	if len(sshKeys) == 0 {
		return fmt.Errorf("no SSH keys found in Bitwarden vault")
	}

	var selectedKey *bitwarden.BitwardenItem

	// If multiple keys found, let user choose
	if len(sshKeys) > 1 {
		log.Message("Multiple SSH keys found in Bitwarden vault:")
		var options []string
		keyMap := make(map[string]*bitwarden.BitwardenItem)

		for _, key := range sshKeys {
			option := fmt.Sprintf("%s (ID: %s)", key.Name, key.ID)
			options = append(options, option)
			keyMap[option] = &key
		}

		selected, err := ui.Select("Select the SSH key to use for GitHub:", options, "")
		if err != nil {
			return fmt.Errorf("key selection canceled: %w", err)
		}

		selectedKey = keyMap[selected]
	} else {
		selectedKey = &sshKeys[0]
		log.Verbose(verbose, "Found SSH key: %s", selectedKey.Name)
	}

	// Extract the SSH key
	sshKey, err := bitwarden.ExtractSSHKeyFromItem(selectedKey)
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
	configContent := fmt.Sprintf(
		"Host github.com\n    PreferredAuthentications publickey\n    HostName github.com\n    IdentityFile %s\n",
		sshKeyPath,
	)

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
