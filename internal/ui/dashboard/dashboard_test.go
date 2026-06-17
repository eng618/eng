package dashboard

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/eng618/eng/internal/config"
)

func TestDashboardResponsiveLayout(t *testing.T) {
	// 1. Setup mock projects
	projects := []config.Project{
		{
			Name: "TestProject",
			Repos: []config.ProjectRepo{
				{URL: "https://github.com/test/repo1"},
				{URL: "https://github.com/test/repo2"},
				{URL: "https://github.com/test/repo3"},
			},
		},
	}

	m := NewModel(projects, "/tmp/dev", "")

	// 2. Simulate terminal window resize
	width := 120
	height := 30
	msg := tea.WindowSizeMsg{Width: width, Height: height}

	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	if !m.ready {
		t.Error("Expected model to be ready after receiving WindowSizeMsg")
	}

	if m.windowWidth != width || m.windowHeight != height {
		t.Errorf("Expected window dimensions %d x %d, got %d x %d", width, height, m.windowWidth, m.windowHeight)
	}

	// 3. Render and check output height & width
	viewStr := m.View()
	lines := strings.Split(viewStr, "\n")

	actualHeight := len(lines)
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		actualHeight--
	}

	if actualHeight != height {
		t.Errorf(
			"Expected rendered view height to be exactly %d, got %d lines (raw length: %d)",
			height,
			actualHeight,
			len(lines),
		)
	}

	for idx, line := range lines {
		if idx > 0 && idx < actualHeight-1 && lipgloss.Width(line) != width {
			t.Errorf("Line %d has visual width %d, expected exactly %d", idx, lipgloss.Width(line), width)
		}
	}
}

func TestDashboardViewportScrolling(t *testing.T) {
	// Create a project with 40 repositories
	repos := make([]config.ProjectRepo, 40)
	for i := 0; i < 40; i++ {
		repos[i] = config.ProjectRepo{
			URL: fmt.Sprintf("https://github.com/test/repo%d", i),
		}
	}

	projects := []config.Project{
		{
			Name:  "LargeProject",
			Repos: repos,
		},
	}

	m := NewModel(projects, "/tmp/dev", "")

	// Resize window to height 20
	// innerRightHeight = 20 - 6 = 14.
	// H_repos = 14 - 4 = 10 rows.
	height := 20
	m, _ = updateModel(m, tea.WindowSizeMsg{Width: 100, Height: height})

	// Focus right pane
	m.focusedPane = FocusRight

	// Initially, scroll offset is 0, selected index is 0
	m.clampScrollOffset()
	if m.repoScrollOffset != 0 {
		t.Errorf("Expected initial scroll offset 0, got %d", m.repoScrollOffset)
	}

	// Scroll down to repository 25
	for i := 0; i < 25; i++ {
		m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	}

	if m.selectedRepoIndex != 25 {
		t.Errorf("Expected selectedRepoIndex to be 25, got %d", m.selectedRepoIndex)
	}

	// Calculate expected scroll offset bounds
	_, repoStarts, repoEnds := m.getRepoLines()
	startLine := repoStarts[25]
	endLine := repoEnds[25]

	H_repos := (height - 6) - 4 // 10

	// Scroll offset must ensure Repo 25 is visible:
	if startLine < m.repoScrollOffset || endLine >= m.repoScrollOffset+H_repos {
		t.Errorf("Expected repo 25 (lines %d-%d) to be visible in viewport [offset: %d, height: %d]",
			startLine, endLine, m.repoScrollOffset, H_repos)
	}

	// Check if view height & width is still exactly height (20) and width (100)
	viewStr := m.View()
	lines := strings.Split(viewStr, "\n")
	actualHeight := len(lines)
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		actualHeight--
	}
	if actualHeight != height {
		t.Errorf("Expected rendered view height to be exactly %d, got %d lines", height, actualHeight)
	}

	for idx, line := range lines {
		if idx > 0 && idx < actualHeight-1 && lipgloss.Width(line) != 100 {
			t.Errorf("Line %d has visual width %d, expected exactly 100", idx, lipgloss.Width(line))
		}
	}
}

func updateModel(m Model, msg tea.Msg) (Model, tea.Cmd) {
	updated, cmd := m.Update(msg)
	m = updated.(Model)
	if cmd != nil {
		go func(c tea.Cmd) {
			for c != nil {
				res := c()
				if logMsg, ok := res.(logLineMsg); ok {
					c = readLogCmd(logMsg.scanner)
				} else {
					break
				}
			}
		}(cmd)
	}
	return m, cmd
}

