package dashboard

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/eng618/eng/internal/editor"
	"github.com/eng618/eng/internal/execx"
	"github.com/eng618/eng/internal/repo"
)

// statusMsg is returned when a repository's status is loaded.
type editorFinishedMsg struct {
	err error
}

func (m Model) openInEditorCmd() (tea.Cmd, error) {
	targetPath, err := m.resolveTargetPath()
	if err != nil {
		return nil, err
	}

	execCmd := resolveEditorCommand(m.editor, targetPath)

	return tea.ExecProcess(execCmd, func(err error) tea.Msg {
		return editorFinishedMsg{err: err}
	}), nil
}

func (m Model) openInCustomEditorCmd() (tea.Cmd, error) {
	targetPath, err := m.resolveTargetPath()
	if err != nil {
		return nil, err
	}

	self, err := os.Executable()
	if err != nil {
		self = "eng"
	}

	execCmd := execx.Command(self, "dashboard", "select-editor", targetPath)

	return tea.ExecProcess(execCmd, func(err error) tea.Msg {
		return editorFinishedMsg{err: err}
	}), nil
}

func (m Model) openInTerminalCmd() (tea.Cmd, error) {
	targetPath, err := m.resolveTargetPath()
	if err != nil {
		return nil, err
	}

	switch runtime.GOOS {
	case "darwin":
		return m.openInTerminalDarwin(targetPath)
	case "linux":
		return m.openInTerminalLinux(targetPath)
	default:
		return nil, fmt.Errorf("opening a new terminal is not supported on %s", runtime.GOOS)
	}
}

func (m Model) openInTerminalDarwin(targetPath string) (tea.Cmd, error) {
	// Detect terminal app in fallback chain: Ghostty -> iTerm -> Terminal
	terminalApp := "Terminal"
	if _, err := os.Stat("/Applications/Ghostty.app"); err == nil {
		terminalApp = "Ghostty"
	} else if _, err := os.Stat("/Applications/iTerm.app"); err == nil {
		terminalApp = "iTerm"
	}

	execCmd := execx.Command("open", "-a", terminalApp, targetPath)

	return tea.ExecProcess(execCmd, func(err error) tea.Msg {
		return editorFinishedMsg{err: err}
	}), nil
}

func (m Model) openInTerminalLinux(targetPath string) (tea.Cmd, error) {
	// Look for known terminal emulators in fallback chain.
	// $TERMINAL always wins so users can pin a preferred emulator.
	candidates := []string{
		"ghostty", "kitty", "alacritty", "wezterm",
		"gnome-terminal", "konsole", "xfce4-terminal", "xterm",
	}
	terminalApp := os.Getenv("TERMINAL")
	if terminalApp == "" {
		for _, candidate := range candidates {
			if _, err := execx.LookPath(candidate); err == nil {
				terminalApp = candidate
				break
			}
		}
	}
	if terminalApp == "" {
		return nil, fmt.Errorf("no supported terminal emulator found; set $TERMINAL")
	}

	var args []string
	switch terminalApp {
	case "wezterm":
		args = []string{"start", "--cwd", targetPath}
	case "gnome-terminal":
		args = []string{"--working-directory=" + targetPath}
	case "konsole":
		args = []string{"--workdir", targetPath}
	default:
		args = []string{"--working-directory", targetPath}
	}

	execCmd := execx.Command(terminalApp, args...)

	return tea.ExecProcess(execCmd, func(err error) tea.Msg {
		return editorFinishedMsg{err: err}
	}), nil
}

// resolveTargetPath maps the current selection to a filesystem path,
// creating the project directory when focusing the left pane.
func (m Model) resolveTargetPath() (string, error) {
	item, ok := m.list.SelectedItem().(ProjectItem)
	if !ok {
		return "", fmt.Errorf("no project selected")
	}
	p := item.Project

	if m.focusedPane != FocusRight {
		targetPath := filepath.Join(m.devPath, p.Name)
		_ = os.MkdirAll(targetPath, 0o755)
		return targetPath, nil
	}

	if len(p.Repos) == 0 {
		return "", fmt.Errorf("no repository selected")
	}
	repoDef := p.Repos[m.selectedRepoIndex]
	relPath, _ := repo.EffectivePath(repoDef.URL, repoDef.Path)
	targetPath := filepath.Join(m.devPath, p.Name, relPath)

	if !repo.IsCloned(targetPath) {
		return "", fmt.Errorf("repository not cloned yet")
	}
	return targetPath, nil
}

func resolveEditorCommand(editorConfig, targetPath string) *execx.Cmd {
	// Delegates to the shared editor package so dashboard open logic,
	// the select-editor picker, and `eng config git-editor` stay in sync.
	// Precedence: explicit `git.editor` > $VISUAL/$EDITOR > agy-ide > code > nano.
	return editor.Resolve(editorConfig, targetPath)
}

func findAddedDiff(oldProjs, newProjs []Project) (targetProject, addedRepo string) {
	oldRepos := make(map[string]bool)
	for _, p := range oldProjs {
		for _, r := range p.Repos {
			oldRepos[p.Name+":"+r.URL] = true
		}
	}

	for _, p := range newProjs {
		for _, r := range p.Repos {
			if !oldRepos[p.Name+":"+r.URL] {
				return p.Name, r.URL
			}
		}
	}
	return "", ""
}

func (m Model) addProjectOrRepoCmd() tea.Cmd {
	var preSelectedProject string
	if m.focusedPane == FocusRight {
		if item, ok := m.list.SelectedItem().(ProjectItem); ok {
			preSelectedProject = item.Project.Name
		}
	}

	self, err := os.Executable()
	if err != nil {
		self = "eng"
	}

	var args []string
	args = append(args, "project", "add")
	if preSelectedProject != "" {
		args = append(args, "-p", preSelectedProject)
	}

	execCmd := execx.Command(self, args...)

	// Capture the project list state before execution. When no provider is
	// wired (tests), fall back to the in-memory list so the diff is empty
	// and the model simply refreshes.
	oldProjects := m.projects
	if m.listProjects != nil {
		oldProjects = m.listProjects()
	}

	return tea.ExecProcess(execCmd, func(err error) tea.Msg {
		if err != nil {
			return configUpdateFinishedMsg{err: err}
		}
		newProjects := m.projects
		if m.listProjects != nil {
			newProjects = m.listProjects()
		}
		targetProj, addedURL := findAddedDiff(oldProjects, newProjects)
		return configUpdateFinishedMsg{
			projects:      newProjects,
			addedRepo:     addedURL,
			targetProject: targetProj,
		}
	})
}
