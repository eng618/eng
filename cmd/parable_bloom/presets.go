package parable_bloom

// GetPresetProfile returns a VarietyProfile tuned for the given difficulty tier.
func GetPresetProfile(difficulty string) VarietyProfile {
	spec := DifficultySpecs[difficulty]
	minL, maxL := spec.AvgLengthRange[0], spec.AvgLengthRange[1]
	median := (minL + maxL) / 2

	lengthMix := map[string]float64{"short": 0.33, "medium": 0.33, "long": 0.34}
	turnMix := 0.35
	regionBias := "balanced"
	dirBalance := map[string]float64{"right": 0.25, "left": 0.25, "up": 0.25, "down": 0.25}

	// Adjust length mix depending on median length
	if median >= 6 {
		// Favor longer vines
		lengthMix = map[string]float64{"short": 0.15, "medium": 0.35, "long": 0.5}
		turnMix = 0.25
		regionBias = "edge"
	} else if median <= 4 {
		// Favor shorter vines
		lengthMix = map[string]float64{"short": 0.6, "medium": 0.3, "long": 0.1}
		turnMix = 0.5
		regionBias = "center"
	} else {
		// Medium lengths
		lengthMix = map[string]float64{"short": 0.3, "medium": 0.5, "long": 0.2}
		turnMix = 0.4
		regionBias = "balanced"
	}

	return VarietyProfile{
		LengthMix:  lengthMix,
		TurnMix:    turnMix,
		RegionBias: regionBias,
		DirBalance: dirBalance,
	}
}
