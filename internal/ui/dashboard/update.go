package dashboard

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/eng618/eng/internal/repo"
)

// statusMsg is returned when a repository's status is loaded.
type statusMsg struct {
	ProjectName string
	RepoURL     string
	Status      RepoStatus
}

func (m Model) Init() tea.Cmd {
	// Tick the spinner so the "Initializing..." view animates,
	// and automatically load statuses for the initially selected project.
	return tea.Batch(m.spinner.Tick, m.loadSelectedProjectStatusesCmd())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height

		var listWidth, listHeight int
		if m.isCompactMode() {
			listWidth = msg.Width - 6
			listHeight = msg.Height - 6
		} else {
			totalPanesWidth := msg.Width - 4
			leftPaneOuterWidth := totalPanesWidth / 4
			if leftPaneOuterWidth < 20 {
				leftPaneOuterWidth = 20
			}
			if leftPaneOuterWidth > 30 {
				leftPaneOuterWidth = 30
			}

			leftStyleWidth := leftPaneOuterWidth - 2
			listWidth = leftStyleWidth - 2
			listHeight = msg.Height - 6
		}

		if listWidth < 1 {
			listWidth = 1
		}
		if listHeight < 1 {
			listHeight = 1
		}
		m.list.SetSize(listWidth, listHeight)
		m.ready = true

	case tea.KeyMsg:
		// While the project list filter input is active, let the list consume
		// all keys (so typing q/f/p/s/c/o/e/t/r/a filters instead of triggering
		// actions or quitting). Ctrl+C still forces quit.
		if m.list.FilterState() == list.Filtering {
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			m.list, cmd = m.list.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}

		// Help overlay closes first — q/Esc/? close it instead of quitting.
		// Any other key also dismisses it (keeps discoverability high).
		if m.showHelp {
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			m.showHelp = false
			return m, nil
		}

		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "q" {
			// Don't accidentally quit mid-action (background io.Pipe
			// goroutine would be orphaned). Ctrl+C still force-quits.
			if m.actionState != "" {
				return m, nil
			}
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
		case "tab":
			if m.focusedPane == FocusLeft {
				m.focusedPane = FocusRight
				m.selectedRepoIndex = 0
			} else {
				m.focusedPane = FocusLeft
			}
			m.clampScrollOffset()
			return m, nil
		case "1":
			m.focusedPane = FocusLeft
			m.clampScrollOffset()
			return m, nil
		case "2":
			m.focusedPane = FocusRight
			m.selectedRepoIndex = 0
			m.clampScrollOffset()
			return m, nil
		case "enter", "l", "right":
			if m.focusedPane == FocusLeft {
				m.focusedPane = FocusRight
				m.selectedRepoIndex = 0
				m.clampScrollOffset()
				return m, nil
			}
		case "esc", "h", "left":
			if m.focusedPane == FocusRight {
				m.focusedPane = FocusLeft
				m.clampScrollOffset()
				return m, nil
			}
		case "j", "down":
			if m.focusedPane == FocusRight {
				item, ok := m.list.SelectedItem().(ProjectItem)
				if ok && m.selectedRepoIndex < len(item.Project.Repos)-1 {
					m.selectedRepoIndex++
				}
				m.clampScrollOffset()
				return m, nil
			}
		case "k", "up":
			if m.focusedPane == FocusRight {
				if m.selectedRepoIndex > 0 {
					m.selectedRepoIndex--
				}
				m.clampScrollOffset()
				return m, nil
			}
		case "?":
			m.showHelp = true
			return m, nil
		case "e":
			cmd, err := m.openInEditorCmd()
			if err != nil {
				m.notificationID++
				m.notification = fmt.Sprintf("Editor error: %v", err)
				m.notificationStyle = notificationErrorStyle
				m.notificationType = NotifyError
				return m, m.delayClearNotificationCmd(m.notificationID)
			}
			return m, cmd
		case "E":
			cmd, err := m.openInCustomEditorCmd()
			if err != nil {
				m.notificationID++
				m.notification = fmt.Sprintf("Editor error: %v", err)
				m.notificationStyle = notificationErrorStyle
				m.notificationType = NotifyError
				return m, m.delayClearNotificationCmd(m.notificationID)
			}
			return m, cmd
		case "t":
			cmd, err := m.openInTerminalCmd()
			if err != nil {
				m.notificationID++
				m.notification = fmt.Sprintf("Terminal error: %v", err)
				m.notificationStyle = notificationErrorStyle
				m.notificationType = NotifyError
				return m, m.delayClearNotificationCmd(m.notificationID)
			}
			return m, cmd
		case "f", "p", "o", "c", "s":
			// Handle actions based on focus
			resModel, cmd := handleAction(m, msg.String())
			m = resModel.(Model)
			m.clampScrollOffset()
			return m, cmd
		case "r":
			m.notificationID++
			m.notification = "Refreshing statuses…"
			m.notificationStyle = notificationWarnStyle
			m.notificationType = NotifyWarn
			return m, tea.Batch(
				m.delayClearNotificationCmd(m.notificationID),
				m.forceRefreshSelectedProjectStatusesCmd(),
			)
		case "a":
			cmd := m.addProjectOrRepoCmd()
			return m, cmd
		}

		if m.focusedPane == FocusLeft {
			previousIndex := m.list.Index()
			m.list, cmd = m.list.Update(msg)
			cmds = append(cmds, cmd)

			if m.list.Index() != previousIndex {
				m.selectedRepoIndex = 0
				m.repoScrollOffset = 0
				cmds = append(cmds, m.loadSelectedProjectStatusesCmd())
			}
		}

	case configUpdateFinishedMsg:
		if msg.err != nil {
			if errors.Is(msg.err, huh.ErrUserAborted) {
				m.notificationID++
				m.notification = "Add canceled."
				m.notificationStyle = notificationWarnStyle
				m.notificationType = NotifyWarn
				return m, m.delayClearNotificationCmd(m.notificationID)
			}
			m.notificationID++
			m.notification = fmt.Sprintf("Error adding repository: %s", msg.err.Error())
			m.notificationStyle = notificationErrorStyle
			m.notificationType = NotifyError
			return m, m.delayClearNotificationCmd(m.notificationID)
		}

		m.projects = msg.projects
		items := make([]list.Item, len(m.projects))
		for i, p := range m.projects {
			items[i] = ProjectItem{Project: p}
		}
		m.list.SetItems(items)

		// Focus the target project
		for idx, item := range m.list.Items() {
			if projItem, ok := item.(ProjectItem); ok && projItem.Project.Name == msg.targetProject {
				m.list.Select(idx)
				break
			}
		}

		// Focus the new repository
		m.selectedRepoIndex = 0
		m.repoScrollOffset = 0
		for _, project := range m.projects {
			if project.Name != msg.targetProject {
				continue
			}
			for idx, r := range project.Repos {
				if r.URL == msg.addedRepo {
					m.selectedRepoIndex = idx
					break
				}
			}
			break
		}

		m.clampScrollOffset()

		m.notificationID++
		prettyName, err := repo.RepoNameFromURL(msg.addedRepo)
		if err != nil {
			prettyName = msg.addedRepo
		}
		m.notification = fmt.Sprintf("Added repo '%s' to project '%s'", prettyName, msg.targetProject)
		m.notificationStyle = notificationSuccessStyle
		m.notificationType = NotifySuccess

		return m, tea.Batch(
			m.delayClearNotificationCmd(m.notificationID),
			m.loadSelectedProjectStatusesCmd(),
		)

	case actionDoneMsg:
		if msg.err != nil {
			m.hasError = true
			m.lastError = msg.err
		}
		if m.actionCurrent >= 0 && m.actionCurrent < len(m.actionRows) {
			row := &m.actionRows[m.actionCurrent]
			switch {
			case msg.err != nil:
				row.Status = ActionFailed
				row.Detail = shortErr(msg.err)
			case row.Status != ActionSkipped:
				row.Status = ActionDone
				if strings.Contains(m.actionTail, "Already up to date") {
					row.Detail = "Up to date"
				} else if row.Detail == "" || row.Detail == "Working…" {
					row.Detail = "Done"
				}
			}
		}
		m.completedActions++

		if len(m.actionQueue) > 0 {
			var cmd tea.Cmd
			m, cmd = m.popAndRunNextAction()
			cmds = append(cmds, cmd)
		} else {
			m.actionState = ""
			m.actionCurrent = -1
			cmds = append(cmds, m.forceRefreshSelectedProjectStatusesCmd())

			m.notificationID++
			if m.hasError {
				if m.lastError != nil {
					m.notification = fmt.Sprintf("Completed with errors: %v", m.lastError)
				} else {
					m.notification = "Completed with errors"
				}
				m.notificationStyle = notificationErrorStyle
				m.notificationType = NotifyError
			} else {
				m.notification = "Action completed successfully"
				m.notificationStyle = notificationSuccessStyle
				m.notificationType = NotifySuccess
			}
			cmds = append(cmds, m.delayClearNotificationCmd(m.notificationID))
			m.hasError = false
			m.lastError = nil
		}

	case clearNotificationMsg:
		if msg.id == m.notificationID {
			m.notification = ""
		}

	case editorFinishedMsg:
		if msg.err != nil {
			m.notificationID++
			m.notification = fmt.Sprintf("Editor exited with error: %v", msg.err)
			m.notificationStyle = notificationErrorStyle
			m.notificationType = NotifyError
			return m, m.delayClearNotificationCmd(m.notificationID)
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
		if m.actionState != "" || !m.ready {
			// Keep ticking while actions run or while initializing.
			cmds = append(cmds, m.spinner.Tick)
		}

	case logLineMsg:
		m.actionLogs = append(m.actionLogs, msg.line)
		if tail := tailLine(msg.line); tail != "" {
			m.actionTail = tail
		}
		cmds = append(cmds, readLogCmd(msg.scanner))

	case statusMsg:
		key := msg.ProjectName + msg.RepoURL
		m.repoStatuses[key] = msg.Status
	}

	m.clampScrollOffset()
	return m, tea.Batch(cmds...)
}

