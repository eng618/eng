package system

import (
	"fmt"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/utils"
	"github.com/eng618/eng/internal/utils/log"
)

// UpdateCmd represents the system update command.
// It provides functionality to update system packages and perform cleanup operations.
// Supports Ubuntu/Linux systems with optional cleanup including apt autoremove, autoclean, and docker prune.
// The -y flag can be used to auto-approve cleanup operations without prompting.
var UpdateCmd = &cobra.Command{
	Use:     "update",
	Aliases: []string{"update", "u"},
	Short:   "Update the system",
	Long:    `This command updates the system. It supports Ubuntu and WSL Linux systems and logs a message for unsupported systems.`,
	Run: func(cmd *cobra.Command, _args []string) {
		isVerbose := utils.IsVerbose(cmd)
		autoApprove, _ := cmd.Flags().GetBool("yes")
		cleanupTimeout, _ := cmd.Flags().GetInt("cleanup-timeout")
		log.Verbose(isVerbose, "Checking system type...")

		checkCmd := execCommand("uname", "-a")
		output, err := checkCmd.Output()
		if err != nil {
			log.Error("Error checking system type: %s", err)
			return
		}

		uname := strings.ToLower(string(output))
		log.Verbose(isVerbose, "System type detected: %s", strings.TrimSpace(string(output)))

		if strings.Contains(uname, "ubuntu") || strings.Contains(uname, "linux") {
			log.Verbose(isVerbose, "Detected Ubuntu/Linux system, running system update...")
			updateUbuntu(isVerbose, autoApprove, cleanupTimeout)
		} else if strings.Contains(uname, "darwin") {
			log.Verbose(isVerbose, "Detected macOS system, running macOS update...")
			updateMacOS(isVerbose)
		} else if strings.Contains(uname, "raspberrypi") || strings.Contains(uname, "raspbian") {
			log.Verbose(isVerbose, "Detected Raspberry Pi system, running Raspberry Pi update...")
			updateRaspberryPi(isVerbose)
		} else {
			log.Warn("This system is not yet supported for updates.")
			log.Verbose(isVerbose, "Unsupported system type: %s", strings.TrimSpace(string(output)))
		}
	},
}

// BrewCmd represents the Homebrew update subcommand.
// It provides functionality to update only Homebrew packages, skipping system updates.
// This is useful when you want to update Homebrew packages independently of system packages.
var BrewCmd = &cobra.Command{
	Use:   "brew",
	Short: "Update Homebrew packages only",
	Long:  `This command updates only Homebrew packages, skipping system updates.`,
	Run: func(cmd *cobra.Command, _args []string) {
		isVerbose := utils.IsVerbose(cmd)
		updateBrew(isVerbose)
	},
}

func init() {
	UpdateCmd.Flags().BoolP("yes", "y", false, "Auto-approve cleanup operations without prompting")
	UpdateCmd.Flags().Int("cleanup-timeout", 60, "Timeout in seconds for cleanup confirmation prompt")
	UpdateCmd.AddCommand(BrewCmd)
}

// updateUbuntu performs system updates for Ubuntu and Linux systems.
// It runs apt-get update and upgrade commands, then optionally performs cleanup operations.
// If autoApprove is true, cleanup operations run automatically without prompting.
// If autoApprove is false, the user is prompted to confirm cleanup operations with a countdown.
// cleanupTimeout specifies the seconds to wait before auto-approving cleanup.
func updateUbuntu(isVerbose, autoApprove bool, cleanupTimeout int) {
	log.Message("Running system update for Ubuntu/Linux...")
	log.Message("About to run a command with sudo. You may be prompted for your system password.")

	log.Verbose(isVerbose, "Running: sudo apt-get update && sudo apt-get upgrade -y")
	updateCmd := execCommand("bash", "-c", "sudo apt-get update && sudo apt-get upgrade -y")
	updateCmd.Stdout = log.Writer()
	updateCmd.Stderr = log.ErrorWriter()
	if err := updateCmd.Run(); err != nil {
		log.Error("Error updating system: %s", err)
		log.Verbose(isVerbose, "Update command failed with error: %v", err)
		return
	}
	log.Success("System updated successfully.")
	log.Verbose(isVerbose, "APT update and upgrade completed successfully")

	// Run cleanup operations
	runCleanup(isVerbose, autoApprove, cleanupTimeout)

	updateBrew(isVerbose)
	updateAsdf(isVerbose)
}

// updateMacOS is a placeholder function for macOS system updates.
// Currently displays a message indicating the feature is coming soon and falls back to Homebrew updates.
func updateMacOS(isVerbose bool) {
	log.Message("System update for macOS is coming soon.")
	log.Verbose(isVerbose, "macOS system update functionality not yet implemented")
	updateBrew(isVerbose)
	updateAsdf(isVerbose)
}

// updateRaspberryPi is a placeholder function for Raspberry Pi system updates.
// Currently displays a message indicating the feature is coming soon and falls back to Homebrew updates.
func updateRaspberryPi(isVerbose bool) {
	log.Message("System update for Raspberry Pi is coming soon.")
	log.Verbose(isVerbose, "Raspberry Pi system update functionality not yet implemented")
	updateBrew(isVerbose)
	updateAsdf(isVerbose)
}

