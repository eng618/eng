package system

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/log"
)

type PortInfo struct {
	Command string
	PID     string
	Port    string
	User    string
}

func findPortTool() string {
	if _, err := exec.LookPath("lsof"); err == nil {
		return "lsof"
	}
	if _, err := exec.LookPath("ss"); err == nil {
		return "ss"
	}
	if _, err := exec.LookPath("netstat"); err == nil {
		return "netstat"
	}
	return ""
}

func listPorts(filter string) ([]PortInfo, error) {
	tool := findPortTool()
	if tool == "" {
		return nil, errors.New("no suitable tool found for listing ports (lsof, ss, netstat)")
	}

	var cmd *exec.Cmd
	switch tool {
	case "lsof":
		cmd = exec.Command("lsof", "-i", "-P", "-n")
	case "ss":
		cmd = exec.Command("ss", "-tulpn")
	case "netstat":
		cmd = exec.Command("netstat", "-tulpn")
	}

	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to run %s: %w", tool, err)
	}

	return parsePortOutput(string(outputBytes), tool, filter)
}

func parsePortOutput(output, tool, filter string) ([]PortInfo, error) {
	var ports []PortInfo
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for i, line := range lines {
		if i == 0 || !strings.Contains(line, "LISTEN") {
			continue // Skip header or non-listening
		}

		var pi PortInfo
		switch tool {
		case "lsof":
			// COMMAND PID USER FD TYPE DEVICE SIZE/OFF NODE NAME
			fields := strings.Fields(line)
			if len(fields) < 9 {
				continue
			}
			pi.Command = fields[0]
			pi.PID = fields[1]
			pi.User = fields[2]
			name := fields[8] // NAME field

			if match := lsofPortRe.FindStringSubmatch(name); len(match) > 1 {
				pi.Port = match[1]
			}
		case "ss":
			// Netid State Recv-Q Send-Q Local Address:Port Peer Address:Port Process
			fields := strings.Fields(line)
			if len(fields) < 6 {
				continue
			}
			local := fields[4]
			if strings.Contains(local, ":") {
				parts := strings.Split(local, ":")
				pi.Port = parts[len(parts)-1]
			}
			process := fields[len(fields)-1]
			if strings.Contains(process, "pid=") {

				if match := ssPidRe.FindStringSubmatch(process); len(match) > 1 {
					pi.PID = match[1]
				}

				if match := ssCmdRe.FindStringSubmatch(process); len(match) > 1 {
					pi.Command = match[1]
				}
			}
		case "netstat":
			// Proto Recv-Q Send-Q Local Address Foreign Address State PID/Program name
			fields := strings.Fields(line)
			if len(fields) < 7 {
				continue
			}
			local := fields[3]
			if strings.Contains(local, ":") {
				parts := strings.Split(local, ":")
				pi.Port = parts[len(parts)-1]
			}
			pidProg := fields[len(fields)-1]
			if strings.Contains(pidProg, "/") {
				parts := strings.Split(pidProg, "/")
				pi.PID = parts[0]
				pi.Command = parts[1]
			}
		}

		if pi.Port != "" && (filter == "" || strings.Contains(strings.ToLower(pi.Command), strings.ToLower(filter))) {
			ports = append(ports, pi)
		}
	}

	return ports, nil
}

type portTableModel struct {
	table    table.Model
	selected PortInfo
	canceled bool
}

func (m portTableModel) Init() tea.Cmd { return nil }

func (m portTableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				m.selected = PortInfo{
					PID:     row[0],
					Command: row[1],
					User:    row[2],
					Port:    row[3],
				}
			}
			return m, tea.Quit
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m portTableModel) View() string {
	return "\n" + m.table.View() + "\n\n  enter: select • q/esc: cancel\n"
}

func selectPort(ports []PortInfo) (PortInfo, error) {
	if len(ports) == 0 {
		return PortInfo{}, errors.New("no ports found")
	}

	columns := []table.Column{
		{Title: "PID", Width: 10},
		{Title: "Command", Width: 20},
		{Title: "User", Width: 15},
		{Title: "Port", Width: 10},
	}

	var rows []table.Row
	for _, p := range ports {
		rows = append(rows, table.Row{p.PID, p.Command, p.User, p.Port})
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

	m := portTableModel{table: t}
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return PortInfo{}, fmt.Errorf("failed to run TUI: %w", err)
	}

	tm := finalModel.(portTableModel)
	if tm.canceled {
		return PortInfo{}, errors.New("operation canceled")
	}

	if tm.selected.PID == "" {
		return PortInfo{}, errors.New("no port selected")
	}

	return tm.selected, nil
}

var (
	interactive bool
	signal      string
	filter      string

	lsofPortRe = regexp.MustCompile(`:(\d+)`)
	ssPidRe    = regexp.MustCompile(`pid=(\d+)`)
	ssCmdRe    = regexp.MustCompile(`\("([^"]+)"`)
)

