package system

import (
	"os"
	"os/exec"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

var (
	execCommand = exec.Command // This declaration is kept in system.go as per the instruction's implication.
	lookPath    = exec.LookPath
	userHomeDir = os.UserHomeDir
	stat        = os.Stat
	// askOne is a wrapper for survey.AskOne to allow mocking prompts in tests.
	askOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
		return survey.AskOne(p, response, opts...)
	}
)

var SystemCmd = &cobra.Command{
	Use:   "system",
	Short: "A command for managing the system",
	Long:  `This command will help manage various aspects of MacOS and Linux systems.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		cobra.CheckErr(err)
	},
}

func init() {
	SystemCmd.AddCommand(KillPortCmd)
	SystemCmd.AddCommand(UpdateCmd)
	SystemCmd.AddCommand(ProxyCmd)
	SystemCmd.AddCommand(CompauditFixCmd)
	SystemCmd.AddCommand(SetupCmd)

	// Add flags for subcommands if needed
}
