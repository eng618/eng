package repair

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/cmd/parable_bloom/common"
	"github.com/eng618/eng/cmd/parable_bloom/generate"
	"github.com/eng618/eng/internal/utils"
	"github.com/eng618/eng/internal/utils/log"
)

// LevelRepairCmd repairs corrupted or truncated level files by regenerating them.
var LevelRepairCmd = &cobra.Command{
	Use:   "level-repair",
	Short: "Repair corrupted level JSON files by regenerating them",
	Long: `Scan a levels directory and regenerate any files that fail to parse.
This helps recover from partial writes or corrupted files produced by earlier runs.`,
	Run: func(cmd *cobra.Command, _args []string) {
		isVerbose := utils.IsVerbose(cmd)
		log.Start("Repairing level files (scan + regenerate)")

		directory, _ := cmd.Flags().GetString("directory")
		overwrite, _ := cmd.Flags().GetBool("overwrite")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if directory == "" {
			directory = "assets/levels"
		}

		repairDirectory(directory, overwrite, dryRun, isVerbose)
	},
}

func init() {
	LevelRepairCmd.Flags().
		StringP("directory", "d", "", "Directory containing level files to repair (default: assets/levels)")
	LevelRepairCmd.Flags().BoolP("overwrite", "o", true, "Overwrite repaired files")
	LevelRepairCmd.Flags().BoolP("dry-run", "n", false, "Scan and report without writing files")
}

var levelFileRE = regexp.MustCompile(`^level_(\d+)\.json$`)

func repairDirectory(dir string, overwrite, dryRun, verbose bool) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Error("Failed to read directory %s: %v", dir, err)
		os.Exit(1)
	}

	fixed := 0
	failed := 0
	checked := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		m := levelFileRE.FindStringSubmatch(name)
		if m == nil {
			continue
		}
		checked++
		path := filepath.Join(dir, name)
		if verbose {
			log.Verbose(verbose, "Checking %s", path)
		}

		if repaired, err := repairFileIfNeeded(path, m[1], overwrite, dryRun, verbose); repaired {
			if err != nil {
				failed++
			} else {
				fixed++
			}
		}
	}

	log.Info("Repair summary: checked=%d repaired=%d failed=%d", checked, fixed, failed)
	if failed > 0 {
		os.Exit(1)
	}
}

// repairFileIfNeeded checks a single file and regenerates if parsing fails.
func repairFileIfNeeded(path, idStr string, overwrite, dryRun, verbose bool) (bool, error) {
	_, err := common.ReadLevel(path)
	if err == nil {
		return false, nil
	}

	log.Warn("Failed to parse %s: %v (scheduling regenerate)", path, err)
	id, _ := strconv.Atoi(idStr)
	level := generate.GenerateLevel(
		id,
		fmt.Sprintf("Level %d", id),
		common.DifficultyForLevel(id, nil),
		0,
		0,
		verbose,
		0,
		false,
	)

	if dryRun {
		log.Info("Would regenerate level %d -> %s", id, path)
		return true, nil
	}

	err = common.WriteLevel(path, level, overwrite)
	if err != nil {
		log.Error("Failed to write regenerated level %d to %s: %v", id, path, err)
		return true, err
	}

	log.Info("Repaired level %d", id)
	return true, nil
}
