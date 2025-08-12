// Package codemod provides helpers for codemods and project automation.
package codemod

import (
	"os/exec"

	"github.com/spf13/cobra"
)

// execCommand is a variable holding the exec.Command function, allowing for test overrides.
var execCommand = exec.Command

// CodemodCmd is the root command for codemod-related helpers and automation.
var CodemodCmd = &cobra.Command{
	Use:   "codemod",
	Short: "Helpers for codemods and project automation",
	Long:  `Run codemods or setup helpers for various project types.`,
}

func init() {
	CodemodCmd.AddCommand(LintSetupCmd)
	CodemodCmd.AddCommand(CopilotSetupCmd)
}
