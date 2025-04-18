package system

import (
	"os/exec"
	"strings"

	"github.com/eng618/eng/utils/log"
	"github.com/spf13/cobra"
)

var UpdateCmd = &cobra.Command{
	Use:     "update",
	Aliases: []string{"update", "u"},
	Short:   "Update the system",
	Long:    `This command updates the system. It supports Ubuntu systems and logs a message for unsupported systems.`,
	Run: func(cmd *cobra.Command, args []string) {
		checkCmd := exec.Command("uname", "-a")
		output, err := checkCmd.Output()
		if err != nil {
			log.Error("Error checking system type: %s", err)
			return
		}

		uname := strings.ToLower(string(output))
		if strings.Contains(uname, "ubuntu") {
			updateUbuntu()
		} else if strings.Contains(uname, "darwin") {
			updateMacOS()
		} else if strings.Contains(uname, "raspberrypi") || strings.Contains(uname, "raspbian") {
			updateRaspberryPi()
		} else {
			log.Warn("This system is not yet supported for updates.")
		}
	},
}

func updateUbuntu() {
	log.Message("Running system update for Ubuntu...")
	log.Message("About to run a command with sudo. You may be prompted for your system password.")
	updateCmd := exec.Command("sudo", "apt", "update", "&&", "sudo", "apt", "upgrade", "-y")
	if err := updateCmd.Run(); err != nil {
		log.Error("Error updating system: %s", err)
	} else {
		log.Message("System updated successfully.")
	}
	updateBrew()
}

func updateMacOS() {
	log.Message("System update for macOS is coming soon.")
	updateBrew()
}

func updateRaspberryPi() {
	log.Message("System update for Raspberry Pi is coming soon.")
	updateBrew()
}

func updateBrew() {
	_, err := exec.LookPath("brew")
	if err != nil {
		log.Message("Homebrew (brew) is not installed on this system.")
		return
	}
	log.Message("Running Homebrew update and upgrade...")
	updateCmd := exec.Command("bash", "-c", "brew update && brew outdated && brew upgrade")
	updateCmd.Stdout = log.Writer()
	updateCmd.Stderr = log.ErrorWriter()
	if err := updateCmd.Run(); err != nil {
		log.Error("Error updating Homebrew packages: %s", err)
	} else {
		log.Message("Homebrew packages updated successfully.")
	}
}
