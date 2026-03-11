package system

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/utils"
	"github.com/eng618/eng/internal/utils/log"
)

var SetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup development tools",
	Long: `Setup various development tools.
Running this command without subcommands will run all setup steps:
- Oh My Zsh
- ASDF plugins
- Dotfiles installation`,
	Run: func(cmd *cobra.Command, _args []string) {
		verbose := utils.IsVerbose(cmd)
		if err := EnsurePrerequisites(verbose); err != nil {
			log.Fatal("Prerequisites check failed: %v", err)
		}

		setupOhMyZsh(verbose)
		setupASDF(verbose)
		if err := setupDotfiles(verbose); err != nil {
			log.Error("Dotfiles setup failed: %v", err)
		}
		setupSoftware(verbose)
	},
}

var SetupASDFCmd = &cobra.Command{
	Use:   "asdf",
	Short: "Setup asdf plugins from $HOME/.tool-versions",
	Long:  `Reads $HOME/.tool-versions and installs asdf plugins listed there.`,
	Run: func(cmd *cobra.Command, args []string) {
		setupASDF(utils.IsVerbose(cmd))
	},
}

var SetupDotfilesCmd = &cobra.Command{
	Use:   "dotfiles",
	Short: "Setup dotfiles from your git repository",
	Long: `Setup dotfiles from your git repository. This command will:
	- Check and install prerequisites (Homebrew, Git, Bash)
	- Setup SSH keys for GitHub when required by the repository URL
  - Clone your dotfiles repository as a bare repository
  - Backup any conflicting files
  - Checkout dotfiles to your home directory
  - Initialize git submodules
  - Configure git to hide untracked files`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := setupDotfiles(utils.IsVerbose(cmd)); err != nil {
			log.Fatal("Dotfiles setup failed: %v", err)
		}
	},
}

var SetupOhMyZshCmd = &cobra.Command{
	Use:   "oh-my-zsh",
	Short: "Install Oh My Zsh",
	Long:  `Downloads and installs Oh My Zsh. Skips if already installed.`,
	Run: func(cmd *cobra.Command, args []string) {
		setupOhMyZsh(utils.IsVerbose(cmd))
	},
}

var SetupSSHCmd = &cobra.Command{
	Use:   "ssh",
	Short: "Setup SSH keys for GitHub access",
	Long: `Setup SSH keys for GitHub access. This command will:
  - Check for existing SSH keys
  - Attempt to retrieve SSH keys from Bitwarden vault
  - Generate new SSH keys if none found
  - Configure SSH config for GitHub`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := setupSSH(utils.IsVerbose(cmd)); err != nil {
			log.Fatal("SSH setup failed: %v", err)
		}
	},
}

func init() {
	SetupCmd.AddCommand(SetupASDFCmd)
	SetupCmd.AddCommand(SetupDotfilesCmd)
	SetupCmd.AddCommand(SetupOhMyZshCmd)
	SetupCmd.AddCommand(SetupSSHCmd)
}

