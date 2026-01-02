package parable_bloom

import (
	"fmt"
	"math/rand"
	"os"
	"sync"

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
This command creates solvable levels with the required structure and metadata.`,
	Run: func(cmd *cobra.Command, args []string) {
		isVerbose := utils.IsVerbose(cmd)
		log.Start("Generating game levels")

		name, _ := cmd.Flags().GetString("name")
		module, _ := cmd.Flags().GetString("module")
		width, _ := cmd.Flags().GetInt("grid-width")
		height, _ := cmd.Flags().GetInt("grid-height")
		output, _ := cmd.Flags().GetString("output")
		stdout, _ := cmd.Flags().GetBool("stdout")
		overwrite, _ := cmd.Flags().GetBool("overwrite")
		count, _ := cmd.Flags().GetInt("count")

		if output == "" {
			output = "assets/levels"
		}

		// Load modules
		modules, err := LoadModules("")
		if err != nil {
			log.Error("Failed to load modules: %v", err)
			os.Exit(1)
		}

		if count > 1 {
			// Batch generation by module
			generateBatch(modules, module, count, output, overwrite, isVerbose)
		} else {
			// Single level generation
			generateSingle(name, module, width, height, output, stdout, overwrite, modules, isVerbose)
		}
	},
}

func init() {
	// Add flags for level generation
	LevelGenerateCmd.Flags().StringP("name", "n", "", "Name of the new level")
	LevelGenerateCmd.Flags().StringP("module", "m", "", "Module/parable (or level ID for single)")
	LevelGenerateCmd.Flags().IntP("grid-width", "w", 0, "Width of the game grid (0 = auto)")
	LevelGenerateCmd.Flags().IntP("grid-height", "H", 0, "Height of the game grid (0 = auto)")
	LevelGenerateCmd.Flags().StringP("output", "o", "", "Output directory for level files (default: assets/levels)")
	LevelGenerateCmd.Flags().BoolP("stdout", "", false, "Print level JSON to stdout instead of file")
	LevelGenerateCmd.Flags().BoolP("overwrite", "", false, "Overwrite existing level files")
	LevelGenerateCmd.Flags().IntP("count", "c", 1, "Generate multiple levels (batch mode)")
	LevelGenerateCmd.Flags().BoolP("dry-run", "", false, "Generate without writing to disk")
}

func generateSingle(name, module string, width, height int, output string, stdout, overwrite bool, modules []ModuleRange, verbose bool) {
	log.Verbose(verbose, "Generating single level")

	// If name is numeric, treat it as level ID
	var levelID int
	var difficulty string

	if name != "" {
		// Try parsing as level ID
		n, _ := fmt.Sscanf(name, "%d", &levelID)
		if n != 1 {
			// It's a name, generate next available ID
			levelID = GenerateLevelID(output, 1)
		}
	} else {
		levelID = GenerateLevelID(output, 1)
	}

	// Determine difficulty
	if module != "" {
		difficulty = module
	} else {
		difficulty = DifficultyForLevel(levelID, modules)
	}

	log.Verbose(verbose, "Level ID: %d, Difficulty: %s", levelID, difficulty)

	// Generate level
	level := generateLevel(levelID, name, difficulty, width, height, verbose)

	// Validate
	violations, warnings := level.Validate()
	if len(violations) > 0 {
		log.Warn("Generated level has violations:")
		for _, v := range violations {
			log.Warn("  - %s", v)
		}
	}
	if len(warnings) > 0 && verbose {
		for _, w := range warnings {
			log.Verbose(verbose, "  ⚠ %s", w)
		}
	}

	// Output
	if stdout {
		outputLevelToStdout(level)
	} else {
		filePath := GetLevelFilePath(level.ID, output)
		err := WriteLevel(filePath, level, overwrite)
		if err != nil {
			log.Error("%v", err)
			os.Exit(1)
		}
		log.Info("Level written to %s", filePath)
	}
}

func generateBatch(modules []ModuleRange, moduleName string, count int, output string, overwrite bool, verbose bool) {
	log.Verbose(verbose, "Generating batch of %d levels for module: %s", count, moduleName)

	// Find module range
	var moduleRange *ModuleRange
	if moduleName != "" {
		moduleRange = GetModuleRange(moduleName, modules)
		if moduleRange == nil {
			// Try as module ID
			var modID int
			fmt.Sscanf(moduleName, "%d", &modID)
			for i := range modules {
				if modules[i].ID == modID {
					moduleRange = &modules[i]
					break
				}
			}
		}
	}

	if moduleRange == nil {
		log.Error("Module not found: %s", moduleName)
		os.Exit(1)
	}

	startID := moduleRange.Start
	endID := startID + count - 1
	if endID > moduleRange.End {
		endID = moduleRange.End
	}

	log.Verbose(verbose, "Generating levels %d to %d", startID, endID)

	// Generate in parallel
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 4) // Limit to 4 concurrent generators

	successCount := 0
	var mu sync.Mutex

	for levelID := startID; levelID <= endID; levelID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			difficulty := DifficultyForLevel(id, modules)
			level := generateLevel(id, fmt.Sprintf("Level %d", id), difficulty, 0, 0, verbose)

			filePath := GetLevelFilePath(level.ID, output)
			err := WriteLevel(filePath, level, overwrite)

			mu.Lock()
			if err == nil {
				log.Info("Generated level %d", id)
				successCount++
			} else {
				log.Error("Failed to generate level %d: %v", id, err)
			}
			mu.Unlock()
		}(levelID)
	}

	wg.Wait()
	log.Info("Batch generation complete: %d/%d levels generated", successCount, (endID - startID + 1))
}

func generateLevel(id int, name, difficulty string, width, height int, verbose bool) *Level {
	if name == "" {
		name = fmt.Sprintf("Level %d", id)
	}

	// Determine grid size
	gridSize := [2]int{width, height}
	if width == 0 || height == 0 {
		gridSize = GridSizeForLevel(id, difficulty)
	}

	log.Verbose(verbose, "Generating level %d with grid %dx%d", id, gridSize[0], gridSize[1])

	// Create level with minimum required fields
	level := &Level{
		ID:         id,
		Name:       name,
		Difficulty: difficulty,
		GridSize:   gridSize,
		Mask: &Mask{
			Mode:   "show-all",
			Points: []any{},
		},
		Vines:      generateVines(gridSize, difficulty, id),
		MaxMoves:   estimateMaxMoves(difficulty),
		MinMoves:   estimateMinMoves(difficulty),
		Complexity: ComplexityForDifficulty(difficulty),
		Grace:      GraceForDifficulty(difficulty),
	}

	return level
}

func generateVines(gridSize [2]int, difficulty string, levelID int) []Vine {
	spec := DifficultySpecs[difficulty]
	rng := rand.New(rand.NewSource(int64(levelID)))

	// Target occupancy
	totalCells := gridSize[0] * gridSize[1]
	targetOccupancy := totalCells * 95 / 100

	var vines []Vine
	occupied := make(map[string]bool)

	// Use a limited palette of colors
	colors := []string{"moss_green", "sunset_orange", "golden_yellow", "royal_purple", "sky_blue"}
	colorIdx := 0

	// Determine target vine length from difficulty spec
	minLen := spec.AvgLengthRange[0]
	maxLen := spec.AvgLengthRange[1]

	// Generate vines with proper lengths until occupancy is reached
	occupiedCount := 0
	vinesCreated := 0
	maxVines := spec.VineCountRange[1]
	consecutiveFailures := 0

	for occupiedCount < targetOccupancy && vinesCreated < maxVines && consecutiveFailures < 10 {
		// Target length varies within the spec range
		targetLength := minLen + rng.Intn(maxLen-minLen+1)

		vine := generateVineWithLength(vinesCreated, gridSize, occupied, rng, targetLength)
		if vine == nil {
			// Try one shorter attempt
			vine = generateVineWithLength(vinesCreated, gridSize, occupied, rng, 3)
			consecutiveFailures++
		} else {
			consecutiveFailures = 0
		}

		if vine != nil {
			// Assign color from limited palette
			vine.VineColor = colors[colorIdx%len(colors)]
			colorIdx++

			// Add vine cells to occupied
			for _, pt := range vine.OrderedPath {
				occupied[fmt.Sprintf("%d,%d", pt.X, pt.Y)] = true
				occupiedCount++
			}

			vines = append(vines, *vine)
			vinesCreated++
		}
	}

	// Fill remaining space with simple paths if needed
	if occupiedCount < targetOccupancy {
		for occupiedCount < targetOccupancy && len(vines) < maxVines {
			path := generateSimplePath(len(vines), gridSize, occupied, rng)
			if len(path) == 0 {
				break
			}

			vine := &Vine{
				ID:            fmt.Sprintf("vine_%d", len(vines)),
				HeadDirection: "right",
				OrderedPath:   path,
				VineColor:     colors[len(vines)%len(colors)],
				Blocks:        []string{},
			}

			// Add to occupied
			for _, pt := range path {
				occupied[fmt.Sprintf("%d,%d", pt.X, pt.Y)] = true
				occupiedCount++
			}

			vines = append(vines, *vine)
		}
	}

	// Ensure at least minimum vine count
	minVines := spec.VineCountRange[0]
	for len(vines) < minVines && len(vines) < maxVines {
		vine := &Vine{
			ID:            fmt.Sprintf("vine_%d", len(vines)),
			HeadDirection: "right",
			OrderedPath:   generateSimplePath(len(vines), gridSize, occupied, rng),
			VineColor:     colors[len(vines)%len(colors)],
			Blocks:        []string{},
		}
		vines = append(vines, *vine)
	}

	// Ensure at least one vine
	if len(vines) == 0 {
		vine := generateDefaultVine(0, gridSize)
		vine.VineColor = colors[0]
		vines = append(vines, vine)
	}

	return vines
}

// generateSimplePath creates a simple horizontal or vertical path
func generateSimplePath(index int, gridSize [2]int, occupied map[string]bool, rng *rand.Rand) []Point {
	// Try a few random positions
	for i := 0; i < 10; i++ {
		x := rng.Intn(gridSize[0])
		y := rng.Intn(gridSize[1])

		if !occupied[fmt.Sprintf("%d,%d", x, y)] {
			// Try horizontal path
			path := []Point{{X: x, Y: y}}
			for len(path) < 5 && x+len(path) < gridSize[0] {
				if occupied[fmt.Sprintf("%d,%d", x+len(path), y)] {
					break
				}
				path = append(path, Point{X: x + len(path), Y: y})
			}

			if len(path) >= 2 {
				return path
			}
		}
	}

	// Fallback: simple 2-cell horizontal at top
	return []Point{{X: gridSize[0] - 1, Y: 0}, {X: gridSize[0] - 2, Y: 0}}
}

func generateRandomVine(index int, gridSize [2]int, occupied map[string]bool, rng *rand.Rand) *Vine {
	return generateVineWithLength(index, gridSize, occupied, rng, 2+rng.Intn(7))
}

func generateVineWithLength(index int, gridSize [2]int, occupied map[string]bool, rng *rand.Rand, targetLength int) *Vine {
	// Increase attempts to find valid placement
	maxAttempts := 500

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Random direction (this is the direction the vine points/moves)
		directions := []string{"up", "down", "left", "right"}
		direction := directions[rng.Intn(len(directions))]
		delta := HeadDirections[direction]

		// Try multiple head positions per attempt
		for tryPos := 0; tryPos < 100; tryPos++ {
			x := rng.Intn(gridSize[0])
			y := rng.Intn(gridSize[1])

			path := []Point{{X: x, Y: y}} // Head position

			// Add next segment (the neck) - must be in opposite direction
			nextX := x - delta[0]
			nextY := y - delta[1]

			if nextX < 0 || nextX >= gridSize[0] || nextY < 0 || nextY >= gridSize[1] {
				continue // Can't place vine with neck out of bounds
			}

			// Check occupancy for initial positions
			if occupied[fmt.Sprintf("%d,%d", x, y)] || occupied[fmt.Sprintf("%d,%d", nextX, nextY)] {
				continue
			}

			path = append(path, Point{X: nextX, Y: nextY})

			// Continue building the rest of the tail - always extend to target
			currentDir := direction
			for len(path) < targetLength {
				// Rarely change direction (10% chance) to keep vines long and straight
				if rng.Float64() < 0.1 && len(path) > 2 {
					// Pick perpendicular direction
					if currentDir == "up" || currentDir == "down" {
						if rng.Float64() < 0.5 {
							currentDir = "left"
						} else {
							currentDir = "right"
						}
					} else {
						if rng.Float64() < 0.5 {
							currentDir = "up"
						} else {
							currentDir = "down"
						}
					}
				}

				delta := HeadDirections[currentDir]
				lastPoint := path[len(path)-1]
				nextX := lastPoint.X - delta[0]
				nextY := lastPoint.Y - delta[1]

				// Check bounds - if out of bounds, stop extending
				if nextX < 0 || nextX >= gridSize[0] || nextY < 0 || nextY >= gridSize[1] {
					break
				}

				// Check occupancy
				key := fmt.Sprintf("%d,%d", nextX, nextY)
				if occupied[key] {
					break
				}

				path = append(path, Point{X: nextX, Y: nextY})
			}

			if len(path) >= 2 {
				// Validate no collisions with occupied
				canPlace := true
				for _, pt := range path {
					if occupied[fmt.Sprintf("%d,%d", pt.X, pt.Y)] {
						canPlace = false
						break
					}
				}

				if canPlace {
					// Recalculate the direction from head to neck to determine headDirection
					head := path[0]
					neck := path[1]
					headDir := "right" // Default

					// The headDirection should be the direction pointing from head TOWARDS where it came from
					// Since neck is behind head, if neck.X < head.X, the vine is pointing right (came from left)
					if neck.X < head.X {
						headDir = "right"
					} else if neck.X > head.X {
						headDir = "left"
					} else if neck.Y < head.Y {
						headDir = "up"
					} else if neck.Y > head.Y {
						headDir = "down"
					}

					return &Vine{
						ID:            fmt.Sprintf("vine_%d", index),
						HeadDirection: headDir,
						OrderedPath:   path,
						VineColor:     "default",
						Blocks:        []string{},
					}
				}
			}
		}
	}

	return nil
}

func generateDefaultVine(index int, gridSize [2]int) Vine {
	// Simple horizontal vine in middle of grid
	path := []Point{{X: gridSize[0] - 1, Y: gridSize[1] / 2}}
	for x := gridSize[0] - 2; x >= 0; x-- {
		path = append(path, Point{X: x, Y: gridSize[1] / 2})
	}

	return Vine{
		ID:            fmt.Sprintf("vine_%d", index),
		HeadDirection: "left",
		OrderedPath:   path,
		VineColor:     "default",
		Blocks:        []string{},
	}
}

func estimateMaxMoves(difficulty string) int {
	switch difficulty {
	case "Tutorial":
		return 5
	case "Seedling":
		return 8
	case "Sprout":
		return 10
	case "Nurturing":
		return 12
	case "Flourishing":
		return 18
	case "Transcendent":
		return 25
	default:
		return 10
	}
}

func estimateMinMoves(difficulty string) int {
	switch difficulty {
	case "Tutorial":
		return 3
	case "Seedling":
		return 4
	case "Sprout":
		return 5
	case "Nurturing":
		return 6
	case "Flourishing":
		return 8
	case "Transcendent":
		return 10
	default:
		return 5
	}
}

func outputLevelToStdout(level *Level) {
	data, err := marshalLevelJSON(level)
	if err != nil {
		log.Error("Failed to marshal level: %v", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

func marshalLevelJSON(level *Level) ([]byte, error) {
	return marshalJSON(level)
}
