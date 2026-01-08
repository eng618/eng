package common

import (
	"fmt"
)

// Validator validates level structure and constraints.
type Validator struct {
	level      *Level
	violations []string
	warnings   []string
}

// Validate performs complete validation on a level.
func (l *Level) Validate() (violations, warnings []string) {
	v := &Validator{level: l}
	v.validateBasicStructure()
	v.validateVinePaths()
	v.validateGridOccupancy()
	v.validateColors()
	v.validateVineLengths()
	v.validateBlockingRelationships()
	v.validateDirectionalBalance()
	v.validateDifficultyCompliance()
	return v.violations, v.warnings
}

func (v *Validator) addViolation(msg string) {
	v.violations = append(v.violations, msg)
}

func (v *Validator) addWarning(msg string) {
	v.warnings = append(v.warnings, msg)
}

func (v *Validator) validateBasicStructure() {
	if v.level.ID <= 0 {
		v.addViolation("id must be positive")
	}
	if v.level.Name == "" {
		v.addViolation("name is required")
	}
	if v.level.Difficulty == "" {
		v.addViolation("difficulty is required")
	}
	if _, ok := DifficultySpecs[v.level.Difficulty]; !ok {
		v.addViolation(fmt.Sprintf("unknown difficulty: %s", v.level.Difficulty))
	}
	if v.level.GridSize[0] <= 0 || v.level.GridSize[1] <= 0 {
		v.addViolation(fmt.Sprintf("grid_size must be positive: %v", v.level.GridSize))
	}
	if len(v.level.Vines) == 0 {
		v.addViolation("at least one vine is required")
	}
	if v.level.MaxMoves <= 0 {
		v.addViolation("max_moves must be positive")
	}
	if v.level.MinMoves < 0 {
		v.addViolation("min_moves must be non-negative")
	}
	if v.level.MinMoves > v.level.MaxMoves {
		v.addViolation(fmt.Sprintf("min_moves (%d) cannot exceed max_moves (%d)", v.level.MinMoves, v.level.MaxMoves))
	}
	if v.level.Grace < 0 {
		v.addViolation("grace must be non-negative")
	}
}

func (v *Validator) validateVinePaths() {
	for _, vine := range v.level.Vines {
		if vine.ID == "" {
			v.addViolation("vine id is required")
			continue
		}
		if len(vine.OrderedPath) < 2 {
			v.addViolation(
				fmt.Sprintf("vine %s: path must have at least 2 segments, has %d", vine.ID, len(vine.OrderedPath)),
			)
			continue
		}
		// Check path contiguity
		for i := 0; i < len(vine.OrderedPath)-1; i++ {
			curr := vine.OrderedPath[i]
			next := vine.OrderedPath[i+1]
			dist := abs(curr.X-next.X) + abs(curr.Y-next.Y)
			if dist != 1 {
				v.addViolation(
					fmt.Sprintf("vine %s: path not contiguous at segment %d->%d (distance %d)", vine.ID, i, i+1, dist),
				)
			}
		}
		// Validate head direction
		if _, ok := HeadDirections[vine.HeadDirection]; !ok {
			v.addViolation(fmt.Sprintf("vine %s: unknown head_direction: %s", vine.ID, vine.HeadDirection))
			continue
		}
		// Validate head-neck relationship
		head := vine.GetHead()
		neck := vine.GetNeck()
		delta := HeadDirections[vine.HeadDirection]
		expectedNeck := Point{X: head.X - delta[0], Y: head.Y - delta[1]}
		if neck.X != expectedNeck.X || neck.Y != expectedNeck.Y {
			v.addViolation(
				fmt.Sprintf(
					"vine %s: head_direction %s doesn't match neck position (head %v, expected neck %v, got %v)",
					vine.ID,
					vine.HeadDirection,
					head,
					expectedNeck,
					neck,
				),
			)
		}
		// Validate vine bounds
		for i, pt := range vine.OrderedPath {
			if pt.X < 0 || pt.X >= v.level.GetGridWidth() || pt.Y < 0 || pt.Y >= v.level.GetGridHeight() {
				v.addViolation(fmt.Sprintf("vine %s: segment %d at %v is out of bounds", vine.ID, i, pt))
			}
		}
	}
}

func (v *Validator) validateGridOccupancy() {
	totalCells := v.level.GetTotalCells()
	occupiedCells := v.level.GetOccupiedCells()
	occupancy := float64(occupiedCells) / float64(totalCells)
	v.level.OccupancyPercent = occupancy * 100

	// Grid occupancy is a requirement
	spec := DifficultySpecs[v.level.Difficulty]
	if occupancy < spec.MinGridOccupancy {
		v.addViolation(fmt.Sprintf("grid occupancy too low: %.1f%% (need %.0f%%), %d/%d cells",
			occupancy*100, spec.MinGridOccupancy*100, occupiedCells, totalCells))
	}
}

