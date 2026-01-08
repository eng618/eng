package system

import (
	"errors"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/utils"
	"github.com/eng618/eng/internal/utils/log"
)

var KillPortCmd = &cobra.Command{
	Use:   "killPort [port]",
	Short: "Find and kill the process listening on a specific port",
	Long: `This command finds the process ID (PID) listening on the specified network port
using the 'lsof' command and then terminates that process forcefully using 'kill -9'.

Requires 'lsof' and 'kill' commands to be available on the system.
Primarily intended for Unix-like systems (Linux, macOS).`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		portStr := args[0]
		isVerbose := utils.IsVerbose(cmd) // Get verbosity flag

		// Validate that the input is a number
		if _, err := strconv.Atoi(portStr); err != nil {
			log.Error("Invalid port number provided: %s. Port must be an integer.", portStr)
			return
		}

		log.Message("Attempting to find process on port %s...", portStr)

		// --- Step 1: Find the PID using lsof ---
		// Use 'lsof -ti:<port>' which is designed for scripting:
		// -t: Output terse process identifiers (PIDs) only.
		// -i:<port>: Select network files listening on the specified port.
		lsofCmd := exec.Command("lsof", "-ti:"+portStr)
		log.Verbose(isVerbose, "Executing: %s", lsofCmd.String())

		// Use CombinedOutput to capture both stdout and stderr from lsof
		outputBytes, err := lsofCmd.CombinedOutput()
		output := strings.TrimSpace(string(outputBytes))

		// Check for errors from lsof execution
		if err != nil {
			// If lsof exits with an error, it might mean the port is not in use,
			// or lsof itself failed.
			log.Verbose(isVerbose, "lsof command finished with error: %v", err)
			log.Verbose(isVerbose, "lsof output: %s", output)

			// Check if the error is ExitError and output is empty - common case for "port not found"
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) && output == "" {
				log.Warn("No process found listening on port %s.", portStr)
			} else {
				// A different error occurred (e.g., lsof not found, permission denied)
				log.Error("Failed to execute lsof command: %v", err)
				log.Error("lsof output (if any): %s", output)
			}
			return // Stop execution if lsof failed or found nothing
		}

		// If lsof succeeded but returned no output (less common with -t but possible)
		if output == "" {
			log.Warn("lsof ran successfully but found no process ID on port %s.", portStr)
			return
		}

		// We expect a single PID from 'lsof -ti'. Handle multiple lines just in case (though unlikely).
		pids := strings.Fields(output) // Split by whitespace, handles multiple PIDs on separate lines if any
		if len(pids) == 0 {
			log.Warn("lsof command returned empty output after trimming.")
			return
		}

		// --- Step 2: Kill the process(es) found ---
		killedCount := 0
		errorCount := 0
		for _, pid := range pids {
			log.Info("Found process with PID %s on port %s. Attempting to kill...", pid, portStr)

			// Use 'kill -9 <pid>' to forcefully terminate.
			killCmd := exec.Command("kill", "-9", pid)
			log.Verbose(isVerbose, "Executing: %s", killCmd.String())

			// Run kill command
			if err := killCmd.Run(); err != nil {
				log.Error("Failed to kill process with PID %s: %v", pid, err)
				errorCount++
			} else {
				log.Success("Successfully sent kill signal to process with PID %s.", pid)
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
