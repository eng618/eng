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
		moduleID, _ := cmd.Flags().GetInt("module")
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
			generateBatch(modules, moduleID, count, output, overwrite, isVerbose)
		} else {
			// Single level generation
			generateSingle(name, width, height, output, stdout, overwrite, modules, isVerbose)
		}
	},
}

func init() {
	// Add flags for level generation
	LevelGenerateCmd.Flags().StringP("name", "n", "", "Name of the new level")
	LevelGenerateCmd.Flags().IntP("module", "m", 0, "Module ID for batch generation (1-8)")
	LevelGenerateCmd.Flags().IntP("grid-width", "w", 0, "Width of the game grid (0 = auto)")
	LevelGenerateCmd.Flags().IntP("grid-height", "H", 0, "Height of the game grid (0 = auto)")
	LevelGenerateCmd.Flags().StringP("output", "o", "", "Output directory for level files (default: assets/levels)")
	LevelGenerateCmd.Flags().BoolP("stdout", "", false, "Print level JSON to stdout instead of file")
	LevelGenerateCmd.Flags().BoolP("overwrite", "", false, "Overwrite existing level files")
	LevelGenerateCmd.Flags().IntP("count", "c", 1, "Generate multiple levels (batch mode)")
	LevelGenerateCmd.Flags().BoolP("dry-run", "", false, "Generate without writing to disk")
}

