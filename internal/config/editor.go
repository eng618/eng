package config

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/editor"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui/theme"
)

// customEditorLabel offers free-form entry in the editor picker.
const customEditorLabel = "Custom (enter command...)"

// GitEditor checks the default editor configuration and prompts to set it.
// An optional argument sets the value directly without prompting.
func GitEditor(optionalEditor ...string) string {
	if len(optionalEditor) > 0 && strings.TrimSpace(optionalEditor[0]) != "" {
		SetGitEditor(optionalEditor[0])
		return viper.GetString("git.editor")
	}

	log.Start("Checking for default editor in config")

	current := strings.TrimSpace(viper.GetString("git.editor"))
	if current == "" {
		return UpdateGitEditor()
	}

	confirmed, err := ConfirmPrompt(
		fmt.Sprintf("Confirm default editor: %s?", theme.PrimaryText.Render(current)),
		true,
	)
	cobra.CheckErr(err)

	if !confirmed {
		return UpdateGitEditor()
	}

	log.Success("Confirmed default editor")
	return current
}

// UpdateGitEditor prompts to pick from PATH-detected editors and persists the choice.
func UpdateGitEditor() string {
	current := strings.TrimSpace(viper.GetString("git.editor"))
	available := editor.Available()

	// Preselect the current value when it matches a detected editor.
	defaultName := ""
	if current != "" {
		for _, opt := range available {
			if strings.EqualFold(opt.Command, current) || strings.EqualFold(opt.Name, current) {
				defaultName = opt.Name
				break
			}
		}
	}
	if defaultName == "" && len(available) > 0 {
		defaultName = available[0].Name
	}

	options := make([]string, 0, len(available)+1)
	for _, opt := range available {
		options = append(options, opt.Name)
	}
	options = append(options, customEditorLabel)
	if defaultName == "" {
		defaultName = customEditorLabel
	}

	picked, err := SelectPrompt("Select default editor:", options, defaultName)
	cobra.CheckErr(err)

	if picked == customEditorLabel {
		command, err := InputPrompt("Enter editor command (e.g. agy-ide, code, nvim):", current)
		cobra.CheckErr(err)
		if strings.TrimSpace(command) == "" {
			command = editor.DefaultCommand(current)
		}
		SetGitEditor(command)
		return viper.GetString("git.editor")
	}

	if opt, ok := editor.FindByName(available, picked); ok {
		SetGitEditor(opt.Command)
		return viper.GetString("git.editor")
	}

	// Fallback: persist the raw pick (e.g. mocked prompts in tests).
	SetGitEditor(picked)
	return viper.GetString("git.editor")
}

// SetGitEditor persists the default editor command in viper and writes to config.
func SetGitEditor(command string) {
	viper.Set("git.editor", strings.TrimSpace(command))

	if err := viper.WriteConfig(); err != nil {
		cobra.CheckErr(
			fmt.Errorf(
				"%s: %w",
				lipgloss.NewStyle().Foreground(theme.Destructive).Render("Error writing config file"),
				err,
			),
		)
	}
	log.Success("Default editor updated successfully")
}
