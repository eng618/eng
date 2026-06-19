package dashboard

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/go-git/go-git/v5"

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

		totalPanesWidth := msg.Width - 4
		leftPaneOuterWidth := totalPanesWidth / 4
		if leftPaneOuterWidth < 20 {
			leftPaneOuterWidth = 20
		}
		if leftPaneOuterWidth > 30 {
			leftPaneOuterWidth = 30
		}

		leftStyleWidth := leftPaneOuterWidth - 2
		listWidth := leftStyleWidth - 2
		if listWidth < 1 {
			listWidth = 1
		}
		listHeight := msg.Height - 6
		if listHeight < 1 {
			listHeight = 1
		}
		m.list.SetSize(listWidth, listHeight)
		m.ready = true

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}

		if m.showHelp {
			m.showHelp = false
			return m, nil
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
		case "f", "p", "o", "c", "s":
			// Handle actions based on focus
			resModel, cmd := handleAction(m, msg.String())
			m = resModel.(Model)
			m.clampScrollOffset()
			return m, cmd
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
		project := config.GetProjectByName(msg.targetProject)
		if project != nil {
			for idx, r := range project.Repos {
				if r.URL == msg.addedRepo {
					m.selectedRepoIndex = idx
					break
				}
			}
		}

		m.clampScrollOffset()

		m.notificationID++
		prettyName, err := config.RepoNameFromURL(msg.addedRepo)
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
		m.completedActions++

		if len(m.actionQueue) > 0 {
			var cmd tea.Cmd
			m, cmd = m.popAndRunNextAction()
			cmds = append(cmds, cmd)
		} else {
			m.actionState = ""
			cmds = append(cmds, m.loadSelectedProjectStatusesCmd())

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

	m.clampScrollOffset()
	return m, tea.Batch(cmds...)
}

func (m *Model) clampScrollOffset() {
	allLines, repoStarts, repoEnds := m.getRepoLines()
	if len(allLines) == 0 {
		m.repoScrollOffset = 0
		return
	}

	innerRightHeight := m.windowHeight - 6
	if innerRightHeight < 5 {
		m.repoScrollOffset = 0
		return
	}
	H_repos := innerRightHeight - 4

	if m.selectedRepoIndex < 0 || m.selectedRepoIndex >= len(repoStarts) {
		m.selectedRepoIndex = 0
	}

	startLine := repoStarts[m.selectedRepoIndex]
	endLine := repoEnds[m.selectedRepoIndex]

	// Adjust scroll offset to make sure the selected repo is visible
	if startLine < m.repoScrollOffset {
		m.repoScrollOffset = startLine
	} else if endLine >= m.repoScrollOffset+H_repos {
		m.repoScrollOffset = endLine - H_repos + 1
	}

	// Clamp scroll offset to valid range
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

func handleAction(m Model, action string) (tea.Model, tea.Cmd) {
	item, ok := m.list.SelectedItem().(ProjectItem)
	if !ok {
		return m, nil
	}
	p := item.Project

	m.actionQueue = []ActionItem{}
	m.actionLogs = []string{}
	m.hasError = false
	m.lastError = nil

	if m.focusedPane == FocusRight {
		if len(p.Repos) == 0 {
			return m, nil
		}
		repoDef := p.Repos[m.selectedRepoIndex]
		relPath, _ := repoDef.GetEffectivePath()
		fullPath := filepath.Join(m.devPath, p.Name, relPath)

		repoName, err := repoDef.GetEffectivePath()
		if err != nil {
			repoName = repoDef.URL
		}

		cloned := isRepoCloned(fullPath)
		if action == "c" && cloned {
			m.notificationID++
			m.notification = fmt.Sprintf("Already cloned: %s", repoName)
			m.notificationStyle = notificationWarnStyle
			m.notificationType = NotifyWarn
			return m, m.delayClearNotificationCmd(m.notificationID)
		}
		if action != "c" && !cloned {
			m.notificationID++
			m.notification = fmt.Sprintf("Not cloned: %s", repoName)
			m.notificationStyle = notificationErrorStyle
			m.notificationType = NotifyError
			return m, m.delayClearNotificationCmd(m.notificationID)
		}

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

	m.totalActions = len(m.actionQueue)
	m.completedActions = 0

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

	prettyName, err := config.RepoNameFromURL(item.RepoName)
	if err != nil {
		prettyName = item.RepoName
	}

	m.actionState = fmt.Sprintf("%s %s...", actionName, prettyName)

	pr, pw := io.Pipe()

	go func() {
		// Intercept internal log output
		log.SetWriters(pw, pw)
		defer log.ResetWriters()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var err error

		cloned := isRepoCloned(item.FullPath)
		if item.Action == "c" && cloned {
			log.Warn("Already cloned: %s", prettyName)
			pw.Close()
			return
		}
		if item.Action != "c" && !cloned {
			log.Warn("Skipping: %s (not cloned)", prettyName)
			pw.Close()
			return
		}

		switch item.Action {
		case "f":
			log.Info("Fetching %s...", prettyName)
			err = repo.FetchAllPrune(ctx, item.FullPath)
			if err == nil {
				log.Success("Fetch completed successfully!")
			}
		case "p":
			log.Info("Pulling %s...", prettyName)
			err = repo.PullLatestCode(ctx, item.FullPath)
			if errors.Is(err, git.NoErrAlreadyUpToDate) {
				log.Info("Already up to date.")
				err = nil
			} else if err == nil {
				log.Success("Pull completed successfully!")
			}
		case "s":
			log.Info("Syncing %s...", prettyName)
			log.Info("1/2 Fetching...")
			err = repo.FetchAllPrune(ctx, item.FullPath)
			if err == nil {
				log.Info("2/2 Pulling...")
				err = repo.PullLatestCode(ctx, item.FullPath)
				if errors.Is(err, git.NoErrAlreadyUpToDate) {
					log.Info("Already up to date.")
					err = nil
				}
			}
			if err == nil {
				log.Success("Sync completed successfully!")
			}
		case "c":
			log.Info("Cloning %s...", prettyName)
			parentDir := filepath.Dir(item.FullPath)
			err = os.MkdirAll(parentDir, 0o755)
			if err == nil {
				cmd := exec.CommandContext(ctx, "git", "clone", item.RepoName, item.FullPath)
				cmd.Stdout = pw
				cmd.Stderr = pw
				err = cmd.Run()
			}
			if err == nil {
				log.Success("Clone completed successfully!")
			}
		case "o":
			log.Info("Opening %s...", item.FullPath)
			cmd := exec.CommandContext(ctx, "open", item.FullPath)
			cmd.Stdout = pw
			cmd.Stderr = pw
			err = cmd.Run()
			if err == nil {
				log.Success("Directory opened.")
			}
		}

		if err != nil {
			log.Error("Failed: %v", err)
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

type editorFinishedMsg struct {
	err error
}

func (m Model) openInEditorCmd() (tea.Cmd, error) {
	item, ok := m.list.SelectedItem().(ProjectItem)
	if !ok {
		return nil, fmt.Errorf("no project selected")
	}
	p := item.Project

	var targetPath string
	if m.focusedPane == FocusRight {
		if len(p.Repos) == 0 {
			return nil, fmt.Errorf("no repository selected")
		}
		repoDef := p.Repos[m.selectedRepoIndex]
		relPath, _ := repoDef.GetEffectivePath()
		targetPath = filepath.Join(m.devPath, p.Name, relPath)

		if !isRepoCloned(targetPath) {
			return nil, fmt.Errorf("repository not cloned yet")
		}
	} else {
		targetPath = filepath.Join(m.devPath, p.Name)
		_ = os.MkdirAll(targetPath, 0o755)
	}

	execCmd := resolveEditorCommand(m.editor, targetPath)

	return tea.ExecProcess(execCmd, func(err error) tea.Msg {
		return editorFinishedMsg{err: err}
	}), nil
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

func findAddedDiff(oldProjs, newProjs []config.Project) (targetProject, addedRepo string) {
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

	// Capture the project list state before execution
	oldProjects := config.GetProjects()

	return tea.ExecProcess(execCmd, func(err error) tea.Msg {
		if err != nil {
			return configUpdateFinishedMsg{err: err}
		}
		newProjects := config.GetProjects()
		targetProj, addedURL := findAddedDiff(oldProjects, newProjects)
		return configUpdateFinishedMsg{
			projects:      newProjects,
			addedRepo:     addedURL,
			targetProject: targetProj,
		}
	})
}
