package system

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/eng618/eng/utils/log"
	"github.com/spf13/cobra"
)

var SetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup development tools",
	Long: `Setup various development tools.
Running this command without subcommands will run all setup steps:
- Oh My Zsh
- ASDF plugins
- Dotfiles installation`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := EnsurePrerequisites(); err != nil {
			log.Fatal("Prerequisites check failed: %v", err)
		}
		setupOhMyZsh()
		setupASDF()
		if err := setupDotfiles(); err != nil {
			log.Error("Dotfiles setup failed: %v", err)
		}
	},
}

var SetupASDFCmd = &cobra.Command{
	Use:   "asdf",
	Short: "Setup asdf plugins from $HOME/.tool-versions",
	Long:  `Reads $HOME/.tool-versions and installs asdf plugins listed there.`,
	Run: func(cmd *cobra.Command, args []string) {
		setupASDF()
	},
}

var SetupDotfilesCmd = &cobra.Command{
	Use:   "dotfiles",
	Short: "Setup dotfiles from your git repository",
	Long: `Setup dotfiles from your git repository. This command will:
  - Check and install prerequisites (Homebrew, Git, Bash, SSH keys)
  - Clone your dotfiles repository as a bare repository
  - Backup any conflicting files
  - Checkout dotfiles to your home directory
  - Initialize git submodules
  - Configure git to hide untracked files`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := setupDotfiles(); err != nil {
			log.Fatal("Dotfiles setup failed: %v", err)
		}
	},
}

var SetupOhMyZshCmd = &cobra.Command{
	Use:   "oh-my-zsh",
	Short: "Install Oh My Zsh",
	Long:  `Downloads and installs Oh My Zsh. Skips if already installed.`,
	Run: func(cmd *cobra.Command, args []string) {
		setupOhMyZsh()
	},
}

func init() {
	SetupCmd.AddCommand(SetupASDFCmd)
	SetupCmd.AddCommand(SetupDotfilesCmd)
	SetupCmd.AddCommand(SetupOhMyZshCmd)
}

var execCommand = exec.Command

func setupASDF() {
	// Check error for os.UserHomeDir
	homeDir, err := os.UserHomeDir()
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

func setupOhMyZsh() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Error("Could not determine home directory: %v", err)
		return
	}
	omzPath := filepath.Join(homeDir, ".oh-my-zsh")
	if _, err := os.Stat(omzPath); err == nil {
		log.Success("Oh My Zsh is already installed")
		return
	}

	log.Start("Installing Oh My Zsh...")
	// Use --unattended to prevent switching shell immediately
	cmd := execCommand("sh", "-c", "curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh | sh -s -- --unattended")
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()
	if err := cmd.Run(); err != nil {
		log.Error("Failed to install Oh My Zsh: %v", err)
	} else {
		log.Success("Oh My Zsh installed successfully")
	}
}

// setupDotfiles sets up dotfiles by checking prerequisites and running the install command.
func setupDotfiles() error {
	log.Start("Checking dotfiles prerequisites")

	// Check prerequisites
	if err := EnsurePrerequisites(); err != nil {
		return err
	}

	// Get the path to the current executable
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	log.Start("Running dotfiles install...")
	// Run dependencies install command
	cmd := execCommand(exe, "dotfiles", "install")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("dotfiles install failed: %w", err)
	}

	return nil
}