func generateSingle(name string, width, height int, output string, stdout, overwrite bool, modules []ModuleRange, verbose bool) {
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

	// Determine difficulty based on level ID
	difficulty = DifficultyForLevel(levelID, modules)

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

func generateBatch(modules []ModuleRange, moduleID int, count int, output string, overwrite bool, verbose bool) {
	log.Verbose(verbose, "Generating batch of %d levels for module ID: %d", count, moduleID)

	// Find module range by ID (1-indexed)
	if moduleID < 1 || moduleID > len(modules) {
		log.Error("Invalid module ID: %d (must be 1-%d)", moduleID, len(modules))
		os.Exit(1)
	}

	moduleRange := &modules[moduleID-1]

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

	totalCells := gridSize[0] * gridSize[1]
	targetOccupancy := totalCells * 95 / 100
	if difficulty == "Tutorial" {
		targetOccupancy = totalCells * 50 / 100
	}

	var vines []Vine
	occupied := make(map[string]bool)
	colors := []string{"moss_green", "sunset_orange", "golden_yellow", "royal_purple", "sky_blue"}
	directions := []string{"up", "down", "left", "right"}

	occupiedCount := 0
	maxVines := spec.VineCountRange[1]
	minVines := spec.VineCountRange[0]

	// Phase 1: Random placement with longer vines
	maxAttempts := totalCells * 5
	attempts := 0
	
	for occupiedCount < targetOccupancy && len(vines) < maxVines && attempts < maxAttempts {
		attempts++

		// Pick a random empty cell
		x := rng.Intn(gridSize[0])
		y := rng.Intn(gridSize[1])
		key := fmt.Sprintf("%d,%d", x, y)

		if occupied[key] {
			continue
		}

		// Pick a random direction
		dir := directions[rng.Intn(len(directions))]
		delta := HeadDirections[dir]

		// Build a path from this cell - extend as far as possible (up to 30 cells for thorough filling)
		path := []Point{{X: x, Y: y}}
		curX, curY := x, y

		for len(path) < 30 {
			nextX := curX - delta[0]
			nextY := curY - delta[1]

			// Check bounds
			if nextX < 0 || nextX >= gridSize[0] || nextY < 0 || nextY >= gridSize[1] {
				break
			}

			// Check occupancy
			if occupied[fmt.Sprintf("%d,%d", nextX, nextY)] {
				break
			}

			path = append(path, Point{X: nextX, Y: nextY})
			curX, curY = nextX, nextY
		}

		// Only accept paths with 2+ cells
		if len(path) < 2 {
			continue
		}

		// Calculate head direction from path
		head := path[0]
		neck := path[1]
		headDir := "right"
		if neck.X < head.X {
			headDir = "right"
		} else if neck.X > head.X {
			headDir = "left"
		} else if neck.Y < head.Y {
			headDir = "up"
		} else if neck.Y > head.Y {
			headDir = "down"
		}

		// Create vine
		vine := Vine{
			ID:            fmt.Sprintf("vine_%d", len(vines)),
			HeadDirection: headDir,
			OrderedPath:   path,
			VineColor:     colors[len(vines)%len(colors)],
			Blocks:        []string{},
		}

		// Mark cells as occupied
		for _, pt := range path {
			occupied[fmt.Sprintf("%d,%d", pt.X, pt.Y)] = true
			occupiedCount++
		}

		vines = append(vines, vine)
	}

	// Phase 2: Deterministic grid-based filling of remaining gaps
	if occupiedCount < targetOccupancy && len(vines) < maxVines {
		// Scan grid row by row, fill any remaining empty cells
		for y := 0; y < gridSize[1] && occupiedCount < targetOccupancy && len(vines) < maxVines; y++ {
			for x := 0; x < gridSize[0] && occupiedCount < targetOccupancy && len(vines) < maxVines; x++ {
				if occupied[fmt.Sprintf("%d,%d", x, y)] {
					continue
				}

				// Found empty cell - try to extend it in 4 directions in order
				directions4 := []string{"right", "down", "left", "up"}
				filledCell := false

				for _, dir := range directions4 {
					if occupiedCount >= targetOccupancy || len(vines) >= maxVines {
						break
					}

					delta := HeadDirections[dir]
					path := []Point{{X: x, Y: y}}
					curX, curY := x, y

					// Greedily extend as far as possible
					for {
						nextX := curX - delta[0]
						nextY := curY - delta[1]

						// Check bounds
						if nextX < 0 || nextX >= gridSize[0] || nextY < 0 || nextY >= gridSize[1] {
							break
						}

						// Check occupancy
						if occupied[fmt.Sprintf("%d,%d", nextX, nextY)] {
							break
						}

						path = append(path, Point{X: nextX, Y: nextY})
						curX, curY = nextX, nextY

						// Don't make vines too long in filling phase
						if len(path) >= 8 {
							break
						}
					}

					if len(path) >= 2 {
						// Calculate head direction
						head := path[0]
						neck := path[1]
						headDir := "right"
						if neck.X < head.X {
							headDir = "right"
						} else if neck.X > head.X {
							headDir = "left"
						} else if neck.Y < head.Y {
							headDir = "up"
						} else if neck.Y > head.Y {
							headDir = "down"
						}

						vine := Vine{
							ID:            fmt.Sprintf("vine_%d", len(vines)),
							HeadDirection: headDir,
							OrderedPath:   path,
							VineColor:     colors[len(vines)%len(colors)],
							Blocks:        []string{},
						}

						for _, pt := range path {
							occupied[fmt.Sprintf("%d,%d", pt.X, pt.Y)] = true
							occupiedCount++
						}

						vines = append(vines, vine)
						filledCell = true
						break
					}
				}

				// If we still couldn't fill this cell, at least create a 2-cell vine
				if !filledCell && occupiedCount < targetOccupancy && len(vines) < maxVines {
					// Try to find a neighbor for a 2-cell vine
					for _, dir := range directions4 {
						delta := HeadDirections[dir]
						nextX := x - delta[0]
						nextY := y - delta[1]

						if nextX >= 0 && nextX < gridSize[0] && nextY >= 0 && nextY < gridSize[1] {
							if !occupied[fmt.Sprintf("%d,%d", nextX, nextY)] {
								path := []Point{{X: x, Y: y}, {X: nextX, Y: nextY}}
								headDir := dir

								vine := Vine{
									ID:            fmt.Sprintf("vine_%d", len(vines)),
									HeadDirection: headDir,
									OrderedPath:   path,
									VineColor:     colors[len(vines)%len(colors)],
									Blocks:        []string{},
								}

								for _, pt := range path {
									occupied[fmt.Sprintf("%d,%d", pt.X, pt.Y)] = true
									occupiedCount++
								}

								vines = append(vines, vine)
								break
							}
						}
					}
				}
			}
		}
	}

	// Ensure minimum vine count if needed
	for len(vines) < minVines {
		vine := Vine{
			ID:            fmt.Sprintf("vine_%d", len(vines)),
			HeadDirection: "right",
			OrderedPath:   []Point{{X: rng.Intn(gridSize[0]), Y: rng.Intn(gridSize[1])}, {X: (rng.Intn(gridSize[0]) + 1) % gridSize[0], Y: rng.Intn(gridSize[1])}},
			VineColor:     colors[len(vines)%len(colors)],
			Blocks:        []string{},
		}
		vines = append(vines, vine)
	}

	// Fallback: ensure at least one vine
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
