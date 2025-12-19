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

	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/log"
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
  - Check and install prerequisites (Homebrew, Git, Bash, SSH keys)
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

func init() {
	SetupCmd.AddCommand(SetupASDFCmd)
	SetupCmd.AddCommand(SetupDotfilesCmd)
	SetupCmd.AddCommand(SetupOhMyZshCmd)
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
