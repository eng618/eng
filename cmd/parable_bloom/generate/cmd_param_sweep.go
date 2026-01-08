package generate

import (
	"fmt"
	"time"

	"github.com/eng618/eng/cmd/parable_bloom/common"
)

// cfgProf pairs a generator config with a variety profile for sweeps.
type cfgProf struct {
	cfg  common.GeneratorConfig
	prof common.VarietyProfile
}

// SweepParams runs the parameter sweep and returns the best results.
func SweepParams(difficulty string, iters int) (float64, common.GeneratorConfig, common.VarietyProfile, int64) {
	spec := common.DifficultySpecs[difficulty]
	bestScore := -1e12
	bestCfg := common.GeneratorConfig{}
	bestProf := common.VarietyProfile{}
	bestTime := int64(0)

	lengthPresets := []common.VarietyProfile{
		{LengthMix: map[string]float64{"short": 0.1, "medium": 0.3, "long": 0.6}},
		{LengthMix: map[string]float64{"short": 0.6, "medium": 0.3, "long": 0.1}},
		{LengthMix: map[string]float64{"short": 0.3, "medium": 0.4, "long": 0.3}},
	}
	// Reduce search space for very small iters (used by smoke tests) to keep runtime low.
	seedRetries := []int{8, 16, 32}
	repairRadius := []int{1, 2, 3}
	repairRetries := []int{1, 2, 4}
	if iters <= 1 {
		seedRetries = []int{8, 32}
		repairRadius = []int{1, 2}
		repairRetries = []int{1, 2}
	}

	// Build candidate configurations and evaluate them
	candidates := buildCandidates(seedRetries, repairRadius, repairRetries, lengthPresets)
	for _, item := range candidates {
		avgScore, avgTime := evaluateConfig(difficulty, spec, item.prof, item.cfg, iters)
		// prefer higher score, then lower time
		if avgScore > bestScore || (avgScore == bestScore && (bestTime == 0 || avgTime < bestTime)) {
			bestScore = avgScore
			bestCfg = item.cfg
			bestProf = item.prof
			bestTime = avgTime
			fmt.Printf(
				"New best: score=%.2f cfg=%+v prof=%+v time=%d\n",
				bestScore,
				bestCfg,
				bestProf.LengthMix,
				bestTime,
			)
		}
	}

	return bestScore, bestCfg, bestProf, bestTime
}

func buildCandidates(seedRetries, repairRadius, repairRetries []int, lengthPresets []common.VarietyProfile) []cfgProf {
	candidates := []cfgProf{}
	for _, lr := range seedRetries {
		for _, rr := range repairRadius {
			for _, rtries := range repairRetries {
				for _, prof := range lengthPresets {
					candidates = append(
						candidates,
						cfgProf{
							cfg: common.GeneratorConfig{
								MaxSeedRetries:    lr,
								LocalRepairRadius: rr,
								RepairRetries:     rtries,
							},
							prof: prof,
						},
					)
				}
			}
		}
	}
	return candidates
}

// evaluateConfig runs `iters` generation attempts and returns average score and average elapsed ms.
func evaluateConfig(
	difficulty string,
	spec common.DifficultySpec,
	prof common.VarietyProfile,
	cfg common.GeneratorConfig,
	iters int,
) (float64, int64) {
	totalScore := 0.0
	totalTime := int64(0)
	for i := 0; i < iters; i++ {
		seed := time.Now().UnixNano() + int64(i)
		res := GenerateWithProfile(common.GridSizeForLevel(7, difficulty), spec, prof, cfg, seed, false, nil)
		totalScore += res.Score
		totalTime += res.ElapsedMS
	}
	avgScore := totalScore / float64(iters)
	avgTime := totalTime / int64(iters)
	return avgScore, avgTime
}
