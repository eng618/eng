package dashboard

import (
	"context"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestActionTimeout(t *testing.T) {
	if got := actionTimeout("c"); got != 10*time.Minute {
		t.Errorf("clone timeout = %v, want 10m", got)
	}
	if got := actionTimeout("o"); got != 30*time.Second {
		t.Errorf("open timeout = %v, want 30s", got)
	}
	for _, action := range []string{"f", "F", "p", "P", "s", "S"} {
		if got := actionTimeout(action); got <= 2*time.Minute {
			t.Errorf(
				"action %q timeout = %v, want more than repo's 2m git timeout "+
					"so git exits cleanly before the outer context fires",
				action,
				got,
			)
		}
	}
}

func actionTestModel(t *testing.T) Model {
	t.Helper()
	projects := []Project{
		{
			Name:  "TestProject",
			Repos: []Repo{{URL: "https://github.com/test/repo1"}},
		},
	}
	m := NewModel(projects, t.TempDir(), "", nil)
	m.actionState = "Fetching 1 repository"
	m.actionTitle = m.actionState
	return m
}

func isQuitCmd(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	msg := cmd()
	_, ok := msg.(tea.QuitMsg)
	return ok
}

func TestQuitDuringActionIsGraceful(t *testing.T) {
	cancelled := false
	m := actionTestModel(t)
	m.actionCancel = context.CancelFunc(func() { cancelled = true })
	m.actionQueue = []ActionItem{{Action: "f", RepoName: "x", FullPath: "/tmp/x"}}

	// q must not quit immediately: it drops the queue and waits for the
	// running git operation to release its locks.
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = updated.(Model)
	if isQuitCmd(cmd) {
		t.Error("q during action must not quit immediately (git would be orphaned)")
	}
	if !m.actionQuitRequested {
		t.Error("Expected actionQuitRequested to be set after q during action")
	}
	if len(m.actionQueue) != 0 {
		t.Error("Expected remaining queue to be dropped on graceful quit")
	}
	if cancelled {
		t.Error("q must let the running operation finish cleanly, not cancel it")
	}

	// Once the current action reports back, the program quits.
	updated, cmd = m.Update(actionDoneMsg{})
	if !isQuitCmd(cmd) {
		t.Error("Expected tea.Quit after the final actionDoneMsg with quit requested")
	}
	_ = updated
}

func TestCtrlCDuringActionCancelsThenQuits(t *testing.T) {
	cancelled := false
	m := actionTestModel(t)
	m.actionCancel = context.CancelFunc(func() { cancelled = true })
	m.actionQueue = []ActionItem{{Action: "f", RepoName: "x", FullPath: "/tmp/x"}}

	// First ctrl+c cancels the running git op but waits for lock release.
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	m = updated.(Model)
	if isQuitCmd(cmd) {
		t.Error("First ctrl+c during action must wait for git to exit, not quit immediately")
	}
	if !cancelled {
		t.Error("Expected first ctrl+c to cancel the running git operation")
	}
	if !m.actionQuitRequested {
		t.Error("Expected actionQuitRequested to be set after first ctrl+c")
	}

	// Second ctrl+c force-quits.
	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if !isQuitCmd(cmd) {
		t.Error("Expected second ctrl+c during action to force quit")
	}
}

func TestCtrlCWhenIdleQuitsImmediately(t *testing.T) {
	projects := []Project{
		{
			Name:  "TestProject",
			Repos: []Repo{{URL: "https://github.com/test/repo1"}},
		},
	}
	m := NewModel(projects, t.TempDir(), "", nil)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if !isQuitCmd(cmd) {
		t.Error("Expected ctrl+c when idle to quit immediately")
	}
}
