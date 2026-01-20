//go:build ignore
// +build ignore

package generate

import (
	"math"
	"math/rand"
	"testing"

	"github.com/eng618/eng/cmd/parable_bloom/common"
)

func TestParableBloomTestsRemoved(t *testing.T) {
	t.Skip("Parable Bloom eng CLI tests removed; use tools/level-builder tests instead.")
}

func avgVineLength(vines []common.Vine) float64 {
	total := 0
	for _, v := range vines {
		total += v.Length()
	}
	return float64(total) / float64(len(vines))
}

func avgTurnCount(vines []common.Vine) float64 {
	total := 0
	for _, v := range vines {
		if v.Length() < 3 {
			continue
		}
		turns := 0
		for i := 1; i < len(v.OrderedPath)-1; i++ {
			p := v.OrderedPath[i-1]
			n := v.OrderedPath[i]
			q := v.OrderedPath[i+1]
			dx1 := n.X - p.X
			dy1 := n.Y - p.Y
			dx2 := q.X - n.X
			dy2 := q.Y - n.Y
			if dx1 != dx2 || dy1 != dy2 {
				turns++
			}
		}
		total += turns
	}
	return float64(total) / float64(len(vines))
}

func avgDistanceToCenter(vines []common.Vine, gridSize [2]int) float64 {
	total := 0.0
	count := 0
	cx := float64(gridSize[0]-1) / 2.0
	cy := float64(gridSize[1]-1) / 2.0
	for _, v := range vines {
		for _, p := range v.OrderedPath {
			dx := float64(p.X) - cx
			dy := float64(p.Y) - cy
			d := math.Sqrt(dx*dx + dy*dy)
			total += d
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return total / float64(count)
}

func TestVarietyProfile_LengthMixInfluence(t *testing.T) {
	gridSize := [2]int{8, 8}
	spec := common.DifficultySpecs["Seedling"]
	cfg := common.GeneratorConfig{MaxSeedRetries: 20, LocalRepairRadius: 2, RepairRetries: 3}

	longProf := common.VarietyProfile{
		LengthMix: map[string]float64{"short": 0.1, "medium": 0.3, "long": 0.6},
		TurnMix:   0.2,
	}
	shortProf := common.VarietyProfile{
		LengthMix: map[string]float64{"short": 0.6, "medium": 0.3, "long": 0.1},
		TurnMix:   0.6,
	}

	rng1 := rand.New(rand.NewSource(100))
	v1, _, err := TileGridIntoVines(gridSize, spec, longProf, cfg, rng1)
	if err != nil {
		t.Fatalf("tiling failed: %v", err)
	}
	rng2 := rand.New(rand.NewSource(101))
	v2, _, err := TileGridIntoVines(gridSize, spec, shortProf, cfg, rng2)
	if err != nil {
		t.Fatalf("tiling failed: %v", err)
	}

	if avgVineLength(v1) <= avgVineLength(v2) {
		t.Fatalf(
			"expected long profile average length > short profile (%.2f <= %.2f)",
			avgVineLength(v1),
			avgVineLength(v2),
		)
	}
}

func TestVarietyProfile_TurnMixInfluence(t *testing.T) {
	gridSize := [2]int{10, 10}
	spec := common.DifficultySpecs["Seedling"]
	cfg := common.GeneratorConfig{MaxSeedRetries: 50, LocalRepairRadius: 3, RepairRetries: 4}

	highTurn := common.VarietyProfile{
		LengthMix: map[string]float64{"short": 0.3, "medium": 0.4, "long": 0.3},
		TurnMix:   0.9,
	}
	lowTurn := common.VarietyProfile{
		LengthMix: map[string]float64{"short": 0.3, "medium": 0.4, "long": 0.3},
		TurnMix:   0.1,
	}

	rng1 := rand.New(rand.NewSource(200))
	vh, _, err := TileGridIntoVines(gridSize, spec, highTurn, cfg, rng1)
	if err != nil {
		t.Fatalf("tiling failed: %v", err)
	}
	rng2 := rand.New(rand.NewSource(201))
	vl, _, err := TileGridIntoVines(gridSize, spec, lowTurn, cfg, rng2)
	if err != nil {
		t.Fatalf("tiling failed: %v", err)
	}

	if avgTurnCount(vh) <= avgTurnCount(vl) {
		t.Fatalf("expected high turn mix to have more turns (%.2f <= %.2f)", avgTurnCount(vh), avgTurnCount(vl))
	}
}

func TestVarietyProfile_RegionBias(t *testing.T) {
	gridSize := [2]int{12, 8}
	spec := common.DifficultySpecs["Sprout"]
	cfg := common.GeneratorConfig{MaxSeedRetries: 50, LocalRepairRadius: 3, RepairRetries: 4}

	edgeProf := common.VarietyProfile{RegionBias: "edge"}
	centerProf := common.VarietyProfile{RegionBias: "center"}

	rng1 := rand.New(rand.NewSource(300))
	edgeV, _, err := TileGridIntoVines(gridSize, spec, edgeProf, cfg, rng1)
	if err != nil {
		t.Fatalf("tiling failed: %v", err)
	}
	rng2 := rand.New(rand.NewSource(301))
	centerV, _, err := TileGridIntoVines(gridSize, spec, centerProf, cfg, rng2)
	if err != nil {
		t.Fatalf("tiling failed: %v", err)
	}

	dEdge := avgDistanceToCenter(edgeV, gridSize)
	dCenter := avgDistanceToCenter(centerV, gridSize)
	if dEdge <= dCenter {
		t.Fatalf("expected edge-biased average distance > center-biased (%.2f <= %.2f)", dEdge, dCenter)
	}
}
