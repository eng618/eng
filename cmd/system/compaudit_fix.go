package system

import (
	"os/exec"
	"strings"

	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/log"
	"github.com/spf13/cobra"
)

// CompauditFixCmd runs 'compaudit' and removes group/world write
// permissions from any reported insecure directories.
var CompauditFixCmd = &cobra.Command{
	Use:   "compauditFix",
	Short: "Fix insecure directories reported by compaudit",
	Long:  `Runs 'compaudit' and applies 'chmod g-w,o-w' to any directories reported as insecure. This uses an interactive zsh so zsh functions like compaudit (from oh-my-zsh) are available.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Running compaudit and fixing insecure directories...")
		isVerbose := utils.IsVerbose(cmd)

		// Use an interactive zsh (-i) and run a command (-c) so compaudit (a zsh function) is available.
		execCmd := exec.Command("zsh", "-ic", "compaudit | xargs --no-run-if-empty chmod g-w,o-w")
		log.Verbose(isVerbose, "Executing: %s", execCmd.String())

		outputBytes, err := execCmd.CombinedOutput()
		output := strings.TrimSpace(string(outputBytes))

		if err != nil {
			log.Error("Failed to run compaudit fix: %v", err)
			if output != "" {
				log.Error("Output: %s", output)
			}
			return
		}

		if output != "" {
			// Print any output from the command for visibility.
			log.Message("%s", output)
		}

		log.Success("Finished applying compaudit fixes.")
	},
}
