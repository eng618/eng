package common

import "testing"

func TestGetPresetProfile_BucketBiases(t *testing.T) {
	seed := GetPresetProfile("Seedling")
	trans := GetPresetProfile("Transcendent")

	if seed.LengthMix["long"] <= trans.LengthMix["long"] {
		t.Fatalf(
			"expected Seedling to favor long vines more than Transcendent: seed.long=%f trans.long=%f",
			seed.LengthMix["long"],
			trans.LengthMix["long"],
		)
	}

	if trans.LengthMix["short"] <= seed.LengthMix["short"] {
		t.Fatalf(
			"expected Transcendent to favor short vines more than Seedling: trans.short=%f seed.short=%f",
			trans.LengthMix["short"],
			seed.LengthMix["short"],
		)
	}
}

func TestGetPresetProfile_DirBalanceExists(t *testing.T) {
	p := GetPresetProfile("Nurturing")
	sum := p.DirBalance["up"] + p.DirBalance["down"] + p.DirBalance["left"] + p.DirBalance["right"]
	if sum <= 0 {
		t.Fatalf("expected dir balance to sum > 0, got %f", sum)
	}
}

func TestGetGeneratorConfigForDifficulty(t *testing.T) {
	cSeed := GetGeneratorConfigForDifficulty("Seedling")
	cTrans := GetGeneratorConfigForDifficulty("Transcendent")
	if cSeed.MaxSeedRetries >= cTrans.MaxSeedRetries {
		t.Fatalf(
			"expected Transcendent to have higher MaxSeedRetries than Seedling: %d >= %d",
			cSeed.MaxSeedRetries,
			cTrans.MaxSeedRetries,
		)
	}
}
