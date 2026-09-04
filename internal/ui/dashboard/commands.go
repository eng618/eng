package dashboard

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

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

	execCmd := exec.Command(self, "dashboard", "select-editor", targetPath)

	return tea.ExecProcess(execCmd, func(err error) tea.Msg {
		return editorFinishedMsg{err: err}
	}), nil
}

func (m Model) openInTerminalCmd() (tea.Cmd, error) {
	targetPath, err := m.resolveTargetPath()
	if err != nil {
		return nil, err
	}

	// Detect terminal app in fallback chain: Ghostty -> iTerm -> Terminal
	terminalApp := "Terminal"
	if _, err := os.Stat("/Applications/Ghostty.app"); err == nil {
		terminalApp = "Ghostty"
	} else if _, err := os.Stat("/Applications/iTerm.app"); err == nil {
		terminalApp = "iTerm"
	}

	execCmd := exec.Command("open", "-a", terminalApp, targetPath)

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

func resolveEditorCommand(editorConfig, targetPath string) *exec.Cmd {
	cmdStr := editorConfig
	if cmdStr == "" {
		cmdStr = os.Getenv("VISUAL")
		if cmdStr == "" {
			cmdStr = os.Getenv("EDITOR")
		}
	}

	if cmdStr == "" {
		_, err := exec.LookPath("code")
		if err == nil {
			cmdStr = "code"
		} else {
			cmdStr = "vim"
		}
	}

	parts := strings.Fields(cmdStr)
	execCmd := exec.Command(parts[0], parts[1:]...)
	execCmd.Args = append(execCmd.Args, targetPath)

	return execCmd
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

	execCmd := exec.Command(self, args...)

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
