package editor

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/eng618/eng/internal/execx"
	"github.com/eng618/eng/internal/paths"
)

// Option describes a single editor candidate.
type Option struct {
	Name    string
	Command string
	IsApp   bool
}

// Potential lists known editors in `ide()` precedence order.
var Potential = []Option{
	{Name: "agy-ide (CLI)", Command: "agy-ide", IsApp: false},
	{Name: "antigravity-ide (CLI)", Command: "antigravity-ide", IsApp: false},
	{Name: "Antigravity IDE", Command: "Antigravity IDE", IsApp: true},
	{Name: "Antigravity VS Code", Command: "Antigravity", IsApp: true},
	{Name: "Visual Studio Code (CLI)", Command: "code", IsApp: false},
	{Name: "Visual Studio Code (App)", Command: "Visual Studio Code", IsApp: true},
	{Name: "Cursor (CLI)", Command: "cursor", IsApp: false},
	{Name: "Cursor (App)", Command: "Cursor", IsApp: true},
	{Name: "Neovim", Command: "nvim", IsApp: false},
	{Name: "Vim", Command: "vim", IsApp: false},
	{Name: "Nano", Command: "nano", IsApp: false},
	{Name: "Emacs", Command: "emacs", IsApp: false},
	{Name: "Sublime Text (CLI)", Command: "subl", IsApp: false},
	{Name: "Sublime Text (App)", Command: "Sublime Text", IsApp: true},
	{Name: "Xcode", Command: "Xcode", IsApp: true},
	{Name: "Android Studio", Command: "Android Studio", IsApp: true},
}

// Mockable entry points for tests.
var (
	// LookPath resolves a CLI binary on PATH.
	LookPath = execx.LookPath
	// Stat checks for macOS .app bundles.
	Stat = os.Stat
)

// Available returns the subset of Potential installed on this machine,
// preserving precedence order. CLI entries use PATH lookup; app entries
// check /Applications and ~/Applications.
func Available() []Option {
	var out []Option
	home := paths.MustHome()
	for _, opt := range Potential {
		if opt.IsApp {
			if _, err := Stat(filepath.Join("/Applications", opt.Command+".app")); err == nil {
				out = append(out, opt)
				continue
			}
			if home != "" {
				if _, err := Stat(filepath.Join(home, "Applications", opt.Command+".app")); err == nil {
					out = append(out, opt)
				}
			}
			continue
		}
		if _, err := LookPath(opt.Command); err == nil {
			out = append(out, opt)
		}
	}
	return out
}

// Commands returns the executable command for each available editor.
func Commands() []string {
	available := Available()
	out := make([]string, len(available))
	for i, opt := range available {
		out[i] = opt.Command
	}
	return out
}

// DefaultCommand resolves the editor command without a target path.
// Precedence: explicit config > $VISUAL/$EDITOR > agy-ide > code > nano.
func DefaultCommand(editorConfig string) string {
	cmdStr := strings.TrimSpace(editorConfig)
	if cmdStr == "" {
		cmdStr = strings.TrimSpace(os.Getenv("VISUAL"))
		if cmdStr == "" {
			cmdStr = strings.TrimSpace(os.Getenv("EDITOR"))
		}
	}
	if cmdStr != "" {
		return cmdStr
	}
	if _, err := LookPath("agy-ide"); err == nil {
		return "agy-ide"
	}
	if _, err := LookPath("code"); err == nil {
		return "code"
	}
	return "nano"
}

// Resolve builds the exec command for opening targetPath.
// It splits a configured multi-word command (e.g. "code --wait") and
// appends the target path.
func Resolve(editorConfig, targetPath string) *execx.Cmd {
	parts := strings.Fields(DefaultCommand(editorConfig))
	if len(parts) == 0 {
		parts = []string{"nano"}
	}
	execCmd := execx.Command(parts[0], parts[1:]...)
	execCmd.Args = append(execCmd.Args, targetPath)
	return execCmd
}

// FindByName returns the option matching a display name or command.
func FindByName(available []Option, name string) (Option, bool) {
	for _, opt := range available {
		if opt.Name == name || opt.Command == name {
			return opt, true
		}
	}
	return Option{}, false
}
