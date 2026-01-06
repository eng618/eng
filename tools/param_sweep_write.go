package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/eng618/eng/cmd/parable_bloom"
)

// WriteSweepToFile runs a parameter sweep and writes the best result to outPath in JSON format.
func WriteSweepToFile(difficulty string, iters int, outPath string) error {
	bestScore, bestCfg, bestProf, bestTime := parable_bloom.SweepParams(difficulty, iters)
	res := map[string]interface{}{
		"difficulty": difficulty,
		"bestScore":  bestScore,
		"bestCfg":    bestCfg,
		"bestProf":   bestProf.LengthMix,
		"bestTime":   bestTime,
	}
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(res); err != nil {
		return err
	}
	fmt.Printf("Wrote results to %s\n", outPath)
	return nil
}
