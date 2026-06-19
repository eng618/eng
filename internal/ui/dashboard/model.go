package dashboard

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"

	"github.com/eng618/eng/internal/config"
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
}

// ProjectItem adapts config.Project to the list.Item interface.
type ProjectItem struct {
	Project config.Project
}

func (i ProjectItem) Title() string       { return i.Project.Name }
func (i ProjectItem) Description() string { return "" }
func (i ProjectItem) FilterValue() string { return i.Project.Name }

type ActionItem struct {
	Action   string
	RepoName string
	FullPath string
}

type configUpdateFinishedMsg struct {
	projects      []config.Project
	addedRepo     string
	targetProject string
	err           error
}

// Model is the Bubble Tea model for the dashboard.
type Model struct {
	list         list.Model
	projects     []config.Project
	repoStatuses map[string]RepoStatus // Keyed by project.Name + repo.URL
	devPath      string
	editor       string

	focusedPane       PaneFocus
	selectedRepoIndex int
	repoScrollOffset  int

	actionState string // empty if idle, otherwise the loading message
	actionQueue []ActionItem
	actionLogs  []string
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

// NewModel initializes the dashboard model with configured projects.
func NewModel(projects []config.Project, devPath, editor string) Model {
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
		repoStatuses:      make(map[string]RepoStatus),
		devPath:           devPath,
		editor:            editor,
		focusedPane:       FocusLeft,
		selectedRepoIndex: 0,
		spinner:           s,
	}
	return m
}
