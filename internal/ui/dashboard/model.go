package dashboard

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/eng618/eng/internal/config"
)

type PaneFocus int

const (
	FocusLeft PaneFocus = iota
	FocusRight
)

// RepoStatus holds the asynchronously loaded status of a repository.
type RepoStatus struct {
	IsCloned bool
	IsDirty  bool
	Branch   string
	Error    error
	Loading  bool
}

// ProjectItem adapts config.Project to the list.Item interface.
type ProjectItem struct {
	Project config.Project
}

func (i ProjectItem) Title() string       { return i.Project.Name }
func (i ProjectItem) Description() string { return "Select to view repositories" }
func (i ProjectItem) FilterValue() string { return i.Project.Name }

type ActionItem struct {
	Action   string
	RepoName string
	FullPath string
}

// Model is the Bubble Tea model for the dashboard.
type Model struct {
	list         list.Model
	projects     []config.Project
	repoStatuses map[string]RepoStatus // Keyed by project.Name + repo.URL
	devPath      string

	focusedPane       PaneFocus
	selectedRepoIndex int

	actionState   string // empty if idle, otherwise the loading message
	actionQueue   []ActionItem
	spinner       spinner.Model

	windowWidth  int
	windowHeight int
	ready        bool
}

// NewModel initializes the dashboard model with configured projects.
func NewModel(projects []config.Project, devPath string) Model {
	items := make([]list.Item, len(projects))
	for i, p := range projects {
		items[i] = ProjectItem{Project: p}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Projects"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = listTitleStyle

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	m := Model{
		list:              l,
		projects:          projects,
		repoStatuses:      make(map[string]RepoStatus),
		devPath:           devPath,
		focusedPane:       FocusLeft,
		selectedRepoIndex: 0,
		spinner:           s,
	}
	return m
}
