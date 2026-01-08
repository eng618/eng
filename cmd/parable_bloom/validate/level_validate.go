package validate

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/cmd/parable_bloom/common"
	"github.com/eng618/eng/internal/utils"
	"github.com/eng618/eng/internal/utils/log"
)

// LevelValidateCmd represents the 'parable-bloom level-validate' command for validating game levels.
// It checks that levels are properly formatted and solvable using the game's solver.
var LevelValidateCmd = &cobra.Command{
	Use:   "level-validate",
	Short: "Validate game levels for solvability",
	Long: `Validate game levels to ensure they are properly formatted and solvable.
This command uses the Parable Bloom level solver to verify that levels can be completed.`,
	Run: func(cmd *cobra.Command, _args []string) {
		isVerbose := utils.IsVerbose(cmd)
		log.Start("Validating game levels")

		file, _ := cmd.Flags().GetString("file")
		directory, _ := cmd.Flags().GetString("directory")
		checkSolvability, _ := cmd.Flags().GetBool("check-solvability")
		strict, _ := cmd.Flags().GetBool("strict")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		showWarnings, _ := cmd.Flags().GetBool("warnings")

		if file == "" && directory == "" {
			directory = "assets/levels"
		}

		if file != "" {
			validateSingleFile(file, checkSolvability, strict, dryRun, showWarnings, isVerbose)
		} else {
			validateDirectory(directory, checkSolvability, strict, dryRun, showWarnings, isVerbose)
		}
	},
}

func init() {
	// Add flags for level validation
	LevelValidateCmd.Flags().StringP("file", "f", "", "Path to a specific level file to validate")
	LevelValidateCmd.Flags().StringP("directory", "d", "", "Directory containing level files to validate (default: assets/levels)")
	LevelValidateCmd.Flags().BoolP("check-solvability", "s", true, "Check if levels are solvable (default: true)")
	LevelValidateCmd.Flags().BoolP("strict", "S", false, "Enable strict validation mode (BFS solver)")
	LevelValidateCmd.Flags().BoolP("dry-run", "", false, "Validate without persisting changes")
	LevelValidateCmd.Flags().BoolP("warnings", "w", false, "Show warnings in addition to violations (default: violations only)")
}

func validateSingleFile(filePath string, checkSolvability, strict, _dryRun, showWarnings, verbose bool) {
	log.Verbose(verbose, "Validating single file: %s", filePath)

	level, err := common.ReadLevel(filePath)
	if err != nil {
		log.Error("Failed to read level: %v", err)
		os.Exit(1)
	}

	result := validateLevel(level, checkSolvability, strict, verbose)
	printValidationResult(result, showWarnings, verbose)

	if !result.Valid {
		os.Exit(1)
	}
}

func validateDirectory(dirPath string, checkSolvability, strict, _dryRun, showWarnings, verbose bool) {
	log.Verbose(verbose, "Validating directory: %s", dirPath)

	levels, err := common.ReadLevelsFromDir(dirPath)
	if err != nil {
		log.Error("Failed to read directory: %v", err)
		os.Exit(1)
	}

	if len(levels) == 0 {
		log.Warn("No level files found in %s", dirPath)
		return
	}

	log.Verbose(verbose, "Found %d level files", len(levels))

	// Validate in parallel
	results := make(chan common.ValidationResult, len(levels))
	var wg sync.WaitGroup

	for _, level := range levels {
		wg.Add(1)
		go func(l *common.Level) {
			defer wg.Done()
			result := validateLevel(l, checkSolvability, strict, verbose)
			results <- result
		}(level)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var validCount int
	var totalViolations int
	var totalWarnings int

	for result := range results {
		printValidationResult(result, showWarnings, verbose)
		if result.Valid {
			validCount++
		}
		totalViolations += len(result.Violations)
		totalWarnings += len(result.Warnings)
	}

	// Summary
	fmt.Printf("\n%s\n", "=============================================")
	fmt.Printf("Summary: %d/%d files valid\n", validCount, len(levels))
	fmt.Printf("Violations: %d, Warnings: %d\n", totalViolations, totalWarnings)
	fmt.Printf("%s\n", "=============================================")

	if validCount < len(levels) {
		os.Exit(1)
	}
}

func validateLevel(level *common.Level, checkSolvability, strict, verbose bool) common.ValidationResult {
	result := common.ValidationResult{
		Filename:  fmt.Sprintf("level_%d.json", level.ID),
		Timestamp: time.Now(),
	}

	violations, warnings := level.Validate()
	result.Violations = violations
	result.Warnings = warnings

	if checkSolvability && len(violations) == 0 {
		log.Verbose(verbose, "Checking solvability for level %d", level.ID)
		solver := common.NewSolver(level)
		solvable := false

		if strict {
			solvable = solver.IsSolvableBFS()
		} else {
			solvable = solver.IsSolvableGreedy()
		}

		if !solvable {
			result.Violations = append(result.Violations, "level is not solvable")
		}
	}

	result.Valid = len(result.Violations) == 0
	return result
}

func printValidationResult(result common.ValidationResult, showWarnings, _verbose bool) {
	if result.Valid {
		fmt.Printf("✓ %s\n", result.Filename)
		return
	}

	status := "✗"
	if len(result.Violations) == 0 && len(result.Warnings) > 0 {
		status = "⚠"
	}

	fmt.Printf("\n%s %s\n", status, result.Filename)

	if len(result.Violations) > 0 {
		fmt.Printf("  VIOLATIONS (errors):\n")
		for _, v := range result.Violations {
			fmt.Printf("    - %s\n", v)
		}
	}

	if showWarnings && len(result.Warnings) > 0 {
		fmt.Printf("  WARNINGS (advisories):\n")
		for _, w := range result.Warnings {
			fmt.Printf("    - %s\n", w)
		}
	}
}
