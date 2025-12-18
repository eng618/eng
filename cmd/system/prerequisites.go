package system

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/eng618/eng/utils/log"
	"github.com/spf13/cobra"
)

// EnsurePrerequisites checks and installs all prerequisites needed for dotfiles installation.
// Returns an error if any critical prerequisite cannot be satisfied.
func EnsurePrerequisites() error {
	log.Start("Checking prerequisites for dotfiles installation")

	// Sequential checks with progress logging
	if err := ensureHomebrew(); err != nil {
		return err
	}

	if err := ensureGit(); err != nil {
		return err
	}

	if err := ensureBash(); err != nil {
		return err
	}

	if err := ensureGitHubSSH(); err != nil {
		return err
	}

	log.Success("All prerequisites satisfied")
	return nil
}

// ensureHomebrew checks if Homebrew is installed, and if not, prompts to install it.
func ensureHomebrew() error {
	log.Start("Checking for Homebrew")

	_, err := exec.LookPath("brew")
	if err == nil {
		log.Success("Homebrew is installed")
		return nil
	}

	log.Warn("Homebrew is not installed")
	log.Message("Homebrew is required to install Git and Bash")

	var confirm bool
	prompt := &survey.Confirm{
		Message: "Would you like to install Homebrew now?",
		Default: true,
	}
	err = survey.AskOne(prompt, &confirm)
	cobra.CheckErr(err)

	if !confirm {
		return fmt.Errorf("homebrew installation declined - cannot proceed without homebrew")
	}

	log.Start("Installing Homebrew (this may take a few minutes)")

	// Determine which shell to use for installation
	shellPath := "/bin/bash"
	if _, err := os.Stat("/bin/bash"); os.IsNotExist(err) {
		shellPath = "/bin/zsh"
		log.Message("Using zsh for Homebrew installation (bash not found)")
	}

	// Ask user to choose installation method
	var installChoice string
	installPrompt := &survey.Select{
		Message: "Choose Homebrew installation method:",
		Options: []string{
			"user-only - Install to ~/.homebrew (no sudo required, user-only access)",
			"system-wide - Install to /usr/local or /opt/homebrew (requires sudo, system-wide access)",
		},
		Default: "user-only - Install to ~/.homebrew (no sudo required, user-only access)",
		Help:    "User-only installation doesn't require admin access but only works for the current user. System-wide requires sudo but makes Homebrew available to all users.",
	}

	if err := survey.AskOne(installPrompt, &installChoice); err != nil {
		return fmt.Errorf("installation method selection failed: %w", err)
	}

	var installCmd *exec.Cmd
	var env []string

	if installChoice == "user-only - Install to ~/.homebrew (no sudo required, user-only access)" {
		log.Message("Installing Homebrew in user-only mode (no sudo required)...")
		installCmd = exec.Command(shellPath, "-c", "curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh | NONINTERACTIVE=1 bash -s -- --prefix=$HOME/.homebrew")
		env = append(os.Environ(), "NONINTERACTIVE=1")
	} else {
		log.Message("Installing Homebrew system-wide (may require sudo)...")
		installCmd = exec.Command(shellPath, "-c", "curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh | bash")
		env = os.Environ() // No NONINTERACTIVE for system-wide so sudo can prompt
	}

	installCmd.Stdin = os.Stdin
	installCmd.Stdout = log.Writer()
	installCmd.Stderr = log.ErrorWriter()
	installCmd.Env = env

	if err := installCmd.Run(); err != nil {
		log.Error("Failed to install Homebrew: %v", err)
		return fmt.Errorf("homebrew installation failed: %w", err)
	}

	log.Success("Homebrew installed successfully")
	return nil
}

// ensureGit checks if Git is installed, and if not, installs it via Homebrew.
func ensureGit() error {
	log.Start("Checking for Git")

	_, err := exec.LookPath("git")
	if err == nil {
		log.Success("Git is installed")
		return nil
	}

	log.Warn("Git is not installed")
	log.Start("Installing Git via Homebrew")

	cmd := exec.Command("brew", "install", "git")
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
func ensureBash() error {
	log.Start("Checking for Bash")

	_, err := exec.LookPath("bash")
	if err == nil {
		log.Success("Bash is installed")
		return nil
	}

	log.Warn("Bash is not installed")
	log.Start("Installing Bash via Homebrew")

	cmd := exec.Command("brew", "install", "bash")
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
func ensureGitHubSSH() error {
	log.Start("Checking for GitHub SSH key")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}

	sshKeyPath := filepath.Join(homeDir, ".ssh", "github")

	if _, err := os.Stat(sshKeyPath); err == nil {
		log.Success("GitHub SSH key found at ~/.ssh/github")
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
