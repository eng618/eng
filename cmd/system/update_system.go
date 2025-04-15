package system

import (
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/eng618/eng/utils/log"
)

var UpdateSystemCmd = &cobra.Command{
	Use:   "updateSystem",
	Short: "Update the system",
	Long:  `This command updates the system. It supports Ubuntu systems and logs a message for unsupported systems.`,
	Run: func(cmd *cobra.Command, args []string) {
		checkCmd := exec.Command("uname", "-a")
		output, err := checkCmd.Output()
		if err != nil {
			log.Error("Error checking system type: %s", err)
			return
		}

		if strings.Contains(strings.ToLower(string(output)), "ubuntu") {
			log.Message("Running system update for Ubuntu...")
			updateCmd := exec.Command("sudo", "apt", "update", "&&", "sudo", "apt", "upgrade", "-y")
			if err := updateCmd.Run(); err != nil {
				log.Error("Error updating system: %s", err)
			} else {
				log.Message("System updated successfully.")
			}
		} else {
			log.Message("This system is not yet supported for updates.")
		}
	},
}
