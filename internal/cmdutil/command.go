package cmdutil

import (
	"context"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
)

// StartChildProcess starts a child process with the given exec.Cmd configuration.
// It sets up the standard input, output, and error streams to be the same as the parent process.
// Additionally, it captures interrupt signals (ctl + c) and forwards them to the child process.
//
// Parameters:
//   - c: A pointer to an exec.Cmd struct representing the command to be executed.
//
// Returns:
//   - error: An error if the command fails to start or exits with a non-zero status.
//     Returns nil if the command completes successfully.
//
// The function starts the command and waits for it to finish.
func StartChildProcess(c *exec.Cmd) error {
	c.Stdout = log.Writer()
	c.Stderr = log.ErrorWriter()

	// Set up a signal channel to capture ctl + c, so that we can pass it to the child command.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Starting the dev command
	err := c.Start()
	if err != nil {
		return err // Return the error instead of calling cobra.CheckErr
	}

	go func() {
		// Wait for a signal and forward it to the child process.
		if err := c.Process.Signal(<-sigCh); err != nil {
			log.Error("failed to process command with error: %s", err.Error())
		}
	}()

	// Wait for the command to finish.
	if err := c.Wait(); err != nil {
		// Though the child process failed, we can still log the error and exit gracefully.
		log.Debug("child process exited with error: %s", err) // Log as debug, return the error for handling
		return err                                            // Return the error from Wait()
	}
	log.Success("command completed successfully")
	return nil // Return nil on success
}

// IsVerbose checks if the "verbose" flag is set for the given Cobra command.
// It first checks if the verbose flag is explicitly set on the command.
// If not, it falls back to the config value.
//
// Parameters:
//   - cmd: A pointer to a Cobra command from which the "verbose" flag is retrieved.
//
// Returns:
//   - bool: True if the "verbose" flag is set to true, otherwise false.
func IsVerbose(cmd *cobra.Command) bool {
	// Check if the verbose flag is explicitly set on the command
	if cmd != nil && cmd.Flags().Changed("verbose") {
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			log.Error("failed to get verbose flag: %s", err)
			return false
		}
		return verbose
	}
	// Fallback to config value if flag is not set
	return config.IsVerbose()
}

// FallbackContext returns a default background context.
// It is intended to be used in test environments where cmd.Context() returns nil
// because the command wasn't executed via Execute().
func FallbackContext() context.Context {
	return context.Background()
}