func setupASDF(verbose bool) {
	log.Verbose(verbose, "Starting ASDF setup...")
	// Check error for os.UserHomeDir
	homeDir, err := userHomeDir()
	if err != nil {
		log.Error("Could not determine home directory: %v", err)
		return
	}
	toolVersionsPath := filepath.Join(homeDir, ".tool-versions")
	file, err := os.Open(toolVersionsPath)
	if err != nil {
		log.Error("Could not open %s: %v", toolVersionsPath, err)
		return
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Error("Error closing file %s: %v", toolVersionsPath, cerr)
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 1 {
			continue
		}
		plugin := fields[0]
		cmd := execCommand("asdf", "plugin", "add", plugin)
		cmd.Stdout = log.Writer()
		cmd.Stderr = log.ErrorWriter()
		if err := cmd.Run(); err != nil {
			log.Error("Failed to add asdf plugin '%s': %v", plugin, err)
		} else {
			log.Success("Added asdf plugin: %s", plugin)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Error("Error reading %s: %v", toolVersionsPath, err)
		return
	}
	// Install all plugins
	installCmd := execCommand("asdf", "install")
	installCmd.Stdout = log.Writer()
	installCmd.Stderr = log.ErrorWriter()
	log.Start("Running 'asdf install' to install all plugins...")
	if err := installCmd.Run(); err != nil {
		log.Error("Failed to run 'asdf install': %v", err)
	} else {
		log.Success("All asdf plugins installed successfully.")
	}
}

func setupOhMyZsh(verbose bool) {
	homeDir, err := userHomeDir()
	if err != nil {
		log.Error("Could not determine home directory: %v", err)
		return
	}
	omzPath := filepath.Join(homeDir, ".oh-my-zsh")
	if _, err := stat(omzPath); err == nil {
		log.Verbose(verbose, "Oh My Zsh found at %s", omzPath)
		log.Success("Oh My Zsh is already installed")
		return
	}

	log.Start("Installing Oh My Zsh...")
	// Use --unattended to prevent switching shell immediately
	cmd := execCommand(
		"sh",
		"-c",
		"curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh | sh -s -- --unattended",
	)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()
	if err := cmd.Run(); err != nil {
		log.Error("Failed to install Oh My Zsh: %v", err)
	} else {
		log.Success("Oh My Zsh installed successfully")
	}
}

// setupDotfiles sets up dotfiles by checking prerequisites and running the install command.
func setupDotfiles(verbose bool) error {
	log.Verbose(verbose, "Starting dotfiles setup...")

	// Get the path to the current executable
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	log.Start("Running dotfiles install...")
	// Run dependencies install command
	args := []string{"dotfiles", "install"}
	if verbose {
		args = append(args, "-v")
	}
	cmd := execCommand(exe, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("dotfiles install failed: %w", err)
	}

	return nil
}

func setupSoftware(verbose bool) {
	log.Verbose(verbose, "Checking software...")

	allSoftware := getSoftwareList()
	var toInstall []Software
	var optionalOptions []string
	optionalSoftwareMap := make(map[string]Software)

	// Filter and check
	for _, sw := range allSoftware {
		// Skip if OS mismatch
		if sw.OS != "" && sw.OS != runtime.GOOS {
			log.Verbose(verbose, "Skipping %s (OS mismatch: need %s, have %s)", sw.Name, sw.OS, runtime.GOOS)
			continue
		}

		if sw.Check() {
			log.Verbose(verbose, "%s is already installed.", sw.Name)
			continue
		}

		if !sw.Optional {
			toInstall = append(toInstall, sw)
		} else {
			optionalOptions = append(optionalOptions, sw.Name)
			optionalSoftwareMap[sw.Name] = sw
		}
	}

	// Prompt for optional software
	if len(optionalOptions) > 0 {
		var selected []string
		prompt := &survey.MultiSelect{
			Message: "Select additional software to install:",
			Options: optionalOptions,
		}
		if err := askOne(prompt, &selected); err != nil {
			log.Error("Selection canceled: %v", err)
			return
		}
		for _, name := range selected {
			toInstall = append(toInstall, optionalSoftwareMap[name])
		}
	}

	// Install loop
	for _, sw := range toInstall {
		log.Start("Installing %s...", sw.Name)
		if sw.URL != "" {
			log.Info("Opening %s for manual installation...", sw.URL)
			if err := sw.Install(); err != nil {
				log.Error("Failed to open URL: %v", err)
			}
			fmt.Printf("Press Enter after installing %s to continue...", sw.Name)
			_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
		} else {
			if err := sw.Install(); err != nil {
				log.Error("Failed to install %s: %v", sw.Name, err)
			} else {
				log.Success("%s installed successfully.", sw.Name)
			}
		}
	}
}

// setupSSH handles SSH key setup for GitHub access.
func setupSSH(verbose bool) error {
	log.Start("Setting up SSH keys for GitHub access")

	homeDir, err := userHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}

	sshKeyPath := filepath.Join(homeDir, ".ssh", "github")

	// Check if SSH key already exists
	if _, err := stat(sshKeyPath); err == nil {
		log.Success("SSH key already exists at ~/.ssh/github")
		if err := ensureSSHConfig(sshKeyPath); err != nil {
			return err
		}
		return validateGitHubSSHAuth(sshKeyPath, verbose)
	}

	// Try to setup from Bitwarden first
	log.Verbose(verbose, "Attempting to retrieve SSH key from Bitwarden...")
	needsManualGitHubRegistration := false
	sshConfigAlreadyEnsured := false
	if err := setupSSHFromBitwarden(sshKeyPath, verbose); err != nil {
		log.Warn("Could not retrieve SSH key from Bitwarden: %v", err)
		log.Message("Generating new SSH key...")

		// Generate new SSH key
		autoRegistered, err := generateSSHKey(sshKeyPath, verbose)
		if err != nil {
			return fmt.Errorf("failed to generate SSH key: %w", err)
		}
		needsManualGitHubRegistration = !autoRegistered

		log.Message("")
		log.Message("SSH key generated and configured successfully")
		log.Message("Tip: Store this key in Bitwarden manually for backup, or manage it with later phases")
	} else {
		sshConfigAlreadyEnsured = true
	}

	// Ensure SSH config is set up
	if !sshConfigAlreadyEnsured {
		if err := ensureSSHConfig(sshKeyPath); err != nil {
			return err
		}
	}

	if needsManualGitHubRegistration {
		if err := waitForManualGitHubKeyRegistration(sshKeyPath); err != nil {
			return err
		}
	}

	return validateGitHubSSHAuth(sshKeyPath, verbose)
}

