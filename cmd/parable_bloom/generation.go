package parable_bloom

import (
	"math/rand"
)

// GenerationResult contains metadata about a generation attempt.
type GenerationResult struct {
	Vines            []Vine
	Score            float64
	MaxBlockingDepth int
	GreedySolvable   bool
	BFSSolvable      bool
	Attempts         int
}

// FastScoreBlocking computes a simple blocking score and maximum blocking depth
// for a set of vines on a grid. It mutates vine.Blocks via calculateBlocking.
func FastScoreBlocking(vines []Vine, gridSize [2]int) (float64, int) {
	calculateBlocking(vines, gridSize)
	maxDepth := 0
	for _, v := range vines {
		if len(v.Blocks) > maxDepth {
			maxDepth = len(v.Blocks)
		}
	}
	score := scoreVineSetTopological(vines, gridSize)
	return score, maxDepth
}

// GenerateWithProfile tries to create a tiled level using the given profile and config.
// It runs fast pre-validation and greedy solver gating; optionally runs BFS for final verification
// if strictMode is true. It returns a GenerationResult containing score and solvability flags.
func GenerateWithProfile(gridSize [2]int, constraints DifficultySpec, profile VarietyProfile, cfg GeneratorConfig, seed int64, strictMode bool, rng *rand.Rand) GenerationResult {
	result := GenerationResult{Attempts: 0}
	// Try a few tiled variants with different RNG states
	for attempt := 0; attempt < 8; attempt++ {
		result.Attempts++
		// Use the provided rng but also vary seed with attempt to increase diversity
		var localRng *rand.Rand
		if rng != nil {
			localRng = rand.New(rand.NewSource(rng.Int63() + int64(attempt*1000)))
		} else {
			localRng = rand.New(rand.NewSource(seed + int64(attempt*1000)))
		}

		vines, err := TileGridIntoVines(gridSize, constraints, profile, cfg, localRng)
		if err != nil {
			continue
		}

		// Fast structural validation (should pass since tiler ensures this)
		lvl := &Level{GridSize: gridSize, Vines: vines}
		if err := FastValidateLevelCoverage(lvl); err != nil {
			continue
		}

		// Score and compute blocking depth
		score, maxDepth := FastScoreBlocking(vines, gridSize)
		result.Score = score
		result.MaxBlockingDepth = maxDepth
		result.Vines = vines

		// Greedy solvability check
		s := NewSolver(lvl)
		result.GreedySolvable = s.IsSolvableGreedy()
		if !result.GreedySolvable {
			// try next attempt
			continue
		}

		// If strict mode, run BFS check as well
		if strictMode {
			result.BFSSolvable = s.IsSolvableBFS()
			if !result.BFSSolvable {
				// accept or continue? For now, consider BFS failure as reject
				continue
			}
		}

		// Success
		return result
	}

	return result
}
