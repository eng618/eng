package system

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui/theme"
)

type ProcessInfo struct {
	Command string
	PID     string
	User    string
}

func listProcesses(filter string) ([]ProcessInfo, error) {
	cmd := exec.Command("ps", "aux")
	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to run ps: %w", err)
	}

	return parseProcessOutput(string(outputBytes), filter)
}

func parseProcessOutput(output, filter string) ([]ProcessInfo, error) {
	var processes []ProcessInfo
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for i, line := range lines {
		if i == 0 {
			continue // Skip header
		}

		fields := strings.Fields(line)
		if len(fields) < 11 {
			continue
		}

		pi := ProcessInfo{
			User:    fields[0],
			PID:     fields[1],
			Command: strings.Join(fields[10:], " "),
		}

		if filter == "" || strings.Contains(strings.ToLower(pi.Command), strings.ToLower(filter)) {
			processes = append(processes, pi)
		}
	}

	return processes, nil
}

type processTableModel struct {
	table    table.Model
	selected ProcessInfo
	canceled bool
}

func (m processTableModel) Init() tea.Cmd { return nil }

func (m processTableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "ctrl+c":
			m.canceled = true
			return m, tea.Quit
		case "enter":
			row := m.table.SelectedRow()
			if len(row) > 0 {
				m.selected = ProcessInfo{
					PID:     row[0],
					User:    row[1],
					Command: row[2],
				}
			}
			return m, tea.Quit
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m processTableModel) View() string {
	return "\n" + m.table.View() + "\n\n  enter: select • q/esc: cancel\n"
}

func selectProcess(processes []ProcessInfo) (ProcessInfo, error) {
	if len(processes) == 0 {
		return ProcessInfo{}, errors.New("no processes found")
	}

	columns := []table.Column{
		{Title: "PID", Width: 10},
		{Title: "User", Width: 15},
		{Title: "Command", Width: 80},
	}

	var rows []table.Row
	for _, p := range processes {
		rows = append(rows, table.Row{p.PID, p.User, p.Command})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(15),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.MutedForeground).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(theme.Background).
		Background(theme.Primary).
		Bold(true)
	t.SetStyles(s)

	m := processTableModel{table: t}
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return ProcessInfo{}, fmt.Errorf("failed to run TUI: %w", err)
	}

	tm := finalModel.(processTableModel)
	if tm.canceled {
		return ProcessInfo{}, errors.New("operation canceled")
	}

	if tm.selected.PID == "" {
		return ProcessInfo{}, errors.New("no process selected")
	}

	return tm.selected, nil
}

var (
	processInteractive bool
	processSignal      string
	processFilter      string
)

var KillProcessCmd = &cobra.Command{
	Use:   "killProcess [pid]",
	Short: "Find and kill a process by PID or interactively",
	Long: `This command finds and kills a process by its PID, or lists processes for interactive selection.

A comma-separated list of PIDs may be provided to kill multiple processes, e.g.
"eng system killProcess 1234,5678".

If no PID is provided or --interactive is used, it lists running processes for selection.
Requires 'ps' and 'kill' commands to be available on the system.
Primarily intended for Unix-like systems (Linux, macOS).`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		isVerbose := cmdutil.IsVerbose(cmd) // Get verbosity flag

		var pidList []string

		if len(args) == 0 || processInteractive {
			log.Message("Listing running processes...")
			processes, err := listProcesses(processFilter)
			if err != nil {
				log.Error("Failed to list processes: %v", err)
				return
			}
			if len(processes) == 0 {
				log.Warn("No processes found.")
				return
			}
			selectedProcess, err := selectProcess(processes)
			if err != nil {
				log.Error("Failed to select process: %v", err)
				return
			}
			pidList = []string{selectedProcess.PID}
		} else {
			log.Message("Parsing process PIDs: %s", args[0])
			var errs []error
			pidList, errs = parseProcessList(args[0])
			if len(errs) > 0 {
				log.Error("Found %d problem(s) in process PID list %q:", len(errs), args[0])
				for _, err := range errs {
					log.Error("  - %v", err)
				}
				return
			}
		}

		for _, pidStr := range pidList {
			killProcess(pidStr, processSignal, isVerbose)
		}
	},
}

// parseProcessList validates a comma-separated list of process PIDs. It returns the
// normalized PID strings and accumulates every validation problem rather than
// failing on the first one, so callers can report all issues at once.
func parseProcessList(input string) ([]string, []error) {
	if strings.TrimSpace(input) == "" {
		return nil, []error{errors.New("process PID list cannot be empty")}
	}

	var pids []string
	var errs []error
	for _, raw := range strings.Split(input, ",") {
		pid := strings.TrimSpace(raw)
		if pid == "" {
			errs = append(errs, fmt.Errorf("empty PID in list %q", input))
			continue
		}
		n, err := strconv.Atoi(pid)
		if err != nil {
			errs = append(errs, fmt.Errorf("invalid PID %q: PID must be an integer", pid))
			continue
		}
		if n <= 0 {
			errs = append(errs, fmt.Errorf("invalid PID %q: PID must be greater than 0", pid))
			continue
		}
		pids = append(pids, pid)
	}
	return pids, errs
}

// killProcess finds and terminates the process with the given PID.
func killProcess(pidStr, signal string, isVerbose bool) {
	log.Message("Attempting to kill process with PID %s...", pidStr)

	// Kill the process
	killCmd := exec.Command("kill", "-"+signal, pidStr)
	log.Verbose(isVerbose, "Executing: %s", killCmd.String())

	// Run kill command
	if err := killCmd.Run(); err != nil {
		log.Error("Failed to kill process with PID %s: %v", pidStr, err)
		if strings.Contains(err.Error(), "permission denied") {
			log.Warn("Try running with sudo: sudo kill -%s %s", signal, pidStr)
		}
	} else {
		log.Success("Successfully sent kill signal %s to process with PID %s.", signal, pidStr)
	}
}

func init() {
	KillProcessCmd.Flags().
		BoolVarP(&processInteractive, "interactive", "i", false, "List processes interactively for selection")
	KillProcessCmd.Flags().
		StringVarP(&processSignal, "signal", "s", "9", "Signal to send to the process (default 9 for SIGKILL)")
	KillProcessCmd.Flags().StringVarP(&processFilter, "filter", "f", "", "Filter processes by command name")
}
