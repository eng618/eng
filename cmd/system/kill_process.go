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
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
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
If no PID is provided or --interactive is used, it lists running processes for selection.
Requires 'ps' and 'kill' commands to be available on the system.
Primarily intended for Unix-like systems (Linux, macOS).`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		isVerbose := cmdutil.IsVerbose(cmd) // Get verbosity flag

		var pidStr string
		var selectedProcess ProcessInfo

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
			selectedProcess, err = selectProcess(processes)
			if err != nil {
				log.Error("Failed to select process: %v", err)
				return
			}
			pidStr = selectedProcess.PID
		} else {
			pidStr = args[0]
			// Validate that the input is a number
			if _, err := strconv.Atoi(pidStr); err != nil {
				log.Error("Invalid PID provided: %s. PID must be an integer.", pidStr)
				return
			}
		}

		log.Message("Attempting to kill process with PID %s...", pidStr)

		// Kill the process
		killCmd := exec.Command("kill", "-"+processSignal, pidStr)
		log.Verbose(isVerbose, "Executing: %s", killCmd.String())

		// Run kill command
		if err := killCmd.Run(); err != nil {
			log.Error("Failed to kill process with PID %s: %v", pidStr, err)
			if strings.Contains(err.Error(), "permission denied") {
				log.Warn("Try running with sudo: sudo kill -%s %s", processSignal, pidStr)
			}
		} else {
			log.Success("Successfully sent kill signal %s to process with PID %s.", processSignal, pidStr)
		}
	},
}

func init() {
	KillProcessCmd.Flags().
		BoolVarP(&processInteractive, "interactive", "i", false, "List processes interactively for selection")
	KillProcessCmd.Flags().
		StringVarP(&processSignal, "signal", "s", "9", "Signal to send to the process (default 9 for SIGKILL)")
	KillProcessCmd.Flags().StringVarP(&processFilter, "filter", "f", "", "Filter processes by command name")
}
