package parable_bloom

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestGrowFromSeed_SimpleGrow(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	occupied := make(map[string]bool)
	seed := Point{X: 0, Y: 0}
	v, occ, err := GrowFromSeed(seed, occupied, [2]int{3, 3}, 4, VarietyProfile{}, GeneratorConfig{MaxSeedRetries: 5}, rng)
	if err != nil {
		t.Fatalf("expected successful grow, got error: %v", err)
	}
	if v.Length() != 4 {
		t.Fatalf("expected length 4, got %d", v.Length())
	}
	// check occupancy map contains all path coords
	for _, p := range v.OrderedPath {
		k := fmt.Sprintf("%d,%d", p.X, p.Y)
		if !occ[k] {
			t.Fatalf("expected occupied key %s", k)
		}
	}
}

func TestTileGridIntoVines_FullCoverage(t *testing.T) {
	rng := rand.New(rand.NewSource(99))
	cfg := GeneratorConfig{MaxSeedRetries: 20, LocalRepairRadius: 2, RepairRetries: 3}
	constraints := DifficultySpecs["Seedling"]
	vines, err := TileGridIntoVines([2]int{5, 4}, constraints, VarietyProfile{}, cfg, rng)
	if err != nil {
		t.Fatalf("tile generation failed: %v", err)
	}
	level := &Level{GridSize: [2]int{5, 4}, Vines: vines}
	if err := FastValidateLevelCoverage(level); err != nil {
		t.Fatalf("final validation failed: %v", err)
	}
}

func BenchmarkTileGridIntoVines(b *testing.B) {
	cfg := GeneratorConfig{MaxSeedRetries: 20, LocalRepairRadius: 2, RepairRetries: 3}
	constraints := DifficultySpecs["Seedling"]
	for i := 0; i < b.N; i++ {
		rng := rand.New(rand.NewSource(int64(i)))
		_, err := TileGridIntoVines([2]int{12, 8}, constraints, VarietyProfile{}, cfg, rng)
		if err != nil {
			b.Fatalf("tile generation failed: %v", err)
		}
	}
}
