package dashboard

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/repo"
)

// statusMsg is returned when a repository's status is loaded.
type statusMsg struct {
	ProjectName string
	RepoURL     string
	Status      RepoStatus
}

func (m Model) Init() tea.Cmd {
	// Automatically trigger loading statuses for the initially selected project.
	return m.loadSelectedProjectStatusesCmd()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height

		leftWidth := (msg.Width / 3) - 4
		m.list.SetSize(leftWidth, msg.Height-4)
		m.ready = true

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}

		if m.actionState != "" {
			// Ignore other keys while loading
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}

		switch msg.String() {
		case "enter", "l", "right":
			if m.focusedPane == FocusLeft {
				m.focusedPane = FocusRight
				m.selectedRepoIndex = 0
				return m, nil
			}
		case "esc", "h", "left":
			if m.focusedPane == FocusRight {
				m.focusedPane = FocusLeft
				return m, nil
			}
		case "j", "down":
			if m.focusedPane == FocusRight {
				item, ok := m.list.SelectedItem().(ProjectItem)
				if ok && m.selectedRepoIndex < len(item.Project.Repos)-1 {
					m.selectedRepoIndex++
				}
				return m, nil
			}
		case "k", "up":
			if m.focusedPane == FocusRight {
				if m.selectedRepoIndex > 0 {
					m.selectedRepoIndex--
				}
				return m, nil
			}
		case "f", "p", "o", "c":
			// Handle actions based on focus
			return handleAction(m, msg.String())
		}

		if m.focusedPane == FocusLeft {
			previousIndex := m.list.Index()
			m.list, cmd = m.list.Update(msg)
			cmds = append(cmds, cmd)

			if m.list.Index() != previousIndex {
				cmds = append(cmds, m.loadSelectedProjectStatusesCmd())
			}
		}

	case actionDoneMsg:
		if len(m.actionQueue) > 0 {
			var cmd tea.Cmd
			m, cmd = m.popAndRunNextAction()
			cmds = append(cmds, cmd)
		} else {
			m.actionState = ""
			cmds = append(cmds, m.loadSelectedProjectStatusesCmd())
		}

	case spinner.TickMsg:
		if m.actionState != "" {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case logLineMsg:
		m.actionLogs = append(m.actionLogs, msg.line)
		cmds = append(cmds, readLogCmd(msg.scanner))

	case statusMsg:
		key := msg.ProjectName + msg.RepoURL
		m.repoStatuses[key] = msg.Status
	}

	return m, tea.Batch(cmds...)
}

type actionDoneMsg struct {
	err error
}

func handleAction(m Model, action string) (tea.Model, tea.Cmd) {
	item, ok := m.list.SelectedItem().(ProjectItem)
	if !ok {
		return m, nil
	}
	p := item.Project

	m.actionQueue = []ActionItem{}
	m.actionLogs = []string{}

	if m.focusedPane == FocusRight {
		if len(p.Repos) == 0 {
			return m, nil
		}
		repoDef := p.Repos[m.selectedRepoIndex]
		relPath, _ := repoDef.GetEffectivePath()
		fullPath := filepath.Join(m.devPath, p.Name, relPath)
		m.actionQueue = append(m.actionQueue, ActionItem{
			Action:   action,
			RepoName: repoDef.URL,
			FullPath: fullPath,
		})
	} else {
		for _, repoDef := range p.Repos {
			relPath, _ := repoDef.GetEffectivePath()
			fullPath := filepath.Join(m.devPath, p.Name, relPath)
			m.actionQueue = append(m.actionQueue, ActionItem{
				Action:   action,
				RepoName: repoDef.URL,
				FullPath: fullPath,
			})
		}
	}

	if len(m.actionQueue) == 0 {
		return m, nil
	}

	var cmd tea.Cmd
	m, cmd = m.popAndRunNextAction()

	return m, tea.Batch(m.spinner.Tick, cmd)
}

type logLineMsg struct {
	line    string
	scanner *bufio.Scanner
}

func readLogCmd(scanner *bufio.Scanner) tea.Cmd {
	return func() tea.Msg {
		if scanner.Scan() {
			return logLineMsg{
				line:    scanner.Text(),
				scanner: scanner,
			}
		}
		return actionDoneMsg{err: scanner.Err()}
	}
}

