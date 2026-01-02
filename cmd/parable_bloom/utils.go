package parable_bloom

import (
	"fmt"
	"math/rand"
	"sort"
)

// GenerateLevelID generates a unique level ID.
func GenerateLevelID(baseDir string, start int) int {
	// Find next available ID
	for id := start; id < 100000; id++ {
		path := GetLevelFilePath(id, baseDir)
		if !FileExists(path) {
			return id
		}
	}
	return start
}

// ComputeOccupancy calculates the grid occupancy percentage for a level.
func ComputeOccupancy(level *Level) float64 {
	total := level.GetTotalCells()
	if total == 0 {
		return 0
	}
	occupied := level.GetOccupiedCells()
	return float64(occupied) / float64(total)
}

// DetectCircularBlocking detects circular dependencies in blocking relationships.
func DetectCircularBlocking(vines []Vine) bool {
	// Build adjacency list
	graph := make(map[string][]string)
	vineIDSet := make(map[string]bool)

	for _, vine := range vines {
		vineIDSet[vine.ID] = true
		for _, blocked := range vine.Blocks {
			graph[vine.ID] = append(graph[vine.ID], blocked)
		}
	}

	// DFS to detect cycle
	visited := make(map[string]bool)
	stack := make(map[string]bool)

	var hasCycle func(id string) bool
	hasCycle = func(id string) bool {
		visited[id] = true
		stack[id] = true

		for _, next := range graph[id] {
			if !visited[next] {
				if hasCycle(next) {
					return true
				}
			} else if stack[next] {
				return true
			}
		}

		stack[id] = false
		return false
	}

	for vineID := range vineIDSet {
		if !visited[vineID] {
			if hasCycle(vineID) {
				return true
			}
		}
	}

	return false
}

// GetAverageVineLength computes average vine length.
func GetAverageVineLength(vines []Vine) float64 {
	if len(vines) == 0 {
		return 0
	}
	total := 0
	for _, vine := range vines {
		total += vine.Length()
	}
	return float64(total) / float64(len(vines))
}

// GetMinimumVineLength returns the shortest vine length.
func GetMinimumVineLength(vines []Vine) int {
	if len(vines) == 0 {
		return 0
	}
	min := vines[0].Length()
	for _, vine := range vines {
		if vine.Length() < min {
			min = vine.Length()
		}
	}
	return min
}

// GetMaximumVineLength returns the longest vine length.
func GetMaximumVineLength(vines []Vine) int {
	if len(vines) == 0 {
		return 0
	}
	max := vines[0].Length()
	for _, vine := range vines {
		if vine.Length() > max {
			max = vine.Length()
		}
	}
	return max
}

// ShuffleVines randomly shuffles a slice of vines using a deterministic seed.
func ShuffleVines(vines []Vine, seed int64) []Vine {
	result := make([]Vine, len(vines))
	copy(result, vines)

	rng := rand.New(rand.NewSource(seed))
	for i := len(result) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		result[i], result[j] = result[j], result[i]
	}

	return result
}

// FindVineByID finds a vine by its ID.
func FindVineByID(vines []Vine, id string) *Vine {
	for i := range vines {
		if vines[i].ID == id {
			return &vines[i]
		}
	}
	return nil
}

// CountVinesByColor counts how many vines use each color.
func CountVinesByColor(vines []Vine) map[string]int {
	counts := make(map[string]int)
	for _, vine := range vines {
		color := vine.VineColor
		if color == "" {
			color = "default"
		}
		counts[color]++
	}
	return counts
}

// CountVinesByDirection counts how many vines face each direction.
func CountVinesByDirection(vines []Vine) map[string]int {
	counts := make(map[string]int)
	for _, vine := range vines {
		counts[vine.HeadDirection]++
	}
	return counts
}

// SortVinesByID returns vines sorted by ID.
func SortVinesByID(vines []Vine) []Vine {
	result := make([]Vine, len(vines))
	copy(result, vines)
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result
}

// ComplexityForDifficulty returns the recommended complexity level.
func ComplexityForDifficulty(difficulty string) string {
	switch difficulty {
	case "Tutorial":
		return "tutorial"
	case "Seedling":
		return "low"
	case "Sprout":
		return "low"
	case "Nurturing":
		return "medium"
	case "Flourishing":
		return "high"
	case "Transcendent":
		return "extreme"
	default:
		return "medium"
	}
}

// GraceForDifficulty returns the default grace value for a difficulty.
func GraceForDifficulty(difficulty string) int {
	if spec, ok := DifficultySpecs[difficulty]; ok {
		return spec.DefaultGrace
	}
	return 3
}

// DefaultGridSize returns default grid size for a difficulty (used when not specified).
func DefaultGridSize(difficulty string) [2]int {
	ranges, ok := GridSizeRanges[difficulty]
	if !ok {
		return [2]int{9, 12}
	}
	// Return middle of range
	w := (ranges.MinW + ranges.MaxW) / 2
	h := (ranges.MinH + ranges.MaxH) / 2
	return [2]int{w, h}
}

// PointKey creates a unique key for a point (used in maps).
func PointKey(pt Point) string {
	return fmt.Sprintf("%d,%d", pt.X, pt.Y)
}

// ParsePointKey parses a point key back to coordinates.
func ParsePointKey(key string) (x, y int) {
	fmt.Sscanf(key, "%d,%d", &x, &y)
	return
}
