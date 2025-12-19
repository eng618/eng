package system

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/eng618/eng/utils/log"
	"github.com/spf13/cobra"
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
// If not found, it provides instructions and exits.
func ensureGitHubSSH(verbose bool) error {
	log.Verbose(verbose, "Checking for GitHub SSH key")

	homeDir, err := userHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}

	sshKeyPath := filepath.Join(homeDir, ".ssh", "github")

	if _, err := stat(sshKeyPath); err == nil {
		log.Verbose(verbose, "GitHub SSH key found at ~/.ssh/github")
		return nil
	}

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

	return fmt.Errorf("GitHub SSH key not found - please set up SSH access before continuing")
}
