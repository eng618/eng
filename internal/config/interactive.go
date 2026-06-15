package config

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/ui/theme"
)

// RunInteractiveEditor launches a TUI to edit the CLI's configuration.
func RunInteractiveEditor() error {
	// 1. Load current values
	email := viper.GetString("email")
	verbose := viper.GetBool("verbose")

	gitDevPath := viper.GetString("git.dev_path")
	if gitDevPath == "" {
		home, _ := os.UserHomeDir()
		gitDevPath = home + "/Development"
	}

	dfRepoURL := viper.GetString("dotfiles.repo_url")
	dfBranch := viper.GetString("dotfiles.branch")
	if dfBranch == "" {
		dfBranch = "main"
	}
	dfBare := viper.GetString("dotfiles.bare_repo_path")
	dfWorktree := viper.GetString("dotfiles.worktree_path")

	glHost := viper.GetString("gitlab.host")
	if glHost == "" {
		glHost = "gitlab.com"
	}
	glProject := viper.GetString("gitlab.project")

	glTokenMethod := "Raw Token"
	if viper.GetString("gitlab.tokenItem") != "" {
		glTokenMethod = "Bitwarden Item"
	}
	glTokenItem := viper.GetString("gitlab.tokenItem")
	glToken := viper.GetString("gitlab.token")

	// 2. Build the form
	form := huh.NewForm(
		// Group 1: Global Settings
		huh.NewGroup(
			huh.NewInput().
				Title("Email Address").
				Description("Primary email for git and other configs.").
				Value(&email),
			huh.NewInput().
				Title("Git Dev Path").
				Description("Directory where your repositories live.").
				Value(&gitDevPath),
			huh.NewConfirm().
				Title("Enable Verbose Output").
				Description("Show detailed debug logs by default.").
				Value(&verbose),
		),

		// Group 2: Dotfiles Settings
		huh.NewGroup(
			huh.NewInput().
				Title("Dotfiles Repo URL").
				Description("Git URL for your bare dotfiles repository.").
				Value(&dfRepoURL),
			huh.NewInput().
				Title("Dotfiles Branch").
				Value(&dfBranch),
			huh.NewInput().
				Title("Bare Repo Path").
				Description("Where the bare git directory lives (e.g., ~/.dotfiles).").
				Value(&dfBare),
			huh.NewInput().
				Title("Worktree Path").
				Description("Where the dotfiles are checked out (usually $HOME).").
				Value(&dfWorktree),
		),

		// Group 3: GitLab Settings
		huh.NewGroup(
			huh.NewInput().
				Title("GitLab Host").
				Description("e.g. gitlab.com or your self-hosted domain.").
				Value(&glHost),
			huh.NewInput().
				Title("Default Project").
				Description("Your default namespace/project-name.").
				Value(&glProject),
			huh.NewSelect[string]().
				Title("GitLab Token Method").
				Options(
					huh.NewOption("Bitwarden Item", "Bitwarden Item"),
					huh.NewOption("Raw Token", "Raw Token"),
					huh.NewOption("None", "None"),
				).
				Value(&glTokenMethod),
		),

		// Conditional Group: GitLab Token Details
		huh.NewGroup(
			huh.NewInput().
				Title("Bitwarden Item ID/Name").
				Description("The Bitwarden item containing your GitLab token.").
				Value(&glTokenItem),
		).WithHideFunc(func() bool {
			return glTokenMethod != "Bitwarden Item"
		}),

		huh.NewGroup(
			huh.NewInput().
				Title("GitLab Personal Access Token").
				EchoMode(huh.EchoModePassword).
				Value(&glToken),
		).WithHideFunc(func() bool {
			return glTokenMethod != "Raw Token"
		}),
	).WithTheme(theme.EngTheme())

	// 3. Run the form
	err := form.Run()
	if err != nil {
		if err == huh.ErrUserAborted {
			return fmt.Errorf("configuration edit aborted")
		}
		return err
	}

	// 4. Save the results back to viper
	viper.Set("email", email)
	viper.Set("verbose", verbose)
	viper.Set("git.dev_path", gitDevPath)

	viper.Set("dotfiles.repo_url", dfRepoURL)
	viper.Set("dotfiles.branch", dfBranch)
	viper.Set("dotfiles.bare_repo_path", dfBare)
	viper.Set("dotfiles.worktree_path", dfWorktree)

	viper.Set("gitlab.host", glHost)
	viper.Set("gitlab.project", glProject)

	if glTokenMethod == "Bitwarden Item" {
		viper.Set("gitlab.tokenItem", glTokenItem)
		viper.Set("gitlab.token", "")
	} else if glTokenMethod == "Raw Token" {
		viper.Set("gitlab.tokenItem", "")
		viper.Set("gitlab.token", glToken)
	} else {
		viper.Set("gitlab.tokenItem", "")
		viper.Set("gitlab.token", "")
	}

	// 5. Write to config file
	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	theme.SuccessMessage(fmt.Sprintf("Configuration saved to %s", viper.ConfigFileUsed()))
	return nil
}