// SetupSSHForGitHub exposes the SSH setup flow for other commands that need GitHub access.
func SetupSSHForGitHub(verbose bool) error {
	return setupSSH(verbose)
}

// generateSSHKey generates a new SSH key pair.
// Returns true when the key was automatically registered in GitHub.
func generateSSHKey(sshKeyPath string, verbose bool) (bool, error) {
	log.Start("Generating new SSH key pair...")

	// Ensure .ssh directory exists
	sshDir := filepath.Dir(sshKeyPath)
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		return false, fmt.Errorf("failed to create SSH directory: %w", err)
	}
	log.Verbose(verbose, "SSH directory ready at %s", sshDir)

	// Generate SSH key
	log.Verbose(verbose, "Generating SSH key with command: ssh-keygen -t ed25519 -f %s -N '' -C github", sshKeyPath)
	cmd := execCommand("ssh-keygen", "-t", "ed25519", "-f", sshKeyPath, "-N", "", "-C", "github")
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()

	log.Verbose(verbose, "Running command: %s", strings.Join(cmd.Args, " "))
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("failed to generate SSH key: %w", err)
	}

	log.Success("SSH key pair generated successfully")

	// Attempt to add the SSH key to GitHub
	if err := addSSHKeyToGitHub(sshKeyPath); err != nil {
		log.Warn("Could not automatically add SSH key to GitHub: %v", err)
		log.Message("")
		log.Message("Please manually add your SSH key to GitHub:")
		displaySSHKeyInstructions(sshKeyPath)
		return false, nil
	}

	return true, nil
}

func waitForManualGitHubKeyRegistration(sshKeyPath string) error {
	log.Message("After adding the key to GitHub, press Enter to continue...")
	_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')

	if _, err := os.Stat(sshKeyPath + ".pub"); err != nil {
		return fmt.Errorf("could not find public key at %s.pub: %w", sshKeyPath, err)
	}

	return nil
}

