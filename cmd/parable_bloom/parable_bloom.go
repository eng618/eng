// Package parable_bloom provides cobra commands for managing the Parable Bloom Flutter game project.
//
// DEPRECATED: This package is deprecated and will be removed in a future release.
// All functionality has been migrated to the standalone level-builder CLI tool located at:
// parable-bloom/tools/level-builder
//
// Please use the new level-builder tool instead:
//   cd parable-bloom/tools/level-builder
//   go build
//   ./level-builder --help
package parable_bloom

import (
	"github.com/spf13/cobra"

	"github.com/eng618/eng/cmd/parable_bloom/generate"
	"github.com/eng618/eng/cmd/parable_bloom/render"
	"github.com/eng618/eng/cmd/parable_bloom/repair"
	"github.com/eng618/eng/cmd/parable_bloom/validate"
	"github.com/eng618/eng/internal/utils"
	"github.com/eng618/eng/internal/utils/log"
)

// ParableBloomCmd serves as the base command for all Parable Bloom project operations.
// It doesn't perform any action itself but groups subcommands like level generation and validation.
var ParableBloomCmd = &cobra.Command{
	Use:     "parable-bloom",
	Short:   "Manage Parable Bloom Flutter game project",
	Long:    `This command is used to facilitate the management and development of the Parable Bloom Flutter game project.`,
	Aliases: []string{"pb"},
	Run: func(cmd *cobra.Command, args []string) {
		isVerbose := utils.IsVerbose(cmd)

		// If no subcommand is given, print the help information.
		if len(args) == 0 {
			log.Verbose(isVerbose, "No subcommand provided, showing help.")
			err := cmd.Help()
			cobra.CheckErr(err)
		} else {
			log.Verbose(isVerbose, "Subcommand '%s' provided.", args[0])
		}
	},
}

func init() {
	ParableBloomCmd.AddCommand(generate.LevelGenerateCmd)
	ParableBloomCmd.AddCommand(validate.LevelValidateCmd)
	ParableBloomCmd.AddCommand(render.LevelRenderCmd)
	ParableBloomCmd.AddCommand(repair.LevelRepairCmd)
}
