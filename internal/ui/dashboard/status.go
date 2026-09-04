package dashboard

import (
	"context"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/eng618/eng/internal/repo"
)

// statusMsg is returned when a repository's status is loaded.
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

// forceRefreshSelectedProjectStatusesCmd generates tea.Cmds to fetch the status of each repo in the currently selected project, bypassing any caching.
func (m *Model) forceRefreshSelectedProjectStatusesCmd() tea.Cmd {
	item, ok := m.list.SelectedItem().(ProjectItem)
	if !ok {
		return nil
	}

	p := item.Project
	var cmds []tea.Cmd

	for _, r := range p.Repos {
		key := p.Name + r.URL
		s := m.repoStatuses[key]
		s.Loading = true
		s.Error = nil
		m.repoStatuses[key] = s

		// Capture loop variables
		projectName := p.Name
		repoDef := r
		devPath := m.devPath

		cmds = append(cmds, func() tea.Msg {
			return checkRepoStatus(projectName, repoDef, devPath)
		})
	}

	return tea.Batch(cmds...)
}

func checkRepoStatus(projectName string, repoDef Repo, devPath string) tea.Msg {
	// Allow a very brief timeout since this blocks UI refresh momentarily if it hangs
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	repoPath, err := repo.EffectivePath(repoDef.URL, repoDef.Path)
	if err != nil {
		return statusMsg{
			ProjectName: projectName,
			RepoURL:     repoDef.URL,
			Status: RepoStatus{
				Error:       err,
				Loading:     false,
				LastUpdated: time.Now(),
			},
		}
	}

	fullPath := filepath.Join(devPath, projectName, repoPath)

	isCloned := repo.IsCloned(fullPath)

	status := RepoStatus{
		IsCloned: isCloned,
		Loading:  false,
	}

	if isCloned {
		info, err := repo.GetDetailedStatus(ctx, fullPath)
		if err == nil {
			status.Branch = info.Branch
			status.IsDetached = info.IsDetached
			status.AheadCount = info.AheadCount
			status.BehindCount = info.BehindCount
			status.HasUpstream = info.HasUpstream
			status.UnstagedCount = info.UnstagedCount
			status.StagedCount = info.StagedCount
			status.UntrackedCount = info.UntrackedCount
			status.ConflictCount = info.ConflictCount
			status.OngoingOp = info.OngoingOp
			status.IsDirty = info.UnstagedCount > 0 || info.ConflictCount > 0
		} else {
			status.Error = err
		}
	}

	status.LastUpdated = time.Now()

	return statusMsg{
		ProjectName: projectName,
		RepoURL:     repoDef.URL,
		Status:      status,
	}
}