var KillPortCmd = &cobra.Command{
	Use:   "killPort [port]",
	Short: "Find and kill the process listening on a specific port",
	Long: `This command finds the process ID (PID) listening on the specified network port
using available tools (lsof, ss, netstat) and then terminates that process.

If no port is provided or --interactive is used, it lists listening ports for selection.
Requires appropriate tools to be available on the system.
Primarily intended for Unix-like systems (Linux, macOS).`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		isVerbose := cmdutil.IsVerbose(cmd) // Get verbosity flag

		var portStr string
		var selectedPort PortInfo

		if len(args) == 0 || interactive {
			log.Message("Listing listening ports...")
			ports, err := listPorts(filter)
			if err != nil {
				log.Error("Failed to list ports: %v", err)
				return
			}
			if len(ports) == 0 {
				log.Warn("No listening ports found.")
				return
			}
			selectedPort, err = selectPort(ports)
			if err != nil {
				log.Error("Failed to select port: %v", err)
				return
			}
			portStr = selectedPort.Port
		} else {
			portStr = args[0]
			// Validate that the input is a number
			if _, err := strconv.Atoi(portStr); err != nil {
				log.Error("Invalid port number provided: %s. Port must be an integer.", portStr)
				return
			}
		}

		log.Message("Attempting to find process on port %s...", portStr)

		// Find PID using the tool
		tool := findPortTool()
		if tool == "" {
			log.Error("No suitable tool found for finding processes on ports.")
			return
		}

		var lsofCmd *exec.Cmd
		switch tool {
		case "lsof":
			lsofCmd = exec.Command("lsof", "-ti:"+portStr)
		case "ss":
			// ss -tulpn | grep :port | awk '{print $7}' | sed 's/.*pid=\([0-9]*\).*/\1/'
			lsofCmd = exec.Command(
				"sh",
				"-c",
				fmt.Sprintf("ss -tulpn | grep ':%s ' | grep -o 'pid=[0-9]*' | cut -d'=' -f2 | head -1", portStr),
			)
		case "netstat":
			lsofCmd = exec.Command(
				"sh",
				"-c",
				fmt.Sprintf("netstat -tulpn | grep ':%s ' | awk '{print $7}' | cut -d'/' -f1 | head -1", portStr),
			)
		}
		log.Verbose(isVerbose, "Executing: %s", lsofCmd.String())

		// Use CombinedOutput to capture both stdout and stderr from command
		outputBytes, err := lsofCmd.CombinedOutput()
		output := strings.TrimSpace(string(outputBytes))

		// Check for errors from command execution
		if err != nil {
			// If lsof exits with an error, it might mean the port is not in use,
			// or lsof itself failed.
			log.Verbose(isVerbose, "Command finished with error: %v", err)
			log.Verbose(isVerbose, "lsof output: %s", output)

			// Check if the error is ExitError and output is empty - common case for "port not found"
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) && output == "" {
				log.Warn("No process found listening on port %s.", portStr)
			} else {
				// A different error occurred (e.g., lsof not found, permission denied)
				log.Error("Failed to execute command: %v", err)
				log.Error("Command output (if any): %s", output)
			}
			return // Stop execution if lsof failed or found nothing
		}

		// If lsof succeeded but returned no output (less common with -t but possible)
		if output == "" {
			log.Warn("lsof ran successfully but found no process ID on port %s.", portStr)
			return
		}

		// We expect a single PID from the command. Handle multiple lines just in case.
		pids := strings.Fields(output) // Split by whitespace, handles multiple PIDs on separate lines if any
		if len(pids) == 0 {
			log.Warn("Command ran successfully but found no process ID on port %s.", portStr)
			return
		}

		// --- Step 2: Kill the process(es) found ---
		killedCount := 0
		errorCount := 0
		for _, pid := range pids {
			log.Info("Found process with PID %s on port %s. Attempting to kill...", pid, portStr)

			// Use 'kill -<signal> <pid>' to terminate.
			killCmd := exec.Command("kill", "-"+signal, pid)
			log.Verbose(isVerbose, "Executing: %s", killCmd.String())

			// Run kill command
			if err := killCmd.Run(); err != nil {
				log.Error("Failed to kill process with PID %s: %v", pid, err)
				if strings.Contains(err.Error(), "permission denied") {
					log.Warn("Try running with sudo: sudo kill -%s %s", signal, pid)
				}
				errorCount++
			} else {
				log.Success("Successfully sent kill signal %s to process with PID %s.", signal, pid)
				killedCount++
			}
		}

		// --- Final Summary ---
		if killedCount > 0 && errorCount == 0 {
			log.Success("Finished killing process(es) on port %s.", portStr)
		} else if killedCount > 0 && errorCount > 0 {
			log.Warn(
				"Finished attempting to kill process(es) on port %s, but encountered %d error(s).",
				portStr,
				errorCount,
			)
		} else if killedCount == 0 && errorCount > 0 {
			log.Error("Failed to kill any process found on port %s.", portStr)
		}
		// If killedCount == 0 and errorCount == 0, it means lsof found nothing, already handled earlier.
	},
}

func init() {
	KillPortCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "List ports interactively for selection")
	KillPortCmd.Flags().StringVarP(&signal, "signal", "s", "9", "Signal to send to the process (default 9 for SIGKILL)")
	KillPortCmd.Flags().StringVarP(&filter, "filter", "f", "", "Filter ports by command name")
}
