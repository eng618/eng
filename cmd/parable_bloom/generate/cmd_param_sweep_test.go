package generate

import (
	"math/rand"
	"testing"

	"github.com/eng618/eng/cmd/parable_bloom/common"
)

func TestParamSweep_Smoke(t *testing.T) {
	// Smoke test: run a tiny sweep configuration and ensure it completes
	spec := common.DifficultySpecs["Seedling"]
	prof := common.VarietyProfile{LengthMix: map[string]float64{"short": 0.3, "medium": 0.4, "long": 0.3}}
	cfg := common.GeneratorConfig{MaxSeedRetries: 8, LocalRepairRadius: 1, RepairRetries: 1}
	rng := rand.New(rand.NewSource(42))
	res := CreateLevelWithProfile([2]int{8, 9}, spec, prof, cfg, 42, false, rng)
	if len(res.Vines) == 0 {
		t.Fatalf("expected non-empty result from GenerateWithProfile in sweep smoke, got %+v", res)
	}
}
