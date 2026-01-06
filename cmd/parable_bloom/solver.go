package parable_bloom

import (
	"fmt"
	"sort"
)

// Solver provides solvability checking for levels.
type Solver struct {
	level *Level
}

// NewSolver creates a new solver for a level.
func NewSolver(level *Level) *Solver {
	return &Solver{level: level}
}

// IsSolvableGreedy checks solvability using a fast greedy algorithm.
// This is used during level generation for speed.
func (s *Solver) IsSolvableGreedy() bool {
	if len(s.level.Vines) == 0 {
		return true
	}

	currentVines := make([]Vine, len(s.level.Vines))
	copy(currentVines, s.level.Vines)

	occupied := make(map[string]bool)
	for _, vine := range currentVines {
		for _, pt := range vine.OrderedPath {
			occupied[fmt.Sprintf("%d,%d", pt.X, pt.Y)] = true
		}
	}

	// Greedy removal: repeatedly find and remove clearable vines
	maxIterations := len(currentVines) * 2
	iterations := 0

	for len(currentVines) > 0 && iterations < maxIterations {
		iterations++
		foundClearable := false

		for i, vine := range currentVines {
			if s.canVineClear(&vine, occupied) {
				// Remove this vine from occupied
				for _, pt := range vine.OrderedPath {
					delete(occupied, fmt.Sprintf("%d,%d", pt.X, pt.Y))
				}
				// Remove vine from list
				currentVines = append(currentVines[:i], currentVines[i+1:]...)
				foundClearable = true
				break
			}
		}

		if !foundClearable {
			// Deadlock: no clearable vines
			return false
		}
	}

	return len(currentVines) == 0
}

// IsSolvableBFS checks solvability using a thorough BFS algorithm.
// This is the gold standard for validation but slower.
func (s *Solver) IsSolvableBFS() bool {
	if len(s.level.Vines) == 0 {
		return true
	}

	// BFS state: set of remaining vine IDs
	initialState := make(map[string]bool)
	for _, vine := range s.level.Vines {
		initialState[vine.ID] = true
	}

	queue := []map[string]bool{initialState}
	visited := make(map[string]bool)
	visited[stateKey(initialState)] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if len(current) == 0 {
			return true // All vines cleared
		}

		// Try removing each clearable vine
		vines := s.getVinesForIDs(current)
		occupied := s.buildOccupiedMap(vines)

		for _, vine := range vines {
			if s.canVineClear(&vine, occupied) {
				// Create new state without this vine
				next := make(map[string]bool)
				for id := range current {
					if id != vine.ID {
						next[id] = true
					}
				}

				key := stateKey(next)
				if !visited[key] {
					visited[key] = true
					queue = append(queue, next)
				}
			}
		}
	}

	return false
}

// canVineClear checks if a vine can move and eventually exit the grid.
func (s *Solver) canVineClear(vine *Vine, occupiedCells map[string]bool) bool {
	head := vine.GetHead()
	delta := HeadDirections[vine.HeadDirection]

	if delta[0] == 0 && delta[1] == 0 {
		return false
	}

	if len(vine.OrderedPath) < 2 {
		return false
	}

	// Simulate movement for up to (width + height + path length) steps
	maxSteps := s.level.GetGridWidth() + s.level.GetGridHeight() + len(vine.OrderedPath) + 10

	for step := 0; step < maxSteps; step++ {
		// Calculate head position after this move
		nextX := head.X + delta[0]
		nextY := head.Y + delta[1]

		// Check if exited grid (vine clears)
		if nextX < 0 || nextX >= s.level.GetGridWidth() ||
			nextY < 0 || nextY >= s.level.GetGridHeight() {
			return true
		}

		// Check if cell is occupied (by another vine)
		cellKey := fmt.Sprintf("%d,%d", nextX, nextY)
		if occupiedCells[cellKey] {
			// Blocked, can't proceed
			return false
		}

		// Continue from new position
		head = Point{X: nextX, Y: nextY}
	}

	return false
}

// buildOccupiedMap creates a map of occupied cells from vines.
func (s *Solver) buildOccupiedMap(vines []Vine) map[string]bool {
	occupied := make(map[string]bool)
	for _, vine := range vines {
		for _, pt := range vine.OrderedPath {
			occupied[fmt.Sprintf("%d,%d", pt.X, pt.Y)] = true
		}
	}
	return occupied
}

// getVinesForIDs returns vine objects for given IDs.
func (s *Solver) getVinesForIDs(ids map[string]bool) []Vine {
	var result []Vine
	for _, vine := range s.level.Vines {
		if ids[vine.ID] {
			result = append(result, vine)
		}
	}
	return result
}

// stateKey creates a unique, deterministic key for a state (set of vine IDs).
func stateKey(state map[string]bool) string {
	ids := make([]string, 0, len(state))
	for id := range state {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	key := ""
	for _, id := range ids {
		key += id + ","
	}
	return key
}
