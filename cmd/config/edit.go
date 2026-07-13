package config

import (
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/config"
)

var interactiveMode bool

// EditCmd represents the edit command.
var EditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Interactively edit the eng CLI configuration",
	Long: `Launch a beautiful TUI wizard to easily update your .eng.yaml configuration settings.

By default, this command launches the interactive editor. Use --interactive=false if you want to bypass the wizard in the future (though currently, the interactive editor is the primary function of this command).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !interactiveMode {
			// In the future, if interactive mode is false, we might open $EDITOR
			// For now, default to the TUI if they just run 'eng config edit'
			return config.RunInteractiveEditor()
		}
		return config.RunInteractiveEditor()
	},
}

func init() {
	EditCmd.Flags().BoolVarP(&interactiveMode, "interactive", "i", true, "Launch the interactive TUI editor")
}
