package system

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

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

	if distro.IsFedora() {
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
	} else if _, err := lookPath("brew"); err == nil {
		log.Start("Installing Git via Homebrew")
		cmd := execCommand("brew", "install", "git")
		cmd.Stdout = log.Writer()
		cmd.Stderr = log.ErrorWriter()
		if err := cmd.Run(); err != nil {
			log.Error("Failed to install Git via Homebrew: %v", err)
			return fmt.Errorf("git installation failed: %w", err)
		}
	} else {
		return fmt.Errorf("git is not installed and no supported package manager (dnf, apt, brew) was found")
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

	if distro.IsFedora() {
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
	} else if _, err := lookPath("brew"); err == nil {
		log.Start("Installing Bash via Homebrew")
		cmd := execCommand("brew", "install", "bash")
		cmd.Stdout = log.Writer()
		cmd.Stderr = log.ErrorWriter()
		if err := cmd.Run(); err != nil {
			log.Error("Failed to install Bash via Homebrew: %v", err)
			return fmt.Errorf("bash installation failed: %w", err)
		}
	} else {
		return fmt.Errorf("bash is not installed and no supported package manager (dnf, apt, brew) was found")
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
		if strings.Contains(err.Error(), "user aborted") {
			return fmt.Errorf("zsh installation aborted: %w", err)
		}
		// Non-interactive / headless environment (e.g. CI): proceed with default (true)
		log.Verbose(verbose, "Non-interactive confirmation fallback: %v", err)
		confirm = true
	}
	if !confirm {
		return fmt.Errorf("zsh installation declined - zsh is required for shell and dotfiles setup")
	}

	distro := detectDistro()

	if distro.IsFedora() {
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
	} else if _, err := lookPath("brew"); err == nil {
		log.Start("Installing Zsh via Homebrew")
		cmd := execCommand("brew", "install", "zsh")
		cmd.Stdout = log.Writer()
		cmd.Stderr = log.ErrorWriter()
		if err := cmd.Run(); err != nil {
			log.Error("Failed to install Zsh via Homebrew: %v", err)
			return fmt.Errorf("zsh installation failed: %w", err)
		}
	} else {
		return fmt.Errorf("zsh is not installed and no supported package manager (dnf, apt, brew) was found")
	}

	log.Success("Zsh installed successfully")
	return nil
}