// addSSHKeyToGitHub attempts to add the SSH key to GitHub using the gh CLI.
func addSSHKeyToGitHub(sshKeyPath string) error {
	publicKeyPath := sshKeyPath + ".pub"

	// Check if gh CLI is available
	_, err := lookPath("gh")
	if err != nil {
		return fmt.Errorf("gh CLI not found - manual setup required")
	}

	// Check if gh is authenticated
	authCmd := execCommand("gh", "auth", "status")
	if err := authCmd.Run(); err != nil {
		return fmt.Errorf("gh CLI not authenticated")
	}

	// Read public key
	pubKeyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read public key: %w", err)
	}

	// Add SSH key to GitHub using gh CLI
	log.Start("Adding SSH key to GitHub...")
	addCmd := execCommand("gh", "ssh-key", "add", publicKeyPath, "--title", "eng-github")
	addCmd.Stdout = log.Writer()
	addCmd.Stderr = log.ErrorWriter()

	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("failed to add SSH key via gh CLI: %w", err)
	}

	log.Success("SSH key successfully added to GitHub")
	_ = pubKeyBytes
	return nil
}

// displaySSHKeyInstructions displays the SSH public key and GitHub setup instructions.
func displaySSHKeyInstructions(sshKeyPath string) {
	publicKeyPath := sshKeyPath + ".pub"

	pubKeyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		log.Error("Could not read public key: %v", err)
		return
	}

	log.Message("")
	log.Message("Your SSH public key:")
	log.Message("---------------------------------------------")
	log.Message("%s", string(pubKeyBytes))
	log.Message("---------------------------------------------")
	log.Message("")
	log.Message("Steps to add your SSH key to GitHub:")
	log.Message("1. Copy your public key (shown above)")
	log.Message("2. Go to: https://github.com/settings/keys")
	log.Message("3. Click 'New SSH key' and paste your key")
	log.Message("4. Give it a title (e.g., 'My Dev Machine')")
	log.Message("5. Click 'Add SSH key'")
	log.Message("")
	log.Message("Alternatively, if you have gh CLI installed and authenticated:")
	log.Message("  gh ssh-key add %s --title 'My Key Title'", publicKeyPath)
	log.Message("")
}

// validateGitHubSSHAuth verifies GitHub SSH authentication using the configured key.
func validateGitHubSSHAuth(sshKeyPath string, verbose bool) error {
	log.Start("Validating GitHub SSH authentication...")
	log.Verbose(verbose, "Validating with explicit key: %s", sshKeyPath)

	showConfig := execCommand("ssh", "-G", "github.com")
	configOut, configErr := showConfig.Output()
	if configErr == nil {
		for _, line := range strings.Split(string(configOut), "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "identityfile ") {
				log.Verbose(verbose, "ssh -G github.com -> %s", trimmed)
			}
		}
	} else {
		log.Verbose(verbose, "Could not inspect ssh config with `ssh -G`: %v", configErr)
	}

	cmd := execCommand(
		"ssh",
		"-T",
		"-o",
		"BatchMode=yes",
		"-o",
		"IdentitiesOnly=yes",
		"-o",
		"StrictHostKeyChecking=accept-new",
		"-i",
		sshKeyPath,
		"git@github.com",
	)
	output, err := cmd.CombinedOutput()
	outStr := strings.TrimSpace(string(output))
	if outStr != "" {
		log.Verbose(verbose, "GitHub SSH validation output: %s", outStr)
	}

	if strings.Contains(outStr, "successfully authenticated") || strings.Contains(outStr, "Hi ") {
		log.Success("GitHub SSH authentication validated with ~/.ssh/github")
		return nil
	}

	if err != nil {
		return fmt.Errorf("github SSH validation failed with ~/.ssh/github: %w", err)
	}

	return fmt.Errorf("github SSH validation failed with ~/.ssh/github")
}

// ensureSSHConfig ensures SSH config is set up for GitHub.
func ensureSSHConfig(sshKeyPath string) error {
	sshDir := filepath.Dir(sshKeyPath)
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
			if strings.Contains(configStr, "Host github.com") {
				log.Verbose(true, "GitHub SSH config entry already exists")
				return nil
			}
			// Append to existing config
			configContent = configStr + "\n" + configContent
		}
	}

	// Write SSH config
	if err := os.WriteFile(sshConfigPath, []byte(configContent), 0o600); err != nil {
		return fmt.Errorf("failed to write SSH config: %w", err)
	}

	log.Success("SSH config updated for GitHub access")
	return nil
}
