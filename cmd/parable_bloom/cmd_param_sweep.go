package parable_bloom

import (
	"flag"
	"fmt"
	"math/rand"
	"time"
)

// Simple parameter sweep CLI. Run with `go run cmd/parable_bloom/cmd_param_sweep.go -difficulty Seedling -iters 5`.
func main() {
	difficulty := flag.String("difficulty", "Seedling", "difficulty tier to sweep")
	iters := flag.Int("iters", 5, "number of random seeds per parameter set")
	flag.Parse()

	spec := DifficultySpecs[*difficulty]
	bestScore := -1e12
	bestCfg := GeneratorConfig{}
	bestProf := VarietyProfile{}
	bestTime := int64(0)

	lengthPresets := []VarietyProfile{
		{LengthMix: map[string]float64{"short": 0.1, "medium": 0.3, "long": 0.6}},
		{LengthMix: map[string]float64{"short": 0.6, "medium": 0.3, "long": 0.1}},
		{LengthMix: map[string]float64{"short": 0.3, "medium": 0.4, "long": 0.3}},
	}
	seedRetries := []int{8, 16, 32}
	repairRadius := []int{1, 2, 3}
	repairRetries := []int{1, 2, 4}

	for _, lr := range seedRetries {
		for _, rr := range repairRadius {
			for _, rtries := range repairRetries {
				for _, prof := range lengthPresets {
					cfg := GeneratorConfig{MaxSeedRetries: lr, LocalRepairRadius: rr, RepairRetries: rtries}
					// evaluate
					totalScore := 0.0
					totalTime := int64(0)
					for i := 0; i < *iters; i++ {
						rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(i)))
						res := GenerateWithProfile(GridSizeForLevel(7, *difficulty), spec, prof, cfg, int64(i+1), false, rng)
						totalScore += res.Score
						totalTime += res.ElapsedMS
					}
					avgScore := totalScore / float64(*iters)
					avgTime := totalTime / int64(*iters)
					// prefer higher score, then lower time
					if avgScore > bestScore || (avgScore == bestScore && (bestTime == 0 || avgTime < bestTime)) {
						bestScore = avgScore
						bestCfg = cfg
						bestProf = prof
						bestTime = avgTime
					}
				}
			}
		}
	}

	fmt.Printf("Sweep result for %s:\n", *difficulty)
	fmt.Printf("  Best score: %.2f\n", bestScore)
	fmt.Printf("  Best config: %+v\n", bestCfg)
	fmt.Printf("  Best profile (lengthMix): %+v\n", bestProf.LengthMix)
	fmt.Printf("  Avg elapsed ms: %d\n", bestTime)
}
