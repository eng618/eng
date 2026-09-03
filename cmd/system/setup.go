package system

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

var SetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup development tools",
	Long: `Setup various development tools.
Running this command without subcommands will run all setup steps:
- Oh My Zsh
- ASDF plugins
- Dotfiles installation
- Dotfiles secrets restore (when configured)
- Software installation
- GPG keys setup (interactive)
- GPG permissions fix`,
	RunE: func(cmd *cobra.Command, _args []string) error {
		if err := runSetup(cmd, cmdutil.IsVerbose(cmd)); err != nil {
			return fmt.Errorf("setup failed: %w", err)
		}
		return nil
	},
}

var (
	ensurePrerequisitesStep = EnsurePrerequisites
	setupOhMyZshStep        = setupOhMyZsh
	setupASDFStep           = setupASDF
	setupDotfilesStep       = setupDotfiles
	setupSoftwareStep       = setupSoftware
	setupGPGStep            = setupGPG
	setupGPGPermissionsStep = setupGPGPermissions

	runSetupWizard = func(steps []setupStep) ([]string, error) {
		var groups []*huh.Group
		stepActions := make([]string, len(steps))

		for i, step := range steps {
			stepActions[i] = setupActionContinue // default
			groups = append(groups, huh.NewGroup(
				huh.NewSelect[string]().
					Title("Step: "+step.Name).
					Description(step.Purpose).
					Options(
						huh.NewOption(setupActionContinue, setupActionContinue),
						huh.NewOption(setupActionSkip, setupActionSkip),
						huh.NewOption(setupActionExit, setupActionExit),
					).
					Value(&stepActions[i]),
			))
		}

		form := huh.NewForm(groups...).WithTheme(theme.EngTheme())
		if err := form.Run(); err != nil {
			return nil, err
		}
		return stepActions, nil
	}
)

type setupStep struct {
	Name    string
	Purpose string
	Run     func() error
}

const (
	setupActionContinue = "Continue"
	setupActionSkip     = "Skip"
	setupActionExit     = "Exit"
)

