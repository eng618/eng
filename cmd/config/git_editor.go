package config

import (
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
)

// GitEditorCmd defines the command for setting the default editor.
var GitEditorCmd = &cobra.Command{
	Use:     "git-editor [editor]",
	Aliases: []string{"editor"},
	Short:   "Update config default editor",
	Long: `Select the default editor used by the dashboard (e) from installed editors detected on PATH.

You may also pass the editor command directly. The dashboard falls back to $VISUAL/$EDITOR,
then agy-ide, code, and nano when no default is configured.`,
	Example: `  eng config git-editor
  eng config git-editor agy-ide
  eng config editor code`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Checking for default editor in config file...")
		if len(args) > 0 {
			config.GitEditor(args[0])
		} else {
			config.GitEditor()
		}
	},
}
