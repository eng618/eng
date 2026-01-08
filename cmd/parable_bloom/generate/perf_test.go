package generate

import (
	"testing"

	"github.com/eng618/eng/cmd/parable_bloom/common"
)

// Fast CI perf checks: ensure tiling-first generator keeps acceptable runtime for common tiers.
func TestPerf_Tiling_Seedling(t *testing.T) {
	// Keep it small and stable: run a few seeds and average
	N := 5
	total := int64(0)
	for i := 0; i < N; i++ {
		spec := common.DifficultySpecs["Seedling"]
		cfg := common.GetGeneratorConfigForDifficulty("Seedling")
		res := GenerateWithProfile([2]int{8, 9}, spec, common.GetPresetProfile("Seedling"), cfg, int64(i+1), false, nil)
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
		spec := common.DifficultySpecs["Nurturing"]
		cfg := common.GetGeneratorConfigForDifficulty("Nurturing")
		res := GenerateWithProfile(
			common.GridSizeForLevel(15, "Nurturing"),
			spec,
			common.GetPresetProfile("Nurturing"),
			cfg,
			int64(i+100),
			false,
			nil,
		)
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