var SetupASDFCmd = &cobra.Command{
	Use:   "asdf",
	Short: "Setup asdf plugins from $HOME/.tool-versions",
	Long:  `Reads $HOME/.tool-versions and installs asdf plugins listed there.`,
	Run: func(cmd *cobra.Command, args []string) {
		setupASDF(cmdutil.IsVerbose(cmd))
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
	- Configure git to hide untracked files
	- Restore dotfiles secrets when manifest and BWS token are available`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := setupDotfiles(cmdutil.IsVerbose(cmd)); err != nil {
			return fmt.Errorf("dotfiles setup failed: %w", err)
		}
		return nil
	},
}

var SetupOhMyZshCmd = &cobra.Command{
	Use:   "oh-my-zsh",
	Short: "Install Oh My Zsh",
	Long:  `Downloads and installs Oh My Zsh. Skips if already installed.`,
	Run: func(cmd *cobra.Command, args []string) {
		setupOhMyZsh(cmdutil.IsVerbose(cmd))
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
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := SetupSSH(cmdutil.IsVerbose(cmd)); err != nil {
			return fmt.Errorf("ssh setup failed: %w", err)
		}
		return nil
	},
}

func init() {
	SetupCmd.AddCommand(SetupASDFCmd)
	SetupCmd.AddCommand(SetupDotfilesCmd)
	SetupCmd.AddCommand(SetupOhMyZshCmd)
	SetupCmd.AddCommand(SetupSSHCmd)
	SetupCmd.AddCommand(SetupGPGCmd)
	SetupCmd.Flags().BoolP("interactive", "i", false, "Prompt before each setup step with continue/skip/exit options")
}

func runSetup(cmd *cobra.Command, verbose bool) error {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary).
		MarginBottom(1)
	if !ui.DisableProgress {
		fmt.Fprintln(log.Out, headerStyle.Render("🚀 System Environment Setup"))
	}

	interactive, err := cmd.Flags().GetBool("interactive")
	if err != nil {
		return fmt.Errorf("failed to read interactive flag: %w", err)
	}

	steps := []setupStep{
		{
			Name:    "Prerequisites",
			Purpose: "Verify required tools are installed before setup runs.",
			Run: func() error {
				if err := ensurePrerequisitesStep(verbose); err != nil {
					return fmt.Errorf("prerequisites check failed: %w", err)
				}
				return nil
			},
		},
		{
			Name:    "Oh My Zsh",
			Purpose: "Install or verify Oh My Zsh for shell configuration.",
			Run: func() error {
				setupOhMyZshStep(verbose)
				return nil
			},
		},
		{
			Name:    "ASDF Plugins",
			Purpose: "Install ASDF plugins and tool versions from $HOME/.tool-versions.",
			Run: func() error {
				setupASDFStep(verbose)
				return nil
			},
		},
		{
			Name:    "Dotfiles",
			Purpose: "Install dotfiles and restore managed secrets when available.",
			Run: func() error {
				if err := setupDotfilesStep(verbose); err != nil {
					log.Error("Dotfiles setup failed: %v", err)
				}
				return nil
			},
		},
		{
			Name:    "Software",
			Purpose: "Install required software and open download links for optional apps.",
			Run: func() error {
				setupSoftwareStep(verbose)
				return nil
			},
		},
		{
			Name:    "GPG Keys",
			Purpose: "Setup GPG keys for signing commits and encryption (interactive).",
			Run: func() error {
				if err := setupGPGStep(verbose); err != nil {
					log.Error("GPG setup failed: %v", err)
				}
				return nil
			},
		},
		{
			Name:    "GPG Permissions",
			Purpose: "Fix GPG directory permissions to prevent warnings.",
			Run: func() error {
				setupGPGPermissionsStep(verbose)
				return nil
			},
		},
	}

	if interactive {
		stepActions, err := runSetupWizard(steps)
		if err != nil {
			log.Info("Setup wizard canceled.")
			return nil
		}

		// Execute based on choices
		for i, action := range stepActions {
			switch action {
			case setupActionSkip:
				log.Info("Skipping setup step: %s", steps[i].Name)
				continue
			case setupActionExit:
				log.Info("Setup exited early at step: %s", steps[i].Name)
				return nil
			case setupActionContinue:
				if err := steps[i].Run(); err != nil {
					return err
				}
			}
		}
	} else {
		// Non-interactive execution
		for _, step := range steps {
			if err := step.Run(); err != nil {
				return err
			}
		}
	}

	theme.SuccessMessage("System environment setup completed successfully!")
	return nil
}

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

// ensureASDFInstalled verifies that asdf is installed, installing it via Homebrew if available or via git clone.
func ensureASDFInstalled(verbose bool) error {
	if _, err := lookPath("asdf"); err == nil {
		log.Verbose(verbose, "asdf is already installed")
		return nil
	}

	// Check if Homebrew is available
	_, brewErr := lookPath("brew")
	if brewErr != nil {
		if _, err := stat("/home/linuxbrew/.linuxbrew/bin/brew"); err == nil {
			_ = os.Setenv("PATH", os.Getenv("PATH")+":/home/linuxbrew/.linuxbrew/bin")
			brewErr = nil
		}
	}

	if brewErr == nil {
		log.Start("Installing asdf via Homebrew...")
		cmd := execCommand("brew", "install", "asdf")
		cmd.Stdout = log.Writer()
		cmd.Stderr = log.ErrorWriter()
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install asdf via Homebrew: %w", err)
		}

		// Update PATH if brew installed asdf to a specific prefix
		if _, err := lookPath("asdf"); err != nil {
			cmdPrefix := execCommand("brew", "--prefix", "asdf")
			if out, err := cmdPrefix.Output(); err == nil {
				asdfPrefix := strings.TrimSpace(string(out))
				asdfBin := filepath.Join(asdfPrefix, "bin")
				asdfLibexecBin := filepath.Join(asdfPrefix, "libexec", "bin")
				if _, err := stat(asdfBin); err == nil {
					_ = os.Setenv("PATH", os.Getenv("PATH")+":"+asdfBin)
				} else if _, err := stat(asdfLibexecBin); err == nil {
					_ = os.Setenv("PATH", os.Getenv("PATH")+":"+asdfLibexecBin)
				}
			}
		}

		log.Success("asdf installed successfully via Homebrew")
		return nil
	}

	// Check for existing ~/.asdf
	homeDir, err := userHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}

	asdfDir := filepath.Join(homeDir, ".asdf")
	if _, err := stat(asdfDir); err == nil {
		_ = os.Setenv("PATH", os.Getenv("PATH")+":"+filepath.Join(asdfDir, "bin"))
		log.Verbose(verbose, "Found existing asdf directory at %s", asdfDir)
		return nil
	}

	if _, err := lookPath("git"); err != nil {
		return fmt.Errorf("neither Homebrew nor Git is available to install asdf")
	}

	log.Start("Installing asdf via Git clone into ~/.asdf...")
	cmd := execCommand("git", "clone", "https://github.com/asdf-vm/asdf.git", asdfDir, "--branch", "v0.14.0")
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to git clone asdf: %w", err)
	}

	_ = os.Setenv("PATH", os.Getenv("PATH")+":"+filepath.Join(asdfDir, "bin"))
	log.Success("asdf installed via Git to ~/.asdf")
	return nil
}

func setupASDF(verbose bool) {
	log.Verbose(verbose, "Starting ASDF setup...")

	if err := ensureASDFInstalled(verbose); err != nil {
		log.Error("Could not setup asdf: %v", err)
		return
	}

	homeDir, err := userHomeDir()
	if err != nil {
		log.Error("Could not determine home directory: %v", err)
		return
	}

	toolVersionsPath := filepath.Join(homeDir, ".tool-versions")
	if _, err := stat(toolVersionsPath); err != nil {
		if _, currErr := stat(".tool-versions"); currErr == nil {
			toolVersionsPath = ".tool-versions"
		} else {
			log.Message("No .tool-versions file found in %s; skipping asdf plugin setup", homeDir)
			return
		}
	}

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
		out, err := cmd.CombinedOutput()
		outStr := strings.TrimSpace(string(out))
		if err != nil {
			if strings.Contains(strings.ToLower(outStr), "already") {
				log.Verbose(verbose, "asdf plugin '%s' already added", plugin)
			} else {
				log.Error("Failed to add asdf plugin '%s': %v (%s)", plugin, err, outStr)
			}
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

	// Ensure Zsh is installed before running Oh My Zsh installer
	if err := ensureZsh(verbose); err != nil {
		log.Error("Cannot install Oh My Zsh: %v", err)
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

	if err := maybeRestoreDotfilesSecrets(exe, verbose); err != nil {
		return err
	}

	return nil
}

func maybeRestoreDotfilesSecrets(exe string, verbose bool) error {
	manifestPath := filepath.Join(resolveDotfilesWorktreePath(), "bin", "secrets", "server.manifest")
	if _, err := stat(manifestPath); err != nil {
		log.Verbose(verbose, "Skipping dotfiles secrets restore, manifest not found: %s", manifestPath)
		return nil
	}

	if strings.TrimSpace(os.Getenv("BWS_ACCESS_TOKEN")) == "" {
		log.Warn("Skipping dotfiles secrets restore: BWS_ACCESS_TOKEN is not set")
		log.Message("Run manually after exporting BWS_ACCESS_TOKEN: eng dotfiles secrets restore")
		return nil
	}

	log.Start("Restoring dotfiles secrets...")
	args := []string{"dotfiles", "secrets", "restore", "--manifest", manifestPath}
	if verbose {
		args = append(args, "-v")
	}

	cmd := execCommand(exe, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("dotfiles secrets restore failed: %w", err)
	}

	return nil
}

func resolveDotfilesWorktreePath() string {
	worktreePath := strings.TrimSpace(viper.GetString("dotfiles.worktree_path"))
	if worktreePath == "" {
		worktreePath = strings.TrimSpace(os.Getenv("HOME"))
	}
	if worktreePath == "" {
		return "."
	}
	return os.ExpandEnv(worktreePath)
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
		selected, err := ui.MultiSelect("Select additional software to install:", optionalOptions, nil)
		if err != nil {
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
			fmt.Fprintf(log.Out, "Press Enter after installing %s to continue...", sw.Name)
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

// SetupSSHForGitHub exposes the SSH setup flow for other commands that need GitHub access.
func SetupSSHForGitHub(verbose bool) error {
	return SetupSSH(verbose)
}
