package system

import (
	"bufio"
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
	Long:  `Setup various development tools. Example: eng system setup asdf`,
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
  - Check and install prerequisites (Homebrew, Git, Bash, GitHub CLI, SSH keys)
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

func init() {
	SetupCmd.AddCommand(SetupASDFCmd)
	SetupCmd.AddCommand(SetupDotfilesCmd)
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

// setupDotfiles sets up dotfiles by checking prerequisites and installing from the configured repository.
func setupDotfiles() error {
	log.Start("Starting dotfiles setup")

	// Check prerequisites
	if err := EnsurePrerequisites(); err != nil {
		return err
	}

	// Import dotfiles package for install functionality
	// Note: This is handled by calling the install workflow directly
	// through the dotfiles.InstallCmd which is registered in cmd/dotfiles/dotfiles.go
	log.Success("Prerequisites satisfied, proceeding with dotfiles installation")
	log.Message("")
	log.Message("Please run: eng dotfiles install")
	log.Message("This will guide you through the dotfiles installation process")
	
	return nil
}
