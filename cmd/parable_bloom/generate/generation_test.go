package generate

import (
	"math/rand"
	"testing"

	"github.com/eng618/eng/cmd/parable_bloom/common"
)

func TestFastScoreBlocking_Simple(t *testing.T) {
	// Two vines: A blocks B
	vines := []common.Vine{
		{ID: "a", HeadDirection: "right", OrderedPath: []common.Point{{X: 0, Y: 0}, {X: 1, Y: 0}}},
		{ID: "b", HeadDirection: "left", OrderedPath: []common.Point{{X: 2, Y: 0}, {X: 1, Y: 0}}},
	}
	score, maxDepth := FastScoreBlocking(vines, [2]int{3, 1})
	if maxDepth == 0 {
		t.Fatalf("expected maxDepth > 0, got %d", maxDepth)
	}
	if score != -1000000.0 {
		t.Fatalf("expected strongly negative score for cyclic blocking, got %f", score)
	}
}

func TestGenerateWithProfile_Tiling(t *testing.T) {
	gridSize := [2]int{5, 4}
	spec := common.DifficultySpecs["Seedling"]
	cfg := common.GeneratorConfig{MaxSeedRetries: 20, LocalRepairRadius: 2, RepairRetries: 3}
	rng := rand.New(rand.NewSource(7))
	res := GenerateWithProfile(gridSize, spec, common.VarietyProfile{}, cfg, 7, false, rng)
	if !res.GreedySolvable {
		t.Fatalf("expected tiling result to be greedy-solvable, got %+v", res)
	}
	// validate structure
	lvl := &common.Level{GridSize: gridSize, Vines: res.Vines}
	if err := common.FastValidateLevelCoverage(lvl); err != nil {
		t.Fatalf("fast validation failed: %v", err)
	}
}

func TestGenerateWithProfile_Telemetry(t *testing.T) {
	gridSize := [2]int{6, 6}
	spec := common.DifficultySpecs["Seedling"]
	cfg := common.GetGeneratorConfigForDifficulty("Seedling")
	rng := rand.New(rand.NewSource(42))
	res := GenerateWithProfile(gridSize, spec, common.GetPresetProfile("Seedling"), cfg, 42, false, rng)
	if len(res.Vines) == 0 {
		t.Fatalf("expected generation to succeed, got empty result: %+v", res)
	}
	if res.Attempts <= 0 {
		t.Fatalf("expected Attempts>0, got %d", res.Attempts)
	}
	if res.SeedUsed == 0 {
		t.Fatalf("expected SeedUsed to be set, got 0")
	}
	if res.ElapsedMS < 0 {
		t.Fatalf("expected ElapsedMS>=0, got %d", res.ElapsedMS)
	}
}