func (m *Model) clampScrollOffset() {
	item, ok := m.list.SelectedItem().(ProjectItem)
	if !ok {
		m.repoScrollOffset = 0
		return
	}
	p := item.Project
	if len(p.Repos) == 0 {
		m.repoScrollOffset = 0
		return
	}

	var innerRightWidth, innerRightHeight int
	if m.isCompactMode() {
		innerRightWidth = m.windowWidth - 6
		innerRightHeight = m.windowHeight - 6
	} else {
		totalPanesWidth := m.windowWidth - 4
		leftPaneOuterWidth := totalPanesWidth / 4
		if leftPaneOuterWidth < 20 {
			leftPaneOuterWidth = 20
		}
		if leftPaneOuterWidth > 30 {
			leftPaneOuterWidth = 30
		}
		rightPaneOuterWidth := totalPanesWidth - leftPaneOuterWidth
		rightStyleWidth := rightPaneOuterWidth - 2
		innerRightWidth = rightStyleWidth - 2
		innerRightHeight = m.windowHeight - 6
	}

	if innerRightHeight < 3 {
		m.repoScrollOffset = 0
		return
	}
	H_repos := innerRightHeight - 4
	if H_repos < 1 {
		H_repos = 1
	}

	if m.selectedRepoIndex < 0 || m.selectedRepoIndex >= len(p.Repos) {
		m.selectedRepoIndex = 0
	}

	if innerRightWidth >= 75 {
		// Table view scroll clamping (one row per repository)
		H_body := H_repos - 2
		if H_body < 1 {
			H_body = 1
		}

		if m.selectedRepoIndex < m.repoScrollOffset {
			m.repoScrollOffset = m.selectedRepoIndex
		} else if m.selectedRepoIndex >= m.repoScrollOffset+H_body {
			m.repoScrollOffset = m.selectedRepoIndex - H_body + 1
		}

		maxScroll := len(p.Repos) - H_body
		if maxScroll < 0 {
			maxScroll = 0
		}
		if m.repoScrollOffset > maxScroll {
			m.repoScrollOffset = maxScroll
		}
		if m.repoScrollOffset < 0 {
			m.repoScrollOffset = 0
		}
	} else {
		// Stacked view scroll clamping (multiple lines per repository)
		allLines, repoStarts, repoEnds := m.getRepoLines()
		if len(allLines) == 0 {
			m.repoScrollOffset = 0
			return
		}

		if m.selectedRepoIndex >= len(repoStarts) {
			m.selectedRepoIndex = 0
		}
		startLine := repoStarts[m.selectedRepoIndex]
		endLine := repoEnds[m.selectedRepoIndex]

		if startLine < m.repoScrollOffset {
			m.repoScrollOffset = startLine
		} else if endLine >= m.repoScrollOffset+H_repos {
			m.repoScrollOffset = endLine - H_repos + 1
		}

		maxScroll := len(allLines) - H_repos
		if maxScroll < 0 {
			maxScroll = 0
		}
		if m.repoScrollOffset > maxScroll {
			m.repoScrollOffset = maxScroll
		}
		if m.repoScrollOffset < 0 {
			m.repoScrollOffset = 0
		}
	}
}

type actionDoneMsg struct {
	err error
}
type clearNotificationMsg struct {
	id int
}

func (m Model) delayClearNotificationCmd(id int) tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return clearNotificationMsg{id: id}
	})
}
