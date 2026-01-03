package parable_bloom

import (
	"fmt"
	"math"
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

func generateBatch(modules []ModuleRange, moduleID, count int, output string, overwrite, verbose bool) {
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

	// Generate in parallel with reduced concurrency for very large grids to avoid OOM
	var wg sync.WaitGroup
	concurrency := 4
	if moduleID >= 7 {
		concurrency = 1
	} else if moduleID >= 5 {
		concurrency = 2
	}
	semaphore := make(chan struct{}, concurrency)

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

	// Pre-write validation: ensure level is actually solvable using the real solver
	// This should never fail since generateVines loops until solvable, but double-check
	solver := NewSolver(level)
	if !solver.IsSolvableGreedy() {
		// This is a critical error - generateVines should guarantee solvability
		log.Error("CRITICAL: Generated level %d failed solvability check despite generateVines guarantee", id)
		log.Error("This indicates a bug in the generator. Aborting to prevent unsolvable level from reaching production.")
		os.Exit(1)
	}

	return level
}

func generateVines(gridSize [2]int, difficulty string, levelID int) []Vine {
	seedStep := 1000
	var vines []Vine

	// Create temp level for solver validation
	tempLevel := &Level{
		ID:       levelID,
		GridSize: gridSize,
	}

	// Keep trying until we find a solvable level - no max retries
	// Production quality requires 100% solvable levels
	attempt := 0
	for {
		attempt++
		seed := int64(levelID + attempt*seedStep)
		vines = buildVines(gridSize, difficulty, seed)
		calculateBlocking(vines, gridSize)

		// Use the actual solver to check solvability (same as validator)
		tempLevel.Vines = vines
		solver := NewSolver(tempLevel)
		if solver.IsSolvableGreedy() {
			if attempt > 50 {
				log.Verbose(true, "Found solvable configuration for level %d after %d attempts", levelID, attempt)
			}
			return vines
		}

		// Progress logging for difficult levels
		if attempt%100 == 0 {
			log.Verbose(true, "Level %d: %d attempts, still searching for solvable configuration...", levelID, attempt)
		}
	}
}

// buildVines creates a set of vines for the given seed without solvability retries.
// For harder difficulties (Nurturing+), uses solver-aware placement for higher quality.
func buildVines(gridSize [2]int, difficulty string, seed int64) []Vine {
	// Use solver-aware placement for Nurturing and harder
	if difficulty == "Transcendent" || difficulty == "Flourishing" || difficulty == "Nurturing" {
		return buildVinesSolverAware(gridSize, difficulty, seed)
	}

	// Standard fast placement for easier difficulties
	return buildVinesFast(gridSize, difficulty, seed)
}

// buildVinesSolverAware creates vines with solver validation after each placement.
// Slower but produces higher-quality solvable configurations for hard levels.
func buildVinesSolverAware(gridSize [2]int, difficulty string, seed int64) []Vine {
	spec := DifficultySpecs[difficulty]
	rng := rand.New(rand.NewSource(seed))

	totalCells := gridSize[0] * gridSize[1]
	targetOccupancy := int(math.Ceil(float64(totalCells) * spec.MinGridOccupancy))

	var vines []Vine
	occupied := make(map[string]bool)
	colors := []string{"moss_green", "sunset_orange", "golden_yellow", "royal_purple", "sky_blue"}
	directions := []string{"up", "down", "left", "right"}

	maxVines := spec.VineCountRange[1]
	maxPlacementAttempts := totalCells * 20 // Higher limit for vine placement attempts
	attempts := 0

	// Create temp level for solver checks
	tempLevel := &Level{
		ID:       0,
		GridSize: gridSize,
	}

	for len(vines) < maxVines && totalOccupied(occupied) < targetOccupancy && attempts < maxPlacementAttempts {
		attempts++

		// Try to place a new vine
		x := rng.Intn(gridSize[0])
		y := rng.Intn(gridSize[1])
		key := fmt.Sprintf("%d,%d", x, y)

		if occupied[key] {
			continue
		}

		// Build a path in random direction
		direction := directions[rng.Intn(len(directions))]

		// Create empty blockedCells for this attempt (we'll validate with solver anyway)
		blockedCells := make(map[string]bool)
		path := buildPathFromCell(x, y, direction, gridSize, occupied, blockedCells)
		if len(path) < 2 {
			continue
		}

		// Create candidate vine
		vineID := fmt.Sprintf("vine_%d", len(vines))
		vineColor := colors[rng.Intn(len(colors))]
		headDirection := calcHeadDir(path)

		candidateVine := Vine{
			ID:            vineID,
			HeadDirection: headDirection,
			OrderedPath:   path,
			VineColor:     vineColor,
			Blocks:        []string{},
		}

		// Try adding this vine
		testVines := append([]Vine{}, vines...)
		testVines = append(testVines, candidateVine)
		calculateBlocking(testVines, gridSize)

		// Check if level remains solvable
		tempLevel.Vines = testVines
		solver := NewSolver(tempLevel)
		if solver.IsSolvableGreedy() {
			// Accept this vine
			vines = testVines
			markPathOccupied(occupied, path)
		}
		// If not solvable, discard and try again
	}

	// If we couldn't reach target, extend existing vines (without solver check for speed)
	if totalOccupied(occupied) < targetOccupancy {
		extendVinesToFill(&vines, occupied, gridSize, targetOccupancy)
	}

	return vines
}

// buildVinesFast creates vines using the standard fast algorithm (for easier difficulties).
func buildVinesFast(gridSize [2]int, difficulty string, seed int64) []Vine {
	spec := DifficultySpecs[difficulty]
	rng := rand.New(rand.NewSource(seed))

	totalCells := gridSize[0] * gridSize[1]
	targetOccupancy := int(math.Ceil(float64(totalCells) * spec.MinGridOccupancy))

	var vines []Vine
	occupied := make(map[string]bool)
	// blockedCells marks cells that are reserved because they lie in the exit path
	// of an already-placed vine. Prevent placing in these cells to reduce cycles
	// and circular blocking dependencies.
	blockedCells := make(map[string]bool)
	colors := []string{"moss_green", "sunset_orange", "golden_yellow", "royal_purple", "sky_blue"}

	occupiedCount := 0
	maxVines := spec.VineCountRange[1]

	// Phase 1: Random placement with longer vines
	occupiedCount = placeLongVines(&vines, occupied, blockedCells, gridSize, rng, targetOccupancy, maxVines, colors)

	// Phase 2: Fill remaining cells by extending existing vine tails
	if occupiedCount < targetOccupancy {
		occupiedCount = extendVinesToFill(&vines, occupied, gridSize, targetOccupancy)
	}

	// Phase 3: If still short, create minimal new vines for remaining gaps
	// Try to pair adjacent empty cells into 2-cell vines
	if occupiedCount < targetOccupancy && len(vines) < maxVines {
		pairGaps(&vines, occupied, blockedCells, gridSize, targetOccupancy, maxVines, colors)
	}

	return vines
}

// placeLongVines tries to place longer vines randomly until we reach target occupancy
// or hit the maximum number of attempts.
func placeLongVines(vines *[]Vine, occupied, blockedCells map[string]bool, gridSize [2]int, rng *rand.Rand, targetOccupancy, maxVines int, colors []string) int {
	totalCells := gridSize[0] * gridSize[1]
	maxAttempts := totalCells * 5
	attempts := 0
	occupiedCount := 0
	directions := []string{"up", "down", "left", "right"}

	for occupiedCount < targetOccupancy && len(*vines) < maxVines && attempts < maxAttempts {
		attempts++

		// Pick a random empty cell
		x := rng.Intn(gridSize[0])
		y := rng.Intn(gridSize[1])
		key := fmt.Sprintf("%d,%d", x, y)

		if occupied[key] || blockedCells[key] {
			continue
		}

		// Pick a random direction and build a path
		dir := directions[rng.Intn(len(directions))]
		path := buildPathFromCell(x, y, dir, gridSize, occupied, blockedCells)

		// Only accept paths with 2+ cells
		if len(path) < 2 {
			continue
		}

		headDir := calcHeadDir(path)

		// Create vine
		vine := Vine{
			ID:            fmt.Sprintf("vine_%d", len(*vines)),
			HeadDirection: headDir,
			OrderedPath:   path,
			VineColor:     colors[len(*vines)%len(colors)],
			Blocks:        []string{},
		}

		// Mark cells as occupied
		occupiedCount += markPathOccupied(occupied, path)

		// Mark exit path cells as blocked to prevent placing vines there later
		markExitPathBlocked(vine, blockedCells, gridSize)

		*vines = append(*vines, vine)
	}

	return occupiedCount
}

// extendVinesToFill attempts to extend tails of existing vines into adjacent empty cells
// to increase occupancy up to the targetOccupancy. Returns the new occupiedCount.
func extendVinesToFill(vines *[]Vine, occupied map[string]bool, gridSize [2]int, targetOccupancy int) int {
	occupiedCount := 0
	// Calculate existing occupied count
	for key := range occupied {
		if occupied[key] {
			occupiedCount++
		}
	}

	// Multiple passes to catch all stranded cells
	// Continue until we make no progress or hit target
	maxPasses := 10
	for pass := 0; pass < maxPasses && occupiedCount < targetOccupancy; pass++ {
		prevCount := occupiedCount

		// Scan grid for empty cells
		for y := 0; y < gridSize[1] && occupiedCount < targetOccupancy; y++ {
			for x := 0; x < gridSize[0] && occupiedCount < targetOccupancy; x++ {
				key := fmt.Sprintf("%d,%d", x, y)
				if occupied[key] {
					continue
				}

				// Try to extend adjacent vine tail into this cell (avoid blocking exit paths)
				if ok, added := tryExtendCellAt(x, y, vines, occupied, gridSize); ok {
					occupiedCount += added
				}
			}
		}

		// If we made no progress, stop early
		if occupiedCount == prevCount {
			break
		}
	}

	return occupiedCount
}

// pairGaps tries to pair adjacent empty cells into new 2-cell vines.
func pairGaps(vines *[]Vine, occupied, blockedCells map[string]bool, gridSize [2]int, targetOccupancy, maxVines int, colors []string) { // Refresh blockedCells to include all current exit paths before creating pairs
	for _, vine := range *vines {
		markExitPathBlocked(vine, blockedCells, gridSize)
	}
	for y := 0; y < gridSize[1] && totalOccupied(occupied) < targetOccupancy && len(*vines) < maxVines; y++ {
		for x := 0; x < gridSize[0] && totalOccupied(occupied) < targetOccupancy && len(*vines) < maxVines; x++ {
			tryCreatePairAt(x, y, vines, occupied, blockedCells, gridSize, colors)
		}
	}
}

// Helper: build a path starting at (x,y) following direction until blocked or out-of-bounds.
func buildPathFromCell(x, y int, dir string, gridSize [2]int, occupied, blockedCells map[string]bool) []Point {
	delta := HeadDirections[dir]
	path := []Point{{X: x, Y: y}}
	curX, curY := x, y

	for len(path) < 30 {
		nextX := curX - delta[0]
		nextY := curY - delta[1]

		// Check bounds
		if nextX < 0 || nextX >= gridSize[0] || nextY < 0 || nextY >= gridSize[1] {
			break
		}

		// Check occupancy or blocked
		if occupied[fmt.Sprintf("%d,%d", nextX, nextY)] || blockedCells[fmt.Sprintf("%d,%d", nextX, nextY)] {
			break
		}

		path = append(path, Point{X: nextX, Y: nextY})
		curX, curY = nextX, nextY
	}

	return path
}

// Helper: determine head direction given path (head is path[0], neck is path[1]).
func calcHeadDir(path []Point) string {
	head := path[0]
	neck := path[1]
	if neck.X < head.X {
		return "right"
	}
	if neck.X > head.X {
		return "left"
	}
	if neck.Y < head.Y {
		return "up"
	}
	return "down"
}

// Helper: mark all points in path as occupied, returning the number newly occupied.
func markPathOccupied(occupied map[string]bool, path []Point) int {
	added := 0
	for _, pt := range path {
		key := fmt.Sprintf("%d,%d", pt.X, pt.Y)
		if !occupied[key] {
			occupied[key] = true
			added++
		}
	}
	return added
}

// Helper: mark the exit path for a vine as blocked.
func markExitPathBlocked(vine Vine, blockedCells map[string]bool, gridSize [2]int) {
	exitPath := getExitPath(vine.GetHead(), vine.HeadDirection, gridSize)
	for _, cell := range exitPath {
		blockedCells[fmt.Sprintf("%d,%d", cell.X, cell.Y)] = true
	}
}

// totalOccupied returns number of occupied cells in the occupied map.
func totalOccupied(occupied map[string]bool) int {
	count := 0
	for _, v := range occupied {
		if v {
			count++
		}
	}
	return count
}

// tryExtendCellAt attempts to extend an adjacent vine tail into (x,y). Returns (ok, addedCount).
func tryExtendCellAt(x, y int, vines *[]Vine, occupied map[string]bool, gridSize [2]int) (bool, int) {
	// Check four adjacent directions for a tail
	dirs := []struct{ dx, dy int }{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	for _, d := range dirs {
		adjX, adjY := x+d.dx, y+d.dy
		if adjX < 0 || adjX >= gridSize[0] || adjY < 0 || adjY >= gridSize[1] {
			continue
		}
		adjKey := fmt.Sprintf("%d,%d", adjX, adjY)
		if !occupied[adjKey] {
			continue
		}

		if idx := findVineIndexWithTail(vines, adjX, adjY); idx >= 0 {
			(*vines)[idx].OrderedPath = append((*vines)[idx].OrderedPath, Point{X: x, Y: y})
			occupied[fmt.Sprintf("%d,%d", x, y)] = true
			return true, 1
		}
	}
	return false, 0
}

func findVineIndexWithTail(vines *[]Vine, x, y int) int {
	for i := range *vines {
		tail := (*vines)[i].GetTail()
		if tail.X == x && tail.Y == y {
			return i
		}
	}
	return -1
}

// tryCreatePairAt tries to create a paired vine at (x,y) and returns (ok, added).
func tryCreatePairAt(x, y int, vines *[]Vine, occupied, blockedCells map[string]bool, gridSize [2]int, colors []string) (bool, int) {
	key := fmt.Sprintf("%d,%d", x, y)
	if occupied[key] {
		return false, 0
	}

	dirs := []struct{ dx, dy int }{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	for _, d := range dirs {
		adjX, adjY := x+d.dx, y+d.dy
		if adjX < 0 || adjX >= gridSize[0] || adjY < 0 || adjY >= gridSize[1] {
			continue
		}
		adjKey := fmt.Sprintf("%d,%d", adjX, adjY)
		if occupied[adjKey] {
			continue
		}

		vineID := fmt.Sprintf("vine_%d", len(*vines)+1)
		headDir := calcHeadDir([]Point{{X: x, Y: y}, {X: adjX, Y: adjY}})

		newVine := Vine{
			ID:            vineID,
			VineColor:     colors[len(*vines)%len(colors)],
			HeadDirection: headDir,
			OrderedPath:   []Point{{X: x, Y: y}, {X: adjX, Y: adjY}},
		}
		*vines = append(*vines, newVine)
		occupied[key] = true
		occupied[adjKey] = true
		// Mark exit path for the new vine so we don't place inside its exit in later phases
		markExitPathBlocked(newVine, blockedCells, gridSize)
		return true, 2
	}
	return false, 0
}

// calculateBlocking determines which vines block each other based on exit paths.
func calculateBlocking(vines []Vine, gridSize [2]int) {
	for i := range vines {
		vines[i].Blocks = []string{}
		head := vines[i].GetHead()
		direction := vines[i].HeadDirection

		// Calculate exit path cells (from head to grid edge in head direction)
		exitPath := getExitPath(head, direction, gridSize)

		// Find all vines that occupy any cell in the exit path
		for j := range vines {
			if i == j {
				continue
			}

			// Check if vine j occupies any cell in vine i's exit path
			for _, pathCell := range exitPath {
				for _, vineCell := range vines[j].OrderedPath {
					if pathCell.X == vineCell.X && pathCell.Y == vineCell.Y {
						vines[i].Blocks = append(vines[i].Blocks, vines[j].ID)
						goto nextVine
					}
				}
			}
		nextVine:
		}
	}
}

// getExitPath returns all cells from the head in the given direction to the grid edge.
func getExitPath(head Point, direction string, gridSize [2]int) []Point {
	var path []Point
	x, y := head.X, head.Y

	switch direction {
	case "up":
		for y++; y < gridSize[1]; y++ {
			path = append(path, Point{X: x, Y: y})
		}
	case "down":
		for y--; y >= 0; y-- {
			path = append(path, Point{X: x, Y: y})
		}
	case "left":
		for x--; x >= 0; x-- {
			path = append(path, Point{X: x, Y: y})
		}
	case "right":
		for x++; x < gridSize[0]; x++ {
			path = append(path, Point{X: x, Y: y})
		}
	}

	return path
}

// isSolvable checks if the blocking graph has no cycles (can be topologically sorted).
func isSolvable(vines []Vine) bool {
	// Build adjacency list (vine ID -> list of vines it blocks)
	blocked := make(map[string][]string)
	inDegree := make(map[string]int)

	// Initialize all vines
	for _, vine := range vines {
		inDegree[vine.ID] = 0
		blocked[vine.ID] = []string{}
	}

	// Build graph: if A blocks B, then B depends on A (edge A -> B)
	// Reverse the relationship: vine.Blocks contains vines that must be cleared before this vine
	for _, vine := range vines {
		for _, blockedID := range vine.Blocks {
			blocked[blockedID] = append(blocked[blockedID], vine.ID)
			inDegree[vine.ID]++
		}
	}

	// Topological sort using Kahn's algorithm
	queue := []string{}
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	cleared := 0
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		cleared++

		// Remove edges from current to its blocked vines
		for _, blockedID := range blocked[current] {
			inDegree[blockedID]--
			if inDegree[blockedID] == 0 {
				queue = append(queue, blockedID)
			}
		}
	}

	// If we cleared all vines, no cycle exists
	return cleared == len(vines)
}

// isBruteForceSolvable simulates actual gameplay by attempting to clear vines
// from edges toward center. Returns true if all vines can be cleared.
func isBruteForceSolvable(vines []Vine, gridSize [2]int) bool {
	if len(vines) == 0 {
		return true
	}

	// Create copy of blocking relationships to mutate during simulation
	blockers := make(map[string][]string) // vine ID -> list of vines blocking it
	for _, vine := range vines {
		blockers[vine.ID] = append([]string{}, vine.Blocks...)
	}

	cleared := make(map[string]bool)
	maxIterations := len(vines) * 2 // Prevent infinite loops

	for iteration := 0; iteration < maxIterations; iteration++ {
		foundClearable := false

		// Find vines with no blockers that haven't been cleared
		for _, vine := range vines {
			if cleared[vine.ID] {
				continue
			}

			// Check if this vine can be cleared (no remaining blockers)
			if len(blockers[vine.ID]) == 0 {
				cleared[vine.ID] = true
				foundClearable = true

				// Remove this vine from other vines' blocker lists
				for otherID := range blockers {
					if cleared[otherID] {
						continue
					}
					newBlockers := []string{}
					for _, blockerID := range blockers[otherID] {
						if blockerID != vine.ID {
							newBlockers = append(newBlockers, blockerID)
						}
					}
					blockers[otherID] = newBlockers
				}
			}
		}

		// If no vines were cleared this iteration and we haven't cleared all, it's unsolvable
		if !foundClearable {
			return len(cleared) == len(vines)
		}

		// If all vines cleared, success
		if len(cleared) == len(vines) {
			return true
		}
	}

	// Timeout/deadlock
	return false
}

// scoreVineSetTopological evaluates the quality of a vine set using topological metrics.
// Higher score is better. Returns negative infinity if unsolvable.
// Note: This is a fast heuristic; use the actual Solver for definitive solvability.
func scoreVineSetTopological(vines []Vine, gridSize [2]int) float64 {
	if len(vines) == 0 {
		return -1000000.0
	}

	score := 0.0

	// Check topological solvability (fast heuristic)
	if !isSolvable(vines) {
		return -1000000.0 // Circular dependency = worst score
	}

	// Positive: more vines with no blockers (easier start)
	freeVines := 0
	for _, vine := range vines {
		if len(vine.Blocks) == 0 {
			freeVines++
		}
	}
	score += float64(freeVines) * 10.0

	// Positive: balanced blocking depth (prefer moderate dependency chains)
	maxBlockDepth := 0
	for _, vine := range vines {
		depth := len(vine.Blocks)
		if depth > maxBlockDepth {
			maxBlockDepth = depth
		}
	}
	if maxBlockDepth <= 3 {
		score += 20.0 // Shallow dependencies are good
	} else if maxBlockDepth > 5 {
		score -= float64(maxBlockDepth-5) * 5.0 // Penalize deep chains
	}

	// Positive: good occupancy (not too sparse, not too dense)
	occupiedCells := 0
	for _, vine := range vines {
		occupiedCells += len(vine.OrderedPath)
	}
	totalCells := gridSize[0] * gridSize[1]
	occupancy := float64(occupiedCells) / float64(totalCells)
	if occupancy >= 0.6 && occupancy <= 0.8 {
		score += 15.0 // Sweet spot
	} else if occupancy < 0.5 {
		score -= 10.0 // Too sparse
	}

	return score
}

// helper functions below replace unused legacy generators to keep code compact and avoid revive/unused warnings.
// Removed: generateRandomVine, generateVineWithLength, generateDefaultVine - not used by current generator flow.

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
