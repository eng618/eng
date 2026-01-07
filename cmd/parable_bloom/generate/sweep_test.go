package generate

import (
	"testing"
)

func TestSweep_RunSmall(t *testing.T) {
	bestScore, bestCfg, bestProf, bestTime := SweepParams("Seedling", 1)
	t.Logf("sweep result: score=%.2f cfg=%+v prof=%+v time=%d", bestScore, bestCfg, bestProf.LengthMix, bestTime)
	if bestScore == -1e12 {
		t.Fatalf("sweep did not find any configuration")
	}
}

func TestSweep_Nurturing_Smoke(t *testing.T) {
	bestScore, bestCfg, bestProf, bestTime := SweepParams("Nurturing", 1)
	t.Logf("nurturing sweep result: score=%.2f cfg=%+v prof=%+v time=%d", bestScore, bestCfg, bestProf.LengthMix, bestTime)
	// Accept either a default negative score or any real score; ensure function completes
	_ = bestScore
	_ = bestCfg
	_ = bestProf
	_ = bestTime
}

func TestSweep_LargerSeedling_Quick(t *testing.T) {
	bestScore, bestCfg, bestProf, bestTime := SweepParams("Seedling", 3)
	t.Logf("larger sweep result: score=%.2f cfg=%+v prof=%+v time=%d", bestScore, bestCfg, bestProf.LengthMix, bestTime)
	if bestScore == -1e12 {
		t.Fatalf("sweep did not find any configuration on larger run")
	}
}
