package dotfiles

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/secrets"
)

// EditCmd opens a dotfile (or its template) in the default editor and re-renders if needed.
var EditCmd = &cobra.Command{
	Use:   "edit <file>",
	Short: "Edit a dotfile or its underlying template",
	Long:  `Opens the underlying template or dotfile in your editor and re-renders it securely if it's a template.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetFile := args[0]
		_, worktreePath, err := getDotfilesConfig()
		if err != nil {
			return err
		}

		// Check if a .tmpl exists for this file
		tmplPath := filepath.Join(worktreePath, targetFile+".tmpl")
		editPath := filepath.Join(worktreePath, targetFile)
		isTmpl := false

		if _, err := os.Stat(tmplPath); err == nil {
			editPath = tmplPath
			isTmpl = true
			log.Info("Found template for %s, editing template directly.", targetFile)
		} else if strings.HasSuffix(targetFile, ".tmpl") {
			isTmpl = true
			log.Info("Editing template directly.")
		} else {
			log.Info("Editing %s directly (no template found).", targetFile)
		}

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}

		editorCmd := exec.Command(editor, editPath)
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = log.Writer()
		editorCmd.Stderr = log.ErrorWriter()

		if err := editorCmd.Run(); err != nil {
			return fmt.Errorf("editor %s failed: %w", editor, err)
		}

		if isTmpl {
			log.Info("Re-rendering template...")
			// We render all templates, or we could render just one. 
			// For simplicity and safety, we trigger a full render pass.
			if err := secrets.RenderTemplates(worktreePath, "", cmdutil.IsVerbose(cmd), true); err != nil {
				return fmt.Errorf("failed to render templates after editing: %w", err)
			}
			log.Success("Template re-rendered successfully.")
		}

		return nil
	},
}

func init() {
	DotfilesCmd.AddCommand(EditCmd)
}
