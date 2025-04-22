package system

import (
	"os/exec"
	"strings"

	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/log"
	"github.com/spf13/cobra"
)

var UpdateCmd = &cobra.Command{
	Use:     "update",
	Aliases: []string{"update", "u"},
	Short:   "Update the system",
	Long:    `This command updates the system. It supports Ubuntu systems and logs a message for unsupported systems.`,
	Run: func(cmd *cobra.Command, args []string) {
		isVerbose := utils.IsVerbose(cmd)
		log.Verbose(isVerbose, "Checking system type...")

		checkCmd := exec.Command("uname", "-a")
		output, err := checkCmd.Output()
		if err != nil {
			log.Error("Error checking system type: %s", err)
			return
		}

		uname := strings.ToLower(string(output))
		log.Verbose(isVerbose, "System type detected: %s", strings.TrimSpace(string(output)))

		if strings.Contains(uname, "ubuntu") {
			log.Verbose(isVerbose, "Detected Ubuntu system, running Ubuntu update...")
			updateUbuntu(isVerbose)
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

func updateUbuntu(isVerbose bool) {
	log.Message("Running system update for Ubuntu...")
	log.Message("About to run a command with sudo. You may be prompted for your system password.")

	log.Verbose(isVerbose, "Running: sudo apt update && sudo apt upgrade -y")
	updateCmd := exec.Command("sudo", "apt", "update", "&&", "sudo", "apt", "upgrade", "-y")
	if err := updateCmd.Run(); err != nil {
		log.Error("Error updating system: %s", err)
		log.Verbose(isVerbose, "Update command failed with error: %v", err)
	} else {
		log.Message("System updated successfully.")
		log.Verbose(isVerbose, "APT update and upgrade completed successfully")
	}
	updateBrew(isVerbose)
}

func updateMacOS(isVerbose bool) {
	log.Message("System update for macOS is coming soon.")
	log.Verbose(isVerbose, "macOS system update functionality not yet implemented")
	updateBrew(isVerbose)
}

func updateRaspberryPi(isVerbose bool) {
	log.Message("System update for Raspberry Pi is coming soon.")
	log.Verbose(isVerbose, "Raspberry Pi system update functionality not yet implemented")
	updateBrew(isVerbose)
}

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
		log.Message("Homebrew packages updated successfully.")
		log.Verbose(isVerbose, "Homebrew update and upgrade completed successfully")
	}
}
