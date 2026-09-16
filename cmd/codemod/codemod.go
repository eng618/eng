// Package codemod provides helpers for codemods and project automation.
package codemod

import (
	"embed"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/execx"
)

//go:embed assets/*
var AssetsFS embed.FS

// execCommand is a variable holding the exec.Command function, allowing for test overrides.
var execCommand = execx.Command

// CodemodCmd is the root command for codemod-related helpers and automation.
var CodemodCmd = &cobra.Command{
	Use:   "codemod",
	Short: "Helpers for codemods and project automation",
	Long:  `Run codemods or setup helpers for various project types.`,
}

var echo bool

func init() {
	CodemodCmd.AddCommand(LintSetupCmd)
	CodemodCmd.AddCommand(OxcSetupCmd)
	CodemodCmd.AddCommand(CopilotSetupCmd)
	CodemodCmd.AddCommand(PrettierCmd)
	CodemodCmd.AddCommand(NativeCmd)
	CodemodCmd.AddCommand(WebCmd)
	LintSetupCmd.Flags().BoolVarP(&echo, "echo", "e", false, "Use echo linting setup")
	OxcSetupCmd.Flags().
		StringVar(&oxcPreset, "preset", "auto", "Oxc preset: auto, recommended, next, vite, react, typescript, base")
	OxcSetupCmd.Flags().BoolVar(&oxcTypeAware, "type-aware", false, "Enable type-aware linting (oxlint-tsgolint)")
	OxcSetupCmd.Flags().
		BoolVar(&oxcRemoveEslint, "remove-eslint", false, "Remove ESLint/Prettier stack and legacy configs")
	OxcSetupCmd.Flags().BoolVarP(&oxcYes, "yes", "y", false, "Skip confirmation prompts")
}
