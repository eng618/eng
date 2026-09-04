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

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-git/go-git/v5"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/repo"
)

// statusMsg is returned when a repository's status is loaded.
func handleAction(m Model, action string) (tea.Model, tea.Cmd) {
	item, ok := m.list.SelectedItem().(ProjectItem)
	if !ok {
		return m, nil
	}
	p := item.Project

	m.actionQueue = []ActionItem{}
	m.actionLogs = []string{}
	m.actionRows = nil
	m.actionCurrent = -1
	m.actionTail = ""
	m.actionTitle = ""
	m.hasError = false
	m.lastError = nil

	if m.focusedPane == FocusRight {
		if len(p.Repos) == 0 {
			return m, nil
		}
		repoDef := p.Repos[m.selectedRepoIndex]
		relPath, _ := repo.EffectivePath(repoDef.URL, repoDef.Path)
		fullPath := filepath.Join(m.devPath, p.Name, relPath)

		repoName, err := repo.EffectivePath(repoDef.URL, repoDef.Path)
		if err != nil {
			repoName = repoDef.URL
		}

		cloned := repo.IsCloned(fullPath)
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
			relPath, _ := repo.EffectivePath(repoDef.URL, repoDef.Path)
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

	// Build one stable row per queued repo; the modal title stays fixed.
	m.actionRows = make([]ActionRow, len(m.actionQueue))
	for i, q := range m.actionQueue {
		pretty, err := repo.RepoNameFromURL(q.RepoName)
		if err != nil || pretty == "" {
			pretty = q.RepoName
		}
		m.actionRows[i] = ActionRow{RepoName: q.RepoName, PrettyName: pretty, Status: ActionPending}
	}
	m.actionTitle = actionTitle(action, len(m.actionQueue))

	var cmd tea.Cmd
	m, cmd = m.popAndRunNextAction()

	return m, tea.Batch(m.spinner.Tick, cmd)
}

// actionTitle returns a fixed modal title that never includes a repo name,
// so the modal width stays stable while rows update underneath.
func actionTitle(action string, n int) string {
	var verb string
	switch action {
	case "f":
		verb = "Fetching"
	case "p":
		verb = "Pulling"
	case "s":
		verb = "Syncing"
	case "c":
		verb = "Cloning"
	case "o":
		verb = "Opening"
	default:
		verb = "Working"
	}
	if n == 1 {
		return verb + " 1 repository"
	}
	return fmt.Sprintf("%s %d repositories", verb, n)
}

// tailLine condenses a raw log line into a short single-line tail.
// It strips log prefixes/emojis so the modal footer stays stable.
func tailLine(line string) string {
	s := strings.TrimSpace(line)
	for _, p := range []string{"✓ ", "→ ", "··· ", "⚠ ", "✗ ", "==> ", "==x ", "--- "} {
		s = strings.TrimPrefix(s, p)
	}
	s = strings.Join(strings.Fields(s), " ")
	return s
}

// shortErr condenses an error for the fixed-width Detail column.
func shortErr(err error) string {
	if err == nil {
		return ""
	}
	s := strings.Join(strings.Fields(strings.TrimSpace(err.Error())), " ")
	return s
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

	prettyName, err := repo.RepoNameFromURL(item.RepoName)
	if err != nil || prettyName == "" {
		prettyName = item.RepoName
	}

	// Fixed title keeps the modal box stable; per-repo progress lives in rows.
	if m.actionTitle == "" {
		m.actionTitle = actionTitle(item.Action, m.totalActions)
	}
	m.actionState = m.actionTitle
	m.actionCurrent++
	if m.actionCurrent >= 0 && m.actionCurrent < len(m.actionRows) {
		m.actionRows[m.actionCurrent].Status = ActionRunning
		m.actionRows[m.actionCurrent].Detail = "Working…"
	}
	m.actionTail = prettyName

	// Fast-path skips stay sequential but never spawn a pipe: mark the row
	// now so the modal shows Skipped instead of a flickering Running state.
	cloned := repo.IsCloned(item.FullPath)
	if item.Action == "c" && cloned {
		if m.actionCurrent >= 0 && m.actionCurrent < len(m.actionRows) {
			m.actionRows[m.actionCurrent].Status = ActionSkipped
			m.actionRows[m.actionCurrent].Detail = "Already cloned"
		}
		return m, func() tea.Msg { return actionDoneMsg{} }
	}
	if item.Action != "c" && item.Action != "o" && !cloned {
		if m.actionCurrent >= 0 && m.actionCurrent < len(m.actionRows) {
			m.actionRows[m.actionCurrent].Status = ActionSkipped
			m.actionRows[m.actionCurrent].Detail = "Not cloned"
		}
		return m, func() tea.Msg { return actionDoneMsg{} }
	}

	pr, pw := io.Pipe()

	go runActionItem(item, prettyName, pw)

	scanner := bufio.NewScanner(pr)
	return m, readLogCmd(scanner)
}

// runActionItem executes one queued repo action, streaming log output to pw.
// It runs in its own goroutine; completion surfaces as actionDoneMsg via
// the pipe scanner.
func runActionItem(item ActionItem, prettyName string, pw *io.PipeWriter) {
	// Intercept internal log output
	log.SetWriters(pw, pw)
	defer log.ResetWriters()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cloned := repo.IsCloned(item.FullPath)
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

	if err := executeRepoAction(ctx, item, prettyName); err != nil {
		log.Error("Failed: %v", err)
		pw.CloseWithError(err)
		return
	}
	pw.Close()
}

// executeRepoAction runs the git operation for a single repository.
func executeRepoAction(ctx context.Context, item ActionItem, prettyName string) error {
	switch item.Action {
	case "f":
		return fetchActionRepo(ctx, item, prettyName)
	case "p":
		return pullActionRepo(ctx, item, prettyName)
	case "s":
		return syncActionRepo(ctx, item, prettyName)
	case "c":
		return cloneActionRepo(ctx, item, prettyName)
	case "o":
		return openActionRepo(ctx, item)
	default:
		return fmt.Errorf("unknown action %q", item.Action)
	}
}

func fetchActionRepo(ctx context.Context, item ActionItem, prettyName string) error {
	log.Info("Fetching %s...", prettyName)
	if err := repo.FetchAllPrune(ctx, item.FullPath); err != nil {
		return err
	}
	log.Success("Fetch completed successfully!")
	return nil
}

func pullActionRepo(ctx context.Context, item ActionItem, prettyName string) error {
	log.Info("Pulling %s...", prettyName)
	err := repo.PullLatestCode(ctx, item.FullPath)
	if errors.Is(err, git.NoErrAlreadyUpToDate) {
		log.Info("Already up to date.")
		return nil
	}
	if err == nil {
		log.Success("Pull completed successfully!")
	}
	return err
}

func syncActionRepo(ctx context.Context, item ActionItem, prettyName string) error {
	log.Info("Syncing %s...", prettyName)
	log.Info("1/2 Fetching...")
	if err := repo.FetchAllPrune(ctx, item.FullPath); err != nil {
		return err
	}
	log.Info("2/2 Pulling...")
	if err := pullActionRepo(ctx, item, prettyName); err != nil {
		return err
	}
	log.Success("Sync completed successfully!")
	return nil
}

func cloneActionRepo(ctx context.Context, item ActionItem, prettyName string) error {
	log.Info("Cloning %s...", prettyName)
	parentDir := filepath.Dir(item.FullPath)
	if err := os.MkdirAll(parentDir, 0o755); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "git", "clone", item.RepoName, item.FullPath)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()
	if err := cmd.Run(); err != nil {
		return err
	}
	log.Success("Clone completed successfully!")
	return nil
}

func openActionRepo(ctx context.Context, item ActionItem) error {
	log.Info("Opening %s...", item.FullPath)
	cmd := exec.CommandContext(ctx, "open", item.FullPath)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()
	if err := cmd.Run(); err != nil {
		return err
	}
	log.Success("Directory opened.")
	return nil
}
