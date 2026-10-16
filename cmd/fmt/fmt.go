// Package fmtcmd checks and formats web code with Oxfmt using the shared
// Garcia Ventures style (@gv-tech/oxc-config), without requiring a local install.
package fmtcmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/oxfmt"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

var (
	fmtWrite  bool
	fmtCheck  bool
	fmtDiff   bool
	fmtRunner string
	fmtConfig string
)

// FmtCmd checks or formats a file or directory with Oxfmt using the
// @gv-tech/oxc-config style.
var FmtCmd = &cobra.Command{
	Use:   "fmt [path]",
	Short: "Check or format code with Oxfmt using @gv-tech/oxc-config style",
	Long: `Check or format a file or directory with Oxfmt using the shared
@gv-tech/oxc-config style, without requiring a local install.

By default (and with --check) it verifies formatting: it lists files that
would change and exits non-zero when the workspace is dirty, without
modifying anything. Pass --write to format files in place, or --diff to
also show a colored unified-style diff of the pending changes.

The formatter runs through the gv-oxfmt wrapper (bun in bun workspaces,
npx otherwise) so the shared style applies even without a local install. A
local oxfmt binary is only used as a last resort, or when forced with
--runner oxfmt. Use --config to format with a custom oxfmt config instead.

This formats JS/TS and web assets (the Oxfmt domain); Go code is formatted
with gofmt via 'task format'.`,
	Example: `  eng fmt
  eng fmt ./web
  eng fmt src/index.ts --diff
  eng fmt --write .
  eng fmt --write src --runner bun
  eng fmt --check .`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary).
			MarginBottom(1)
		if !ui.DisableProgress {
			fmt.Fprintln(log.Out, headerStyle.Render("✨ Oxfmt Code Formatter"))
		}

		target := "."
		if len(args) > 0 {
			target = args[0]
		}
		opts := oxfmt.Options{
			Path:       target,
			Write:      fmtWrite,
			Check:      fmtCheck,
			Diff:       fmtDiff,
			Runner:     fmtRunner,
			ConfigPath: fmtConfig,
			IsVerbose:  cmdutil.IsVerbose(cmd),
		}
		ctx := cmd.Context()
		if ctx == nil {
			ctx = cmdutil.FallbackContext()
		}
		return oxfmt.Run(ctx, opts, nil)
	},
}

func init() {
	FmtCmd.Flags().BoolVarP(&fmtWrite, "write", "w", false, "Format files in place (default is check mode)")
	FmtCmd.Flags().
		BoolVar(&fmtCheck, "check", false, "Verify formatting without modifying files (default behavior)")
	FmtCmd.Flags().BoolVar(&fmtDiff, "diff", false, "Show a colored diff of pending changes (implies check mode)")
	FmtCmd.Flags().StringVar(&fmtRunner, "runner", oxfmt.RunnerAuto, "JS runner: auto, npx, bun, oxfmt")
	FmtCmd.Flags().
		StringVarP(&fmtConfig, "config", "c", "", "Custom oxfmt config file (default uses @gv-tech/oxc-config style)")
}
