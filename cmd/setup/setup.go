package setup

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

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
	setupGPGStep            = GPGSetup
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

// NOTE (hard break): SSH setup lives under `eng ssh setup` (cmd/ssh) and
// GPG setup under `eng gpg setup` (cmd/gpg). `eng setup` runs all steps
// in sequence via the injected GPGSetup hook wired in cmd/root.go.

func init() {
	SetupCmd.AddCommand(SetupASDFCmd)
	SetupCmd.AddCommand(SetupDotfilesCmd)
	SetupCmd.AddCommand(SetupOhMyZshCmd)
	SetupCmd.AddCommand(CompauditFixCmd)
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
