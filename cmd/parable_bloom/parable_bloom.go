// Package parable_bloom provides cobra commands for managing the Parable Bloom Flutter game project.
//
// DEPRECATED: This package is deprecated and will be removed in a future release.
// All functionality has been migrated to the standalone level-builder CLI tool located at:
// parable-bloom/tools/level-builder
//
// Please use the new level-builder tool instead:
//
//	cd parable-bloom/tools/level-builder
//	go build
//	./level-builder --help
package parable_bloom

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/cmd/parable_bloom/generate"
	"github.com/eng618/eng/cmd/parable_bloom/render"
	"github.com/eng618/eng/cmd/parable_bloom/repair"
	"github.com/eng618/eng/cmd/parable_bloom/validate"
)

// ParableBloomCmd serves as the base command for all Parable Bloom project operations.
// NOTE: This command is DEPRECATED and hidden. See the top-level documentation and
// the standalone `tools/level-builder` CLI for the canonical implementation.
var ParableBloomCmd = &cobra.Command{
	Use:     "parable-bloom",
	Short:   "Manage Parable Bloom Flutter game project (DEPRECATED)",
	Long:    `DEPRECATED: This command is deprecated and will be removed. Use the standalone level-builder tool at parable-bloom/tools/level-builder.`,
	Aliases: []string{"pb"},
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		// Short, friendly deprecation message for users who still invoke the command.
		fmt.Println("DEPRECATED: 'eng parable-bloom' is deprecated and will be removed.")
		fmt.Println("Please use the standalone tool:")
		fmt.Println("  cd parable-bloom/tools/level-builder && go build && ./level-builder --help")

		// Show help for convenience if no args were supplied.
		_ = cmd.Help()
	},
}

func init() {
	ParableBloomCmd.AddCommand(generate.LevelGenerateCmd)
	ParableBloomCmd.AddCommand(validate.LevelValidateCmd)
	ParableBloomCmd.AddCommand(render.LevelRenderCmd)
	ParableBloomCmd.AddCommand(repair.LevelRepairCmd)
}
