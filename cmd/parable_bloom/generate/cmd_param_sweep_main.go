//go:build ignore
// +build ignore

package generate

import (
	"flag"
	"fmt"
)

// Simple parameter sweep CLI. Run with `go run cmd/parable_bloom/cmd_param_sweep.go -difficulty Seedling -iters 5`.
func main() {
	difficulty := flag.String("difficulty", "Seedling", "difficulty tier to sweep")
	iters := flag.Int("iters", 5, "number of random seeds per parameter set")
	flag.Parse()

	bestScore, bestCfg, bestProf, bestTime := SweepParams(*difficulty, *iters)
	fmt.Printf("Sweep result for %s:\n", *difficulty)
	fmt.Printf("  Best score: %.2f\n", bestScore)
	fmt.Printf("  Best config: %+v\n", bestCfg)
	fmt.Printf("  Best profile (lengthMix): %+v\n", bestProf.LengthMix)
	fmt.Printf("  Avg elapsed ms: %d\n", bestTime)
}