func (v *Validator) validateColors() {
	colorCounts := make(map[string]int)
	for _, vine := range v.level.Vines {
		color := vine.VineColor
		if color == "" {
			color = "default"
		}
		colorCounts[color]++
	}

	// Check colors are known
	for color := range colorCounts {
		if color != "default" && color != "unknown" {
			if _, ok := VineColors[color]; !ok {
				v.addWarning(fmt.Sprintf("unknown vine color: %s", color))
			}
		}
	}

	// Check color count range
	spec := DifficultySpecs[v.level.Difficulty]
	uniqueColors := len(colorCounts)
	if uniqueColors < spec.ColorCountRange[0] || uniqueColors > spec.ColorCountRange[1] {
		v.addWarning(fmt.Sprintf("color count %d outside range %v for %s",
			uniqueColors, spec.ColorCountRange, v.level.Difficulty))
	}

	// Check no color exceeds 35%
	vineCount := float64(len(v.level.Vines))
	for color, count := range colorCounts {
		if color == "unknown" {
			continue
		}
		percentage := float64(count) / vineCount
		if percentage > 0.35 {
			v.addWarning(fmt.Sprintf("color %s exceeds 35%% (%.1f%%)", color, percentage*100))
		}
	}

	// Store color distribution
	v.level.ColorDistribution = make(map[string]float64)
	for color, count := range colorCounts {
		v.level.ColorDistribution[color] = float64(count) / vineCount
	}
}

func (v *Validator) validateVineLengths() {
	// Only requirement: every vine must be at least 2 coordinates (head + neck)
	// This is already checked in validateVinePaths()
	// All other length constraints are soft/informational
	if len(v.level.Vines) == 0 {
		return
	}

	spec := DifficultySpecs[v.level.Difficulty]
	lengths := make([]int, len(v.level.Vines))
	for i, vine := range v.level.Vines {
		lengths[i] = vine.Length()
	}

	totalLength := 0
	minLength := lengths[0]
	for _, l := range lengths {
		totalLength += l
		if l < minLength {
			minLength = l
		}
	}
	avgLength := float64(totalLength) / float64(len(lengths))

	// Average length is informational only
	if avgLength < float64(spec.AvgLengthRange[0]) || avgLength > float64(spec.AvgLengthRange[1]) {
		v.addWarning(fmt.Sprintf("average vine length %.1f outside recommended range %v for %s",
			avgLength, spec.AvgLengthRange, v.level.Difficulty))
	}
}

func (v *Validator) validateBlockingRelationships() {
	// Build vine ID set
	vineIDs := make(map[string]bool)
	for _, vine := range v.level.Vines {
		vineIDs[vine.ID] = true
	}

	// Validate blocks references
	blockingGraph := make(map[string][]string)
	for _, vine := range v.level.Vines {
		for _, blockedID := range vine.Blocks {
			if !vineIDs[blockedID] {
				v.addViolation(fmt.Sprintf("vine %s blocks unknown vine: %s", vine.ID, blockedID))
				continue
			}
			blockingGraph[vine.ID] = append(blockingGraph[vine.ID], blockedID)
		}
	}

	// Check for circular dependencies
	visited := make(map[string]bool)
	stack := make(map[string]bool)

	var hasCycle func(id string) bool
	hasCycle = func(id string) bool {
		visited[id] = true
		stack[id] = true

		for _, blocked := range blockingGraph[id] {
			if !visited[blocked] {
				if hasCycle(blocked) {
					return true
				}
			} else if stack[blocked] {
				return true
			}
		}

		stack[id] = false
		return false
	}

	for vineID := range vineIDs {
		if !visited[vineID] {
			if hasCycle(vineID) {
				v.addViolation("circular blocking dependency detected")
				break
			}
		}
	}

	// Check at least one vine is clearable at start
	occupied := make(map[string]bool)
	for _, vine := range v.level.Vines {
		for _, pt := range vine.OrderedPath {
			occupied[fmt.Sprintf("%d,%d", pt.X, pt.Y)] = true
		}
	}

	hasClerable := false
	for _, vine := range v.level.Vines {
		isBlocked := false
		for _, other := range v.level.Vines {
			if other.ID != vine.ID {
				for _, blockedID := range other.Blocks {
					if blockedID == vine.ID {
						isBlocked = true
						break
					}
				}
			}
			if isBlocked {
				break
			}
		}
		if !isBlocked {
			hasClerable = true
			break
		}
	}

	if !hasClerable {
		v.addViolation("no vines are clearable at level start (deadlock)")
	}

	v.level.BlockingGraph = blockingGraph
}

func (v *Validator) validateDirectionalBalance() {
	if len(v.level.Vines) < 10 {
		return
	}

	dirCounts := make(map[string]int)
	for _, vine := range v.level.Vines {
		dirCounts[vine.HeadDirection]++
	}

	expectedRanges := map[string][2]float64{
		"right": {0.25, 0.30},
		"left":  {0.20, 0.25},
		"up":    {0.20, 0.25},
		"down":  {0.20, 0.30},
	}

	total := float64(len(v.level.Vines))
	for dir, rng := range expectedRanges {
		count := float64(dirCounts[dir])
		pct := count / total
		if pct < rng[0] || pct > rng[1] {
			v.addWarning(fmt.Sprintf("direction %s outside expected range: %.1f%% (want %.0f-%.0f%%)",
				dir, pct*100, rng[0]*100, rng[1]*100))
		}
	}
}

func (v *Validator) validateDifficultyCompliance() {
	spec := DifficultySpecs[v.level.Difficulty]
	vineCount := len(v.level.Vines)

	// Grace is a requirement based on difficulty
	if v.level.Grace != spec.DefaultGrace {
		v.addViolation(fmt.Sprintf("grace must be %d for %s difficulty (got %d)",
			spec.DefaultGrace, v.level.Difficulty, v.level.Grace))
	}

	// Vine count is a design guideline, not a hard requirement
	if vineCount < spec.VineCountRange[0] || vineCount > spec.VineCountRange[1] {
		v.addWarning(fmt.Sprintf("vine count %d outside recommended range %v for %s",
			vineCount, spec.VineCountRange, v.level.Difficulty))
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
