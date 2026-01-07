package parable_bloom

import (
	"testing"
)

// Fast CI perf checks: ensure tiling-first generator keeps acceptable runtime for common tiers.
func TestPerf_Tiling_Seedling(t *testing.T) {
	// Keep it small and stable: run a few seeds and average
	N := 5
	total := int64(0)
	for i := 0; i < N; i++ {
		spec := DifficultySpecs["Seedling"]
		cfg := GetGeneratorConfigForDifficulty("Seedling")
		res := GenerateWithProfile([2]int{8, 9}, spec, GetPresetProfile("Seedling"), cfg, int64(i+1), false, nil)
		if len(res.Vines) == 0 {
			t.Fatalf("generation failed on seed %d: %+v", i, res)
		}
		total += res.ElapsedMS
	}
	avg := total / int64(N)
	// Threshold: 25ms average per attempt (CI machines can vary)
	if avg > 25 {
		t.Fatalf("tiling-first average elapsed ms too high for Seedling: %dms (threshold 25ms)", avg)
	}
}

func TestPerf_Tiling_Nurturing(t *testing.T) {
	N := 3
	total := int64(0)
	for i := 0; i < N; i++ {
		spec := DifficultySpecs["Nurturing"]
		cfg := GetGeneratorConfigForDifficulty("Nurturing")
		res := GenerateWithProfile(GridSizeForLevel(15, "Nurturing"), spec, GetPresetProfile("Nurturing"), cfg, int64(i+100), false, nil)
		if len(res.Vines) == 0 {
			t.Fatalf("generation failed on seed %d: %+v", i, res)
		}
		total += res.ElapsedMS
	}
	avg := total / int64(N)
	// Threshold: 200ms average (adjusted for CI variability)
	if avg > 200 {
		t.Fatalf("tiling-first average elapsed ms too high for Nurturing: %dms (threshold 200ms)", avg)
	}
}
