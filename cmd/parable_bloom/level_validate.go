package parable_bloom

import (
	"github.com/spf13/cobra"

	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/log"
)

// LevelValidateCmd represents the 'parable-bloom level-validate' command for validating game levels.
// It checks that levels are properly formatted and solvable using the game's solver.
var LevelValidateCmd = &cobra.Command{
	Use:   "level-validate",
	Short: "Validate game levels for solvability",
	Long: `Validate game levels to ensure they are properly formatted and solvable.
This command uses the Parable Bloom level solver to verify that levels can be completed.`,
	Run: func(cmd *cobra.Command, args []string) {
		isVerbose := utils.IsVerbose(cmd)
		log.Start("Validating game levels")
		log.Verbose(isVerbose, "Level validation initiated")

		// TODO: Implement level validation logic
		log.Info("Level validation functionality coming soon")
	},
}

func init() {
	// Add flags for level validation
	LevelValidateCmd.Flags().StringP("file", "f", "", "Path to a specific level file to validate")
	LevelValidateCmd.Flags().StringP("directory", "d", "", "Directory containing level files to validate")
	LevelValidateCmd.Flags().BoolP("check-solvability", "s", true, "Check if levels are solvable (default: true)")
	LevelValidateCmd.Flags().BoolP("strict", "S", false, "Enable strict validation mode")
}
