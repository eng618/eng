package system

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/utils"
	"github.com/eng618/eng/internal/utils/log"
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

func selectProcess(processes []ProcessInfo) (ProcessInfo, error) {
	if len(processes) == 0 {
		return ProcessInfo{}, errors.New("no processes found")
	}

	options := make([]string, len(processes)+1)
	for i, p := range processes {
		options[i] = fmt.Sprintf("%s (PID %s, User %s)", p.Command, p.PID, p.User)
	}
	options[len(processes)] = "Cancel"

	var selected string
	prompt := &survey.Select{
		Message: "Select process to kill:",
		Options: options,
	}
	err := survey.AskOne(prompt, &selected)
	if err != nil {
		return ProcessInfo{}, err
	}

	if selected == "Cancel" {
		return ProcessInfo{}, errors.New("operation canceled")
	}

	for i, option := range options[:len(processes)] {
		if option == selected {
			return processes[i], nil
		}
	}

	return ProcessInfo{}, errors.New("selection failed")
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
		isVerbose := utils.IsVerbose(cmd) // Get verbosity flag

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