func (m Model) popAndRunNextAction() (Model, tea.Cmd) {
	if len(m.actionQueue) == 0 {
		return m, func() tea.Msg { return actionDoneMsg{} }
	}

	item := m.actionQueue[0]
	m.actionQueue = m.actionQueue[1:]

	var actionName string
	switch item.Action {
	case "f":
		actionName = "Fetching"
	case "p":
		actionName = "Pulling"
	case "s":
		actionName = "Syncing"
	case "c":
		actionName = "Setting up"
	case "o":
		actionName = "Opening"
	}

	m.actionState = fmt.Sprintf("%s %s...", actionName, item.RepoName)

	cloned := isRepoCloned(item.FullPath)
	if item.Action == "c" && cloned {
		return m, func() tea.Msg { return actionDoneMsg{} }
	}
	if item.Action != "c" && !cloned {
		return m, func() tea.Msg { return actionDoneMsg{} }
	}

	pr, pw := io.Pipe()

	go func() {
		// Intercept internal log output
		log.SetWriters(pw, pw)
		defer log.ResetWriters()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var err error

		switch item.Action {
		case "f":
			err = repo.FetchAllPrune(ctx, item.FullPath)
		case "p":
			err = repo.PullLatestCode(ctx, item.FullPath)
		case "s":
			err = repo.FetchAllPrune(ctx, item.FullPath)
			if err == nil {
				err = repo.PullLatestCode(ctx, item.FullPath)
			}
		case "c":
			parentDir := filepath.Dir(item.FullPath)
			err = os.MkdirAll(parentDir, 0o755)
			if err == nil {
				cmd := exec.Command("git", "clone", item.RepoName, item.FullPath)
				cmd.Stdout = pw
				cmd.Stderr = pw
				err = cmd.Run()
			}
		case "o":
			cmd := exec.Command("open", item.FullPath)
			cmd.Stdout = pw
			cmd.Stderr = pw
			err = cmd.Start()
		}

		if err != nil {
			pw.CloseWithError(err)
		} else {
			pw.Close()
		}
	}()

	scanner := bufio.NewScanner(pr)
	return m, readLogCmd(scanner)
}

// loadSelectedProjectStatusesCmd generates tea.Cmds to fetch the status of each repo in the currently selected project.
func (m *Model) loadSelectedProjectStatusesCmd() tea.Cmd {
	item, ok := m.list.SelectedItem().(ProjectItem)
	if !ok {
		return nil
	}

	p := item.Project
	var cmds []tea.Cmd

	for _, r := range p.Repos {
		key := p.Name + r.URL
		// Only load if not already loaded or loading
		if status, exists := m.repoStatuses[key]; !exists || (!status.IsCloned && !status.Loading) {
			// Mark as loading
			s := m.repoStatuses[key]
			s.Loading = true
			m.repoStatuses[key] = s

			// Capture loop variables
			projectName := p.Name
			repoDef := r
			devPath := m.devPath

			cmds = append(cmds, func() tea.Msg {
				return checkRepoStatus(projectName, repoDef, devPath)
			})
		}
	}

	return tea.Batch(cmds...)
}

func checkRepoStatus(projectName string, repoDef config.ProjectRepo, devPath string) tea.Msg {
	// Allow a very brief timeout since this blocks UI refresh momentarily if it hangs
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	repoPath, err := repoDef.GetEffectivePath()
	if err != nil {
		return statusMsg{
			ProjectName: projectName,
			RepoURL:     repoDef.URL,
			Status: RepoStatus{
				Error:   err,
				Loading: false,
			},
		}
	}

	fullPath := filepath.Join(devPath, projectName, repoPath)

	// Wait, isRepoCloned is internal to project package, we might need a quick hack or to use os.Stat
	// Let's just use os.Stat directly to avoid cross-package unexported dependency.
	isCloned := isRepoCloned(fullPath)

	status := RepoStatus{
		IsCloned: isCloned,
		Loading:  false,
	}

	if isCloned {
		dirty, err := repo.IsDirty(ctx, fullPath)
		if err == nil {
			status.IsDirty = dirty
		}

		branch, err := repo.GetCurrentBranch(ctx, fullPath)
		if err == nil {
			status.Branch = branch
		}
	}

	return statusMsg{
		ProjectName: projectName,
		RepoURL:     repoDef.URL,
		Status:      status,
	}
}

func isRepoCloned(repoPath string) bool {
	gitDir := filepath.Join(repoPath, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}
