package parable_bloom

import (
	"math/rand"
	"testing"
)

func TestParamSweep_Smoke(t *testing.T) {
	// Smoke test: run a tiny sweep configuration and ensure it completes
	spec := DifficultySpecs["Seedling"]
	prof := VarietyProfile{LengthMix: map[string]float64{"short": 0.3, "medium": 0.4, "long": 0.3}}
	cfg := GeneratorConfig{MaxSeedRetries: 8, LocalRepairRadius: 1, RepairRetries: 1}
	rng := rand.New(rand.NewSource(42))
	res := GenerateWithProfile([2]int{8, 9}, spec, prof, cfg, 42, false, rng)
	if len(res.Vines) == 0 {
		t.Fatalf("expected non-empty result from GenerateWithProfile in sweep smoke, got %+v", res)
	}
}
