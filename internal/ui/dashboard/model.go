package dashboard

import (
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

type PaneFocus int

const (
	FocusLeft PaneFocus = iota
	FocusRight
)

type NotificationType int

const (
	NotifySuccess NotificationType = iota
	NotifyError
	NotifyWarn
)

// RepoStatus holds the asynchronously loaded status of a repository.
type RepoStatus struct {
	IsCloned       bool
	IsDirty        bool
	Branch         string
	IsDetached     bool
	AheadCount     int
	BehindCount    int
	HasUpstream    bool
	UnstagedCount  int
	StagedCount    int
	UntrackedCount int
	ConflictCount  int
	OngoingOp      string
	Error          error
	Loading        bool
	LastUpdated    time.Time
}

// Repo is a dashboard-local view of a repository: its clone URL plus an
// optional custom directory name. It mirrors config.ProjectRepo without
// importing the config package, keeping presentation decoupled from domain.
type Repo struct {
	URL  string
	Path string
}

// Project is a dashboard-local view of a project: a named group of repos.
type Project struct {
	Name  string
	Repos []Repo
}

// ProjectProvider reloads the current project list (e.g. after the
// `eng project add` subprocess adds a repository). Injected by the caller
// so the dashboard never reads configuration storage directly.
type ProjectProvider func() []Project

// ProjectItem adapts Project to the list.Item interface.
type ProjectItem struct {
	Project Project
}

func (i ProjectItem) Title() string       { return i.Project.Name }
func (i ProjectItem) Description() string { return "" }
func (i ProjectItem) FilterValue() string { return i.Project.Name }

type ActionItem struct {
	Action   string
	RepoName string
	FullPath string
}

// ActionStatus is the Docker-like per-repo state shown in the action modal.
type ActionStatus int

const (
	// ActionPending marks a queued repo that has not started yet.
	ActionPending ActionStatus = iota
	// ActionRunning marks the repo currently being processed.
	ActionRunning
	// ActionDone marks a repo that completed successfully.
	ActionDone
	// ActionSkipped marks a repo that was skipped (not cloned / already cloned).
	ActionSkipped
	// ActionFailed marks a repo whose action returned an error.
	ActionFailed
)

// ActionRow is one stable row in the action modal.
type ActionRow struct {
	RepoName   string
	PrettyName string
	Status     ActionStatus
	Detail     string
}

type configUpdateFinishedMsg struct {
	projects      []Project
	addedRepo     string
	targetProject string
	err           error
}

// Model is the Bubble Tea model for the dashboard.
type Model struct {
	list         list.Model
	projects     []Project
	listProjects ProjectProvider
	repoStatuses map[string]RepoStatus // Keyed by project.Name + repo.URL
	devPath      string
	editor       string

	focusedPane       PaneFocus
	selectedRepoIndex int
	repoScrollOffset  int

	actionState string // empty if idle, otherwise the fixed action title
	actionQueue []ActionItem
	actionLogs  []string
	// actionRows mirrors the queue as stable Docker-like rows.
	actionRows []ActionRow
	// actionCurrent is the index into actionRows currently running (-1 when idle).
	actionCurrent int
	// actionTail is the single-line truncated tail of live output.
	actionTail  string
	actionTitle string
	spinner     spinner.Model

	// Toast notification fields
	notification      string
	notificationStyle lipgloss.Style
	notificationType  NotificationType
	notificationID    int
	hasError          bool
	lastError         error

	// Progress tracking fields
	totalActions     int
	completedActions int

	showHelp bool

	windowWidth  int
	windowHeight int
	ready        bool
}

// NewModel initializes the dashboard model with the given projects.
// provider reloads projects after mutations (may be nil only in tests that
// never add projects; add flows with a nil provider keep current state).
func NewModel(projects []Project, devPath, editor string, provider ProjectProvider) Model {
	items := make([]list.Item, len(projects))
	for i, p := range projects {
		items[i] = ProjectItem{Project: p}
	}

	d := list.NewDefaultDelegate()
	d.ShowDescription = false
	l := list.New(items, d, 0, 0)
	l.Title = "Projects"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.Styles.Title = listTitleStyle

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	m := Model{
		list:              l,
		projects:          projects,
		listProjects:      provider,
		repoStatuses:      make(map[string]RepoStatus),
		devPath:           devPath,
		editor:            editor,
		focusedPane:       FocusLeft,
		selectedRepoIndex: 0,
		actionCurrent:     -1,
		spinner:           s,
	}
	return m
}

func (m Model) isFallbackMode() bool {
	return m.windowWidth < 50 || m.windowHeight < 10
}

func (m Model) isCompactMode() bool {
	return !m.isFallbackMode() && (m.windowWidth < 60 || m.windowHeight < 14)
}
