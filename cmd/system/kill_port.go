package system

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/log"
)

var KillPortCmd = &cobra.Command{
	Use:   "killPort",
	Short: "Kill a supplied port",
	Long:  `This will find what process is running on a supplied port, then kill that process.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			KillString := fmt.Sprintf("kill -9 $(lsof -ti:%s)", strings.Join(args, ","))
			log.Message("This should run: %s", KillString)

			killPortCmd := exec.Command(KillString)
			utils.StartChildProcess(killPortCmd)
		} else {
			log.Warn("You need to supply the port to kill, which will run the following:")
		}
	},
}
