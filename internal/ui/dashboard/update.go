package dashboard

import (
	"context"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/eng618/eng/internal/config"
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
		rightWidth := msg.Width - leftWidth - 8

		m.list.SetSize(leftWidth, msg.Height-4)
		leftPaneStyle = leftPaneStyle.Width(leftWidth).Height(msg.Height - 4)
		rightPaneStyle = rightPaneStyle.Width(rightWidth).Height(msg.Height - 4)
		m.ready = true

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}
		
		// If the user navigates, we want to load statuses for the newly selected project.
		previousIndex := m.list.Index()
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)

		if m.list.Index() != previousIndex {
			cmds = append(cmds, m.loadSelectedProjectStatusesCmd())
		}

	case statusMsg:
		key := msg.ProjectName + msg.RepoURL
		m.repoStatuses[key] = msg.Status
	}

	return m, tea.Batch(cmds...)
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
