package utils

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/utils/log"
)

// StartChildProcess starts a child process with the given exec.Cmd configuration.
// It sets up the standard input, output, and error streams to be the same as the parent process.
// Additionally, it captures interrupt signals (ctl + c) and forwards them to the child process.
//
// Parameters:
//   - c: A pointer to an exec.Cmd struct representing the command to be executed.
//
// The function starts the command and waits for it to finish. If the command exits with an error,
// it logs the error and exits gracefully. If the command completes successfully, it logs a success message
// and exits with a status code of 0.
func StartChildProcess(c *exec.Cmd) {
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	// Set up a signal channel to capture ctl + c, so that we can pass it to the child command.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Starting the dev command
	err := c.Start()
	cobra.CheckErr(err)

	go func() {
		// Wait for a signal and forward it to the child process.
		if err := c.Process.Signal(<-sigCh); err != nil {
			log.Fatal("failed to process command with error: %s", err.Error())
		}
	}()

	// Wait for the command to finish.
	if err := c.Wait(); err != nil {
		// Though the child process failed, we can still log the error and exit gracefully.
		log.Error("child process exited with error: %s", err)

		// c.ProcessState.Exited() = false when the process exited because of a signal
		if !c.ProcessState.Exited() {
			os.Exit(0)
		}
	} else {
		log.Success("command completed successfully")
		os.Exit(0)
	}
}

// IsVerbose checks if the "verbose" flag is set for the given Cobra command.
// It retrieves the value of the "verbose" flag from the command's flags and
// returns true if the flag is set to true. If an error occurs while retrieving
// the flag, it returns false.
//
// Parameters:
//   - cmd: A pointer to a Cobra command from which the "verbose" flag is retrieved.
//
// Returns:
//   - bool: True if the "verbose" flag is set to true, otherwise false.
func IsVerbose(cmd *cobra.Command) bool {
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return false
	}
	return verbose
}
