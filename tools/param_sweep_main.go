package main

import (
	"flag"
	"fmt"

	"github.com/eng618/eng/cmd/parable_bloom/generate"
)

func main() {
	difficulty := flag.String("difficulty", "Seedling", "difficulty tier")
	iters := flag.Int("iters", 10, "iterations per param set")
	flag.Parse()

	bestScore, bestCfg, bestProf, bestTime := generate.SweepParams(*difficulty, *iters)
	fmt.Printf("Sweep result for %s:\n", *difficulty)
	fmt.Printf("  Best score: %.2f\n", bestScore)
	fmt.Printf("  Best config: %+v\n", bestCfg)
	fmt.Printf("  Best profile (lengthMix): %+v\n", bestProf.LengthMix)
	fmt.Printf("  Avg elapsed ms: %d\n", bestTime)
}
