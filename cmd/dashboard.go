package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/dashboard"
)

// dashboardCmd represents the dashboard command.
var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Launch the interactive Project & Git Dashboard",
	Long: `Opens a full-screen "mission control" interface to view all configured projects.
It displays your projects in a list, and shows the live status of their repositories
(cloned state, current branch, and uncommitted changes) in real-time.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return dashboard.Run()
	},
}

// selectEditorCmd represents the hidden command for selecting an editor inside the dashboard.
var selectEditorCmd = &cobra.Command{
	Use:    "select-editor [path]",
	Short:  "Select an editor to open the target path",
	Hidden: true,
	Args:   cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetPath := args[0]

		// 1. Get default configured editor
		defaultEditor := config.GetGitConfig().Editor
		if defaultEditor == "" {
			defaultEditor = os.Getenv("VISUAL")
			if defaultEditor == "" {
				defaultEditor = os.Getenv("EDITOR")
			}
		}

		type EditorOption struct {
			Name    string
			Command string
			IsApp   bool
		}

		potentialEditors := []EditorOption{
			{Name: "Antigravity IDE", Command: "Antigravity IDE", IsApp: true},
			{Name: "Antigravity VS Code", Command: "Antigravity", IsApp: true},
			{Name: "Visual Studio Code (CLI)", Command: "code", IsApp: false},
			{Name: "Visual Studio Code (App)", Command: "Visual Studio Code", IsApp: true},
			{Name: "Neovim", Command: "nvim", IsApp: false},
			{Name: "Vim", Command: "vim", IsApp: false},
			{Name: "Nano", Command: "nano", IsApp: false},
			{Name: "Emacs", Command: "emacs", IsApp: false},
			{Name: "Sublime Text (CLI)", Command: "subl", IsApp: false},
			{Name: "Sublime Text (App)", Command: "Sublime Text", IsApp: true},
			{Name: "Cursor (CLI)", Command: "cursor", IsApp: false},
			{Name: "Cursor (App)", Command: "Cursor", IsApp: true},
			{Name: "Xcode", Command: "Xcode", IsApp: true},
			{Name: "Android Studio", Command: "Android Studio", IsApp: true},
		}

		var available []EditorOption
		var options []string

		// Add default editor if configured
		if defaultEditor != "" {
			available = append(available, EditorOption{
				Name:    fmt.Sprintf("Configured Default (%s)", defaultEditor),
				Command: defaultEditor,
				IsApp:   false,
			})
			options = append(options, available[0].Name)
		}

		for _, opt := range potentialEditors {
			// Avoid duplicate default
			if defaultEditor != "" &&
				(strings.Contains(strings.ToLower(opt.Name), strings.ToLower(defaultEditor)) || strings.Contains(strings.ToLower(defaultEditor), strings.ToLower(opt.Command))) {
				continue
			}

			if opt.IsApp {
				appPath1 := "/Applications/" + opt.Command + ".app"
				home, _ := os.UserHomeDir()
				appPath2 := filepath.Join(home, "Applications", opt.Command+".app")

				if _, err := os.Stat(appPath1); err == nil {
					available = append(available, opt)
					options = append(options, opt.Name)
				} else if _, err := os.Stat(appPath2); err == nil {
					available = append(available, opt)
					options = append(options, opt.Name)
				}
			} else {
				if _, err := exec.LookPath(opt.Command); err == nil {
					available = append(available, opt)
					options = append(options, opt.Name)
				}
			}
		}

		if len(available) == 0 {
			// Fallback to vim/nano
			available = append(available, EditorOption{Name: "Vim", Command: "vim", IsApp: false})
			available = append(available, EditorOption{Name: "Nano", Command: "nano", IsApp: false})
			options = append(options, "Vim", "Nano")
		}

		selectedName, err := ui.Select("Select editor to open:", options, options[0])
		if err != nil {
			return err
		}

		var chosen EditorOption
		for _, opt := range available {
			if opt.Name == selectedName {
				chosen = opt
				break
			}
		}

		var execCmd *exec.Cmd
		if chosen.IsApp {
			execCmd = exec.Command("open", "-a", chosen.Command, targetPath)
		} else {
			parts := strings.Fields(chosen.Command)
			execCmd = exec.Command(parts[0], append(parts[1:], targetPath)...)
		}

		execCmd.Stdin = os.Stdin
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr

		return execCmd.Run()
	},
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
	dashboardCmd.AddCommand(selectEditorCmd)
}
