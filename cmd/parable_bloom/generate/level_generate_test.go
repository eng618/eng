//go:build ignore
// +build ignore

package generate

import (
	"fmt"
	"strings"
	"testing"

	"github.com/eng618/eng/cmd/parable_bloom/common"
	"github.com/eng618/eng/internal/utils/log"
)

func TestParableBloomTestsRemoved(t *testing.T) {
	t.Skip("Parable Bloom eng CLI tests removed; use tools/level-builder tests instead.")
}

// TestDifficultyForLevel_Progression verifies that difficulty increases correctly through a module.
func TestDifficultyForLevel_Progression(t *testing.T) {
	modules, err := common.LoadModules("")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		levelID            int
		expectedDifficulty string
	}{
		// Module 2 (6-20): 15 levels, progression should be monotonic
		{6, "Seedling"},
		{20, "Transcendent"},

		// Module 5 (51-65): 15 levels
		{51, "Seedling"},
		{65, "Transcendent"},

		// Module 8 (91-100): 10 levels
		{91, "Seedling"},
		{100, "Transcendent"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Level%d", tt.levelID), func(t *testing.T) {
			got := common.DifficultyForLevel(tt.levelID, modules)
			if got != tt.expectedDifficulty {
				t.Errorf("common.DifficultyForLevel(%d) = %s, want %s", tt.levelID, got, tt.expectedDifficulty)
			}
		})
	}
}

// TestDifficultyForLevel_LastLevelTranscendent verifies every module ends with Transcendent.
func TestDifficultyForLevel_LastLevelTranscendent(t *testing.T) {
	modules, err := common.LoadModules("")
	if err != nil {
		t.Fatal(err)
	}

	for _, m := range modules {
		if m.Name == "Tutorial" {
			continue // Tutorial is special
		}

		lastLevelID := m.End
		got := common.DifficultyForLevel(lastLevelID, modules)
		if got != "Transcendent" {
			t.Errorf("Module %s last level (%d) difficulty = %s, want Transcendent", m.Name, lastLevelID, got)
		}
	}
}

// TestDifficultyForLevel_FirstLevelSeedling verifies first level in each module is Seedling.
func TestDifficultyForLevel_FirstLevelSeedling(t *testing.T) {
	modules, err := common.LoadModules("")
	if err != nil {
		t.Fatal(err)
	}

	for _, m := range modules {
		if m.Name == "Tutorial" {
			continue // Tutorial is special
		}

		firstLevelID := m.Start
		got := common.DifficultyForLevel(firstLevelID, modules)
		if got != "Seedling" {
			t.Errorf("Module %s first level (%d) difficulty = %s, want Seedling", m.Name, firstLevelID, got)
		}
	}
}

// TestGenerateVines_SolverValidation ensures generated vines are solvable for easy difficulties.
// Note: Harder difficulties may timeout; use benchmarks for performance testing.
func TestGenerateVines_SolverValidation(t *testing.T) {
	tests := []struct {
		difficulty string
		levelID    int
	}{
		{"Seedling", 7},
		{"Sprout", 10},
	}

	for _, tt := range tests {
		t.Run(tt.difficulty, func(t *testing.T) {
			level := CreateGameLevel(tt.levelID, "Test", tt.difficulty, 0, 0, false, 0, false)

			solver := common.NewSolver(level)
			greedySolvable := solver.IsSolvableGreedy()

			if !greedySolvable {
				t.Errorf("Level %d (%s): greedy solver returned false", tt.levelID, tt.difficulty)
			}
		})
	}
}

// TestGenerateLevel_AllFieldsPopulated verifies all required fields are set.
func TestGenerateLevel_AllFieldsPopulated(t *testing.T) {
	level := CreateGameLevel(42, "Test Level", "Nurturing", 0, 0, false, 0, false)

	if level.ID == 0 {
		t.Error("Level.ID not set")
	}
	if level.Name == "" {
		t.Error("Level.Name not set")
	}
	if level.Difficulty == "" {
		t.Error("Level.Difficulty not set")
	}
	if level.GridSize[0] == 0 || level.GridSize[1] == 0 {
		t.Error("Level.GridSize not set")
	}
	if len(level.Vines) == 0 {
		t.Error("Level.Vines empty")
	}
	if level.MaxMoves == 0 {
		t.Error("Level.MaxMoves not set")
	}
	if level.MinMoves == 0 {
		t.Error("Level.MinMoves not set")
	}
	if level.Grace == 0 {
		t.Error("Level.Grace not set")
	}
	if level.Complexity == "" {
		t.Error("Level.Complexity not set")
	}
	if level.Mask == nil {
		t.Error("Level.Mask not set")
	}
}

