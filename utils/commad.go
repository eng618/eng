package utils

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/utils/log"
)

// StartChildProcess starts a monitored child process.
// It executes the given command, and sets up a channel to capture os signals to pass to the child process.
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
		c.Process.Signal(<-sigCh)
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
