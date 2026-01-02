package parable_bloom

import (
	"github.com/spf13/cobra"

	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/log"
)

// LevelGenerateCmd represents the 'parable-bloom level-generate' command for generating new game levels.
// It provides tools to scaffold and generate new level files for the Parable Bloom game.
var LevelGenerateCmd = &cobra.Command{
	Use:   "level-generate",
	Short: "Generate a new game level",
	Long: `Generate a new game level for the Parable Bloom project.
This command scaffolds a new level file with the required structure and metadata.`,
	Run: func(cmd *cobra.Command, args []string) {
		isVerbose := utils.IsVerbose(cmd)
		log.Start("Creating new game level")
		log.Verbose(isVerbose, "Level creation initiated")

		// TODO: Implement level creation logic
		log.Info("Level creation functionality coming soon")
	},
}

func init() {
	// Add flags for level generation
	LevelGenerateCmd.Flags().StringP("name", "n", "", "Name of the new level (required)")
	LevelGenerateCmd.Flags().StringP("module", "m", "", "Module the level belongs to (e.g., tutorial, garden)")
	LevelGenerateCmd.Flags().IntP("grid-width", "w", 4, "Width of the game grid")
	LevelGenerateCmd.Flags().IntP("grid-height", "H", 4, "Height of the game grid")
	LevelGenerateCmd.Flags().StringP("output", "o", "", "Output directory for the level file")
}