// TestGenerateLevel_Occupancy verifies grid occupancy meets specifications.
func TestGenerateLevel_Occupancy(t *testing.T) {
	tests := []struct {
		difficulty   string
		levelID      int
		minOccupancy float64
	}{
		{"Seedling", 7, 0.30},
		{"Nurturing", 15, 0.60},
		{"Flourishing", 18, 0.75},
		{"Transcendent", 20, 0.75},
	}

	for _, tt := range tests {
		t.Run(tt.difficulty, func(t *testing.T) {
			level := CreateGameLevel(tt.levelID, "Test", tt.difficulty, 0, 0, false, 0, false)

			occupied := level.GetOccupiedCells()
			total := level.GetTotalCells()
			actual := float64(occupied) / float64(total)

			if actual < tt.minOccupancy {
				t.Errorf("Level %d occupancy %.1f%% < required %.1f%%", tt.levelID, actual*100, tt.minOccupancy*100)
			}
		})
	}
}

// TestGenerateVines_VinePathValidity checks that all vine paths are contiguous and valid.
func TestGenerateVines_VinePathValidity(t *testing.T) {
	gridSize := [2]int{12, 16}
	vines, _, err := generateVines(gridSize, "Nurturing", 12345, 0, false)
	if err != nil {
		t.Fatal(err)
	}

	for _, vine := range vines {
		// Check minimum length
		if len(vine.OrderedPath) < 2 {
			t.Errorf("Vine %s has too few segments: %d", vine.ID, len(vine.OrderedPath))
		}

		// Check bounds
		for i, pt := range vine.OrderedPath {
			if pt.X < 0 || pt.X >= gridSize[0] || pt.Y < 0 || pt.Y >= gridSize[1] {
				t.Errorf(
					"Vine %s segment %d at (%d,%d) out of bounds for grid %dx%d",
					vine.ID,
					i,
					pt.X,
					pt.Y,
					gridSize[0],
					gridSize[1],
				)
			}
		}

		// Check contiguity (each segment adjacent to previous)
		for i := 1; i < len(vine.OrderedPath); i++ {
			prev := vine.OrderedPath[i-1]
			curr := vine.OrderedPath[i]
			dist := manhattanDist(prev, curr)
			if dist != 1 {
				t.Errorf("Vine %s segments %d and %d not contiguous (distance: %d)", vine.ID, i-1, i, dist)
			}
		}

		// Check head direction matches
		if len(vine.OrderedPath) >= 2 {
			head := vine.OrderedPath[0]
			neck := vine.OrderedPath[1]
			expectedDir := directionFromPoints(head, neck)
			if expectedDir != vine.HeadDirection {
				t.Errorf(
					"Vine %s head_direction %s doesn't match path (expected %s)",
					vine.ID,
					vine.HeadDirection,
					expectedDir,
				)
			}
		}
	}
}

// TestGenerateVines_NoOverlap checks that vines don't overlap.
func TestGenerateVines_NoOverlap(t *testing.T) {
	gridSize := [2]int{16, 20}
	vines, _, err := generateVines(gridSize, "Flourishing", 54321, 0, false)
	if err != nil {
		t.Fatal(err)
	}

	occupied := make(map[string]bool)
	for _, vine := range vines {
		for _, pt := range vine.OrderedPath {
			key := pt.String()
			if occupied[key] {
				t.Errorf("Cell %s occupied by multiple vines", key)
			}
			occupied[key] = true
		}
	}
}

// TestGenerateVines_ColorDistribution verifies color variety.
func TestGenerateVines_ColorDistribution(t *testing.T) {
	gridSize := [2]int{12, 16}
	vines, _, err := generateVines(gridSize, "Nurturing", 11111, 0, false)
	if err != nil {
		t.Fatal(err)
	}

	colorCounts := make(map[string]int)
	for _, vine := range vines {
		colorCounts[vine.VineColor]++
	}

	// Should have at least 2 colors
	if len(colorCounts) < 2 {
		t.Errorf("Expected at least 2 colors, got %d", len(colorCounts))
	}

	// Should have at most 5 colors
	if len(colorCounts) > 5 {
		t.Errorf("Expected at most 5 colors, got %d", len(colorCounts))
	}

	t.Logf("Color distribution: %v", colorCounts)
}

// TestGenerateLevel_Deterministic verifies same seed produces same level.
func TestGenerateLevel_Deterministic(t *testing.T) {
	level1 := CreateGameLevel(42, "Test", "Nurturing", 0, 0, false, 0, false)
	level2 := CreateGameLevel(42, "Test", "Nurturing", 0, 0, false, 0, false)

	if len(level1.Vines) != len(level2.Vines) {
		t.Errorf("Different vine counts: %d vs %d", len(level1.Vines), len(level2.Vines))
	}

	// Note: vine order may differ due to goroutines, but vine set should match
	// For determinism, check grid sizes match
	if level1.GridSize != level2.GridSize {
		t.Errorf("Different grid sizes: %v vs %v", level1.GridSize, level2.GridSize)
	}
}