// updateBrew updates Homebrew packages on macOS and Linux systems.
// It checks if Homebrew is installed, and if so, runs brew update, brew outdated, and brew upgrade commands.
// If Homebrew is not found, it displays a message and returns without error.
func updateBrew(isVerbose bool) {
	_, err := lookPath("brew")
	if err != nil {
		log.Message("Homebrew (brew) is not installed on this system.")
		log.Verbose(isVerbose, "Could not find brew executable in PATH")
		return
	}

	log.Message("Running Homebrew update and upgrade...")
	log.Verbose(isVerbose, "Running: brew update && brew outdated && brew upgrade")

	updateCmd := execCommand("bash", "-c", "brew update && brew outdated && brew upgrade")
	updateCmd.Stdout = log.Writer()
	updateCmd.Stderr = log.ErrorWriter()
	if err := updateCmd.Run(); err != nil {
		log.Error("Error updating Homebrew packages: %s", err)
		log.Verbose(isVerbose, "Homebrew update failed with error: %v", err)
	} else {
		log.Success("Homebrew packages updated successfully.")
		log.Verbose(isVerbose, "Homebrew update and upgrade completed successfully")
	}
}

// updateAsdf updates asdf version manager plugins.
// It checks if asdf is installed, and if so, runs asdf plugin update --all to update all plugins.
// If asdf is not found, it displays a message and returns without error.
// This works on all platforms where asdf is installed.
func updateAsdf(isVerbose bool) {
	_, err := lookPath("asdf")
	if err != nil {
		log.Message("asdf version manager is not installed on this system.")
		log.Verbose(isVerbose, "Could not find asdf executable in PATH")
		return
	}

	log.Message("Running asdf plugin updates...")
	log.Verbose(isVerbose, "Running: asdf plugin update --all")

	updateCmd := execCommand("asdf", "plugin", "update", "--all")
	updateCmd.Stdout = log.Writer()
	updateCmd.Stderr = log.ErrorWriter()
	if err := updateCmd.Run(); err != nil {
		log.Error("Error updating asdf plugins: %s", err)
		log.Verbose(isVerbose, "asdf plugin update failed with error: %v", err)
	} else {
		log.Success("asdf plugins updated successfully.")
		log.Verbose(isVerbose, "asdf plugin update --all completed successfully")
	}
}

// runCleanup performs system cleanup operations for Ubuntu/Linux systems.
// It runs apt autoremove --purge, apt autoclean, and optionally docker system prune.
// If autoApprove is true, cleanup runs automatically without prompting.
// If autoApprove is false, the user is prompted with a multi-select survey that auto-selects all after cleanupTimeout seconds.
// Docker system prune is only executed if Docker is installed on the system.
func runCleanup(isVerbose, autoApprove bool, cleanupTimeout int) {
	// Define available cleanup operations
	operations := []string{
		"apt autoremove --purge",
		"apt autoclean",
	}

	// Check if docker is available
	_, dockerErr := lookPath("docker")
	if dockerErr == nil {
		operations = append(operations, "docker system prune")
	}

	var selectedOperations []string
	if autoApprove {
		selectedOperations = operations
	} else {
		// Show initial message
		log.Message("Select cleanup operations to run (auto-select all in %d seconds):", cleanupTimeout)

		// Use survey multi-select with timeout
		prompt := &survey.MultiSelect{
			Message: "Select cleanup operations to run:",
			Options: operations,
			Default: operations, // Pre-select all
		}

		// Channel to receive survey result
		resultCh := make(chan []string, 1)
		errorCh := make(chan error, 1)

		// Run survey in goroutine
		go func() {
			var result []string
			err := survey.AskOne(prompt, &result)
			if err != nil {
				errorCh <- err
			} else {
				resultCh <- result
			}
		}()

		// Wait for result or timeout
		select {
		case selected := <-resultCh:
			selectedOperations = selected
		case err := <-errorCh:
			log.Error("Error with survey prompt: %v", err)
			selectedOperations = operations // Default to all on error
		case <-time.After(time.Duration(cleanupTimeout) * time.Second):
			log.Message("Timeout reached, running all cleanup operations...")
			selectedOperations = operations
		}
	}

	if len(selectedOperations) == 0 {
		log.Message("No cleanup operations selected.")
		return
	}

	log.Message("Running selected system cleanup operations...")

	// Run selected operations with progress bars
	for _, operation := range selectedOperations {
		switch operation {
		case "apt autoremove --purge":
			runCleanupOperation(isVerbose, "sudo apt autoremove --purge -y", "apt autoremove")
		case "apt autoclean":
			runCleanupOperation(isVerbose, "sudo apt autoclean", "apt autoclean")
		case "docker system prune":
			runCleanupOperation(isVerbose, "docker system prune -f", "docker system prune")
		}
	}

	log.Success("System cleanup completed.")
}

// runCleanupOperation runs a single cleanup operation with a progress bar
func runCleanupOperation(isVerbose bool, command, operationName string) {
	log.Verbose(isVerbose, "Running: %s", command)

	// Create progress bar for this operation
	progress := utils.NewProgressSpinner(fmt.Sprintf("Running %s...", operationName))

	cleanupCmd := execCommand("bash", "-c", command)
	cleanupCmd.Stdout = log.Writer()
	cleanupCmd.Stderr = log.ErrorWriter()

	if err := cleanupCmd.Run(); err != nil {
		progress.Stop()
		log.Error("Error running %s: %s", operationName, err)
		log.Verbose(isVerbose, "%s failed with error: %v", operationName, err)
	} else {
		progress.SetProgressBar(1.0, fmt.Sprintf("%s completed", operationName))
		progress.Stop()
		log.Success("%s completed.", operationName)
		log.Verbose(isVerbose, "%s completed successfully", operationName)
	}
}
