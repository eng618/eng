package system

import (
	"os/exec"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/log"
	"github.com/spf13/cobra"
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
	Run: func(cmd *cobra.Command, args []string) {
		isVerbose := utils.IsVerbose(cmd)
		autoApprove, _ := cmd.Flags().GetBool("yes")
		log.Verbose(isVerbose, "Checking system type...")

		checkCmd := exec.Command("uname", "-a")
		output, err := checkCmd.Output()
		if err != nil {
			log.Error("Error checking system type: %s", err)
			return
		}

		uname := strings.ToLower(string(output))
		log.Verbose(isVerbose, "System type detected: %s", strings.TrimSpace(string(output)))

		if strings.Contains(uname, "ubuntu") || strings.Contains(uname, "linux") {
			log.Verbose(isVerbose, "Detected Ubuntu/Linux system, running system update...")
			updateUbuntu(isVerbose, autoApprove)
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
	Run: func(cmd *cobra.Command, args []string) {
		isVerbose := utils.IsVerbose(cmd)
		updateBrew(isVerbose)
	},
}

func init() {
	UpdateCmd.Flags().BoolP("yes", "y", false, "Auto-approve cleanup operations without prompting")
	UpdateCmd.AddCommand(BrewCmd)
}

// updateUbuntu performs system updates for Ubuntu and Linux systems.
// It runs apt-get update and upgrade commands, then optionally performs cleanup operations.
// If autoApprove is true, cleanup operations run automatically without prompting.
// If autoApprove is false, the user is prompted to confirm cleanup operations.
func updateUbuntu(isVerbose bool, autoApprove bool) {
	log.Message("Running system update for Ubuntu/Linux...")
	log.Message("About to run a command with sudo. You may be prompted for your system password.")

	log.Verbose(isVerbose, "Running: sudo apt-get update && sudo apt-get upgrade -y")
	updateCmd := exec.Command("bash", "-c", "sudo apt-get update && sudo apt-get upgrade -y")
	updateCmd.Stdout = log.Writer()
	updateCmd.Stderr = log.ErrorWriter()
	if err := updateCmd.Run(); err != nil {
		log.Error("Error updating system: %s", err)
		log.Verbose(isVerbose, "Update command failed with error: %v", err)
		return
	} else {
		log.Success("System updated successfully.")
		log.Verbose(isVerbose, "APT update and upgrade completed successfully")
	}

	// Run cleanup operations
	runCleanup(isVerbose, autoApprove)

	updateBrew(isVerbose)
}

// updateMacOS is a placeholder function for macOS system updates.
// Currently displays a message indicating the feature is coming soon and falls back to Homebrew updates.
func updateMacOS(isVerbose bool) {
	log.Message("System update for macOS is coming soon.")
	log.Verbose(isVerbose, "macOS system update functionality not yet implemented")
	updateBrew(isVerbose)
}

// updateRaspberryPi is a placeholder function for Raspberry Pi system updates.
// Currently displays a message indicating the feature is coming soon and falls back to Homebrew updates.
func updateRaspberryPi(isVerbose bool) {
	log.Message("System update for Raspberry Pi is coming soon.")
	log.Verbose(isVerbose, "Raspberry Pi system update functionality not yet implemented")
	updateBrew(isVerbose)
}

// updateBrew updates Homebrew packages on macOS and Linux systems.
// It checks if Homebrew is installed, and if so, runs brew update, brew outdated, and brew upgrade commands.
// If Homebrew is not found, it displays a message and returns without error.
func updateBrew(isVerbose bool) {
	_, err := exec.LookPath("brew")
	if err != nil {
		log.Message("Homebrew (brew) is not installed on this system.")
		log.Verbose(isVerbose, "Could not find brew executable in PATH")
		return
	}

	log.Message("Running Homebrew update and upgrade...")
	log.Verbose(isVerbose, "Running: brew update && brew outdated && brew upgrade")

	updateCmd := exec.Command("bash", "-c", "brew update && brew outdated && brew upgrade")
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

// runCleanup performs system cleanup operations for Ubuntu/Linux systems.
// It runs apt autoremove --purge, apt autoclean, and optionally docker system prune.
// If autoApprove is true, cleanup runs automatically without prompting.
// If autoApprove is false, the user is prompted to confirm running cleanup operations.
// Docker system prune is only executed if Docker is installed on the system.
func runCleanup(isVerbose bool, autoApprove bool) {
	var runCleanup bool
	if autoApprove {
		runCleanup = true
	} else {
		prompt := &survey.Confirm{
			Message: "Run system cleanup operations (autoremove, autoclean, docker prune)?",
			Default: true,
		}
		err := survey.AskOne(prompt, &runCleanup)
		if err != nil {
			log.Error("Error getting user confirmation: %s", err)
			return
		}
	}

	if !runCleanup {
		log.Message("Skipping cleanup operations.")
		return
	}

	log.Message("Running system cleanup operations...")

	// Run apt autoremove --purge
	log.Verbose(isVerbose, "Running: sudo apt autoremove --purge -y")
	cleanupCmd := exec.Command("bash", "-c", "sudo apt autoremove --purge -y")
	cleanupCmd.Stdout = log.Writer()
	cleanupCmd.Stderr = log.ErrorWriter()
	if err := cleanupCmd.Run(); err != nil {
		log.Error("Error running apt autoremove: %s", err)
		log.Verbose(isVerbose, "apt autoremove failed with error: %v", err)
	} else {
		log.Success("apt autoremove completed.")
		log.Verbose(isVerbose, "apt autoremove --purge completed successfully")
	}

	// Run apt autoclean
	log.Verbose(isVerbose, "Running: sudo apt autoclean")
	cleanupCmd = exec.Command("bash", "-c", "sudo apt autoclean")
	cleanupCmd.Stdout = log.Writer()
	cleanupCmd.Stderr = log.ErrorWriter()
	if err := cleanupCmd.Run(); err != nil {
		log.Error("Error running apt autoclean: %s", err)
		log.Verbose(isVerbose, "apt autoclean failed with error: %v", err)
	} else {
		log.Success("apt autoclean completed.")
		log.Verbose(isVerbose, "apt autoclean completed successfully")
	}

	// Check if docker is available and run docker system prune
	_, err := exec.LookPath("docker")
	if err != nil {
		log.Message("Docker is not installed on this system, skipping docker system prune.")
		log.Verbose(isVerbose, "Could not find docker executable in PATH")
	} else {
		log.Verbose(isVerbose, "Running: docker system prune -f")
		cleanupCmd = exec.Command("bash", "-c", "docker system prune -f")
		cleanupCmd.Stdout = log.Writer()
		cleanupCmd.Stderr = log.ErrorWriter()
		if err := cleanupCmd.Run(); err != nil {
			log.Error("Error running docker system prune: %s", err)
			log.Verbose(isVerbose, "docker system prune failed with error: %v", err)
		} else {
			log.Success("docker system prune completed.")
			log.Verbose(isVerbose, "docker system prune completed successfully")
		}
	}

	log.Success("System cleanup completed.")
}