// TestGridSizeForLevelIncreasesDifficulty checks that grid sizes increase with difficulty.
func TestGridSizeForLevelIncreasesDifficulty(t *testing.T) {
	difficulties := []string{"Seedling", "Sprout", "Nurturing", "Flourishing", "Transcendent"}
	var prevArea int

	for _, diff := range difficulties {
		gridSize := common.GridSizeForLevel(50, diff)
		area := gridSize[0] * gridSize[1]

		if prevArea > 0 && area < prevArea {
			t.Errorf("Grid area should increase with difficulty: %s (%d) < previous (%d)", diff, area, prevArea)
		}
		prevArea = area
		t.Logf("%s: %dx%d (area=%d)", diff, gridSize[0], gridSize[1], area)
	}
}

// TestVineBlocking_CalculateBlocking verifies blocking relationships are computed.
func TestVineBlocking_CalculateBlocking(t *testing.T) {
	// Generate a level with multiple vines
	level := CreateGameLevel(30, "Test", "Nurturing", 0, 0, false, 0, false)

	// All vines should have Blocks field populated
	for _, vine := range level.Vines {
		if vine.Blocks == nil {
			t.Errorf("Vine %s has nil Blocks field", vine.ID)
		}
	}

	// Some vines should block others
	hasBlocking := false
	for _, vine := range level.Vines {
		if len(vine.Blocks) > 0 {
			hasBlocking = true
			break
		}
	}

	if !hasBlocking {
		t.Log("Warning: no blocking relationships detected in level")
	}
}

// TestRandomize_PersistsSeedAndRepro ensures randomize records a seed and that seed reproduces the level when reused.
func TestRandomize_PersistsSeedAndRepro(t *testing.T) {
	// First generate with randomize=true (seed derived from time)
	levelA := CreateGameLevel(999, "RandomTest", "Nurturing", 0, 0, false, 0, true)
	if levelA.GenerationSeed == 0 {
		t.Fatalf("Expected non-zero GenerationSeed when randomize=true")
	}
	t.Logf("Randomized base seed recorded: %d (attempts=%d)\n", levelA.GenerationSeed, levelA.GenerationAttempts)

	// Re-generate twice with explicit recorded seed and compare for exact reproduction
	levelB := CreateGameLevel(999, "RandomTest", "Nurturing", 0, 0, false, levelA.GenerationSeed, false)
	levelC := CreateGameLevel(999, "RandomTest", "Nurturing", 0, 0, false, levelA.GenerationSeed, false)
	t.Logf("Reproduced using seed: %d (attempts=%d)\n", levelB.GenerationSeed, levelB.GenerationAttempts)

	// Compare B vs C as unordered sets (vine order may vary due to concurrency)
	sig := func(v common.Vine) string {
		parts := []string{v.VineColor, v.HeadDirection}
		for _, p := range v.OrderedPath {
			parts = append(parts, fmt.Sprintf("%d,%d", p.X, p.Y))
		}
		return strings.Join(parts, ";")
	}

	mapB := make(map[string]int)
	for _, v := range levelB.Vines {
		mapB[sig(v)]++
	}
	mapC := make(map[string]int)
	for _, v := range levelC.Vines {
		mapC[sig(v)]++
	}

	if len(mapB) != len(mapC) {
		t.Fatalf("Mismatch in unique vine signatures for reproduced seed: %d vs %d", len(mapB), len(mapC))
	}

	for k, cb := range mapB {
		cc := mapC[k]
		if cb != cc {
			t.Fatalf("Vine signature %s count mismatch for reproduced seed: %d vs %d", k, cb, cc)
		}
	}
}

// TestGenerateSingle_RendersAfterCreation ensures CLI render option displays a quick render right after generation.
func TestGenerateSingle_RendersAfterCreation(t *testing.T) {
	// Capture log output
	var buf strings.Builder
	log.SetWriters(&buf, nil)
	defer log.ResetWriters()

	// Use a temp output dir
	tmpDir := t.TempDir()
	// Generate single level with render enabled
	generateSingle("", 8, 8, tmpDir, false, true, nil, false, 0, false, true, "ascii", true, "")

	out := buf.String()
	if !strings.Contains(out, "Level") {
		t.Fatalf("Expected rendered output to contain header 'Level', got: %s", out)
	}
}

// Helper: compute Manhattan distance between two points.
func manhattanDist(p1, p2 common.Point) int {
	dx := p1.X - p2.X
	dy := p1.Y - p2.Y
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

// Helper: determine direction from head to neck.
func directionFromPoints(head, neck common.Point) string {
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