func TestDashboardMinimumSizeFallback(t *testing.T) {
	projects := []config.Project{
		{
			Name: "TestProject",
			Repos: []config.ProjectRepo{
				{URL: "https://github.com/test/repo1"},
			},
		},
	}

	m := NewModel(projects, "/tmp/dev", "")

	// 1. Resize terminal below the threshold (e.g. 50 width, 10 height)
	m, _ = updateModel(m, tea.WindowSizeMsg{Width: 50, Height: 10})

	// 2. Call View and assert the fallback screen is displayed
	viewStr := m.View()

	if !strings.Contains(viewStr, "Terminal Too Small") {
		t.Error("Expected fallback screen to display 'Terminal Too Small'")
	}
	if !strings.Contains(viewStr, "Width: 50/60") || !strings.Contains(viewStr, "Height: 10/12") {
		t.Errorf("Expected fallback screen to display current dimensions, got:\n%s", viewStr)
	}

	// 3. Resize terminal back to valid dimensions (e.g. 80 width, 15 height)
	m, _ = updateModel(m, tea.WindowSizeMsg{Width: 80, Height: 15})

	// 4. Call View and assert that it renders the standard dashboard rather than the fallback screen
	viewStr = m.View()
	if strings.Contains(viewStr, "Terminal Too Small") {
		t.Error("Expected standard dashboard layout to render when dimensions are valid, but fallback screen was shown")
	}
}

func TestDashboardCommandsAndNotifications(t *testing.T) {
	tempDev := t.TempDir()

	// Set up a project with one cloned and one not cloned repo
	projects := []config.Project{
		{
			Name: "TestProject",
			Repos: []config.ProjectRepo{
				{URL: "https://github.com/test/cloned-repo"},
				{URL: "https://github.com/test/not-cloned-repo"},
			},
		},
	}

	// Create the .git directory in the cloned repo
	clonedPath := filepath.Join(tempDev, "TestProject", "cloned-repo")
	if err := os.MkdirAll(filepath.Join(clonedPath, ".git"), 0o755); err != nil {
		t.Fatalf("failed to create mock git dir: %v", err)
	}

	m := NewModel(projects, tempDev, "code")
	m.focusedPane = FocusRight
	m.ready = true
	m.windowWidth = 100
	m.windowHeight = 20

	// Test 1: Clone (c) on already cloned repo
	m.selectedRepoIndex = 0 // cloned-repo
	m, cmd := updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if cmd == nil {
		t.Error("Expected a command for delay clearing notification")
	}
	if !strings.Contains(m.notification, "Already cloned") {
		t.Errorf("Expected 'Already cloned' notification, got: %q", m.notification)
	}
	if m.notificationType != NotifyWarn {
		t.Errorf("Expected notification type NotifyWarn, got: %v", m.notificationType)
	}

	// Test 2: Pull (p) on not cloned repo
	m.selectedRepoIndex = 1 // not-cloned-repo
	m, cmd = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	if cmd == nil {
		t.Error("Expected a command for delay clearing notification")
	}
	if !strings.Contains(m.notification, "Not cloned") {
		t.Errorf("Expected 'Not cloned' notification, got: %q", m.notification)
	}
	if m.notificationType != NotifyError {
		t.Errorf("Expected notification type NotifyError, got: %v", m.notificationType)
	}

	// Test 3: Clear notification message clears the correct notification
	notifID := m.notificationID
	m, _ = updateModel(m, clearNotificationMsg{id: notifID})
	if m.notification != "" {
		t.Error("Expected notification to be cleared on clearNotificationMsg")
	}

	// Test 4: Clear notification message with old ID does NOT clear current notification
	m.notification = "New notification"
	m.notificationID = 999
	m, _ = updateModel(m, clearNotificationMsg{id: notifID}) // old ID
	if m.notification != "New notification" {
		t.Error("Expected notification to NOT be cleared by outdated clearNotificationMsg")
	}

	// Test 5: Help toggle "?"
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if !m.showHelp {
		t.Error("Expected showHelp to be true after pressing '?'")
	}

	// Test 6: Pressing any key closes help
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if m.showHelp {
		t.Error("Expected showHelp to be false after pressing any key when active")
	}

	// Test 7: Editor resolution check
	m.focusedPane = FocusRight
	m.selectedRepoIndex = 0 // cloned-repo
	cmd, err := m.openInEditorCmd()
	if err != nil {
		t.Fatalf("Expected no error launching editor, got: %v", err)
	}
	if cmd == nil {
		t.Fatal("Expected tea.Cmd to be returned")
	}

	// Test 8: Batch progress calculation
	m.focusedPane = FocusLeft                                                 // project wide
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}}) // Fetch all
	if m.totalActions != 2 {
		t.Errorf("Expected totalActions to be 2, got %d", m.totalActions)
	}
	if m.completedActions != 0 {
		t.Errorf("Expected completedActions to start at 0, got %d", m.completedActions)
	}

	// Wait for the first background goroutine to exit (it skips and exits immediately)
	time.Sleep(50 * time.Millisecond)

	// Simulate one action done
	m, _ = updateModel(m, actionDoneMsg{})
	if m.completedActions != 1 {
		t.Errorf("Expected completedActions to increment to 1, got %d", m.completedActions)
	}

	// Wait for the second background goroutine to exit too
	time.Sleep(50 * time.Millisecond)
}
