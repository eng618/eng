package setup

import (
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/log"
)

var SetupOhMyZshCmd = &cobra.Command{
	Use:   "oh-my-zsh",
	Short: "Install Oh My Zsh",
	Long:  `Downloads and installs Oh My Zsh. Skips if already installed.`,
	Run: func(cmd *cobra.Command, args []string) {
		setupOhMyZsh(cmdutil.IsVerbose(cmd))
	},
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
