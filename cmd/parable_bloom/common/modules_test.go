package common

import (
	"math"
	"testing"
)

func TestGridSizeForLevel(t *testing.T) {
	t.Run("Determinism", func(t *testing.T) {
		levelID := 42
		difficulty := "Flourishing"
		expected := GridSizeForLevel(levelID, difficulty)

		for i := 0; i < 10; i++ {
			result := GridSizeForLevel(levelID, difficulty)
			if result != expected {
				t.Errorf(
					"Expected deterministic result %v, got %v on iteration %d",
					expected,
					result,
					i,
				)
			}
		}
	})

	t.Run("Bounds checking across difficulties", func(t *testing.T) {
		difficulties := []string{
			"Tutorial",
			"Seedling",
			"Sprout",
			"Nurturing",
			"Flourishing",
			"Transcendent",
		}

		for _, diff := range difficulties {
			t.Run(diff, func(t *testing.T) {
				ranges := GridSizeRanges[diff]

				// Test several levels to ensure coverage within bounds
				for levelID := 1; levelID <= 100; levelID++ {
					result := GridSizeForLevel(levelID, diff)
					width, height := result[0], result[1]

					if width < ranges.MinW || width > ranges.MaxW {
						t.Errorf(
							"Difficulty %s, Level %d: Width %d is out of bounds [%d, %d]",
							diff,
							levelID,
							width,
							ranges.MinW,
							ranges.MaxW,
						)
					}
					if height < ranges.MinH || height > ranges.MaxH {
						t.Errorf(
							"Difficulty %s, Level %d: Height %d is out of bounds [%d, %d]",
							diff,
							levelID,
							height,
							ranges.MinH,
							ranges.MaxH,
						)
					}
				}
			})
		}
	})

	t.Run("Fallback for unknown difficulty", func(t *testing.T) {
		levelID := 15
		unknownResult := GridSizeForLevel(levelID, "UnknownDifficulty")
		seedlingResult := GridSizeForLevel(levelID, "Seedling")

		if unknownResult != seedlingResult {
			t.Errorf(
				"Expected unknown difficulty to fallback to Seedling. Got %v, expected %v",
				unknownResult,
				seedlingResult,
			)
		}

		ranges := GridSizeRanges["Seedling"]
		width, height := unknownResult[0], unknownResult[1]
		if width < ranges.MinW || width > ranges.MaxW || height < ranges.MinH ||
			height > ranges.MaxH {
			t.Errorf(
				"Fallback result out of Seedling bounds. Result: %v, Bounds: [%d-%d, %d-%d]",
				unknownResult,
				ranges.MinW,
				ranges.MaxW,
				ranges.MinH,
				ranges.MaxH,
			)
		}
	})

	t.Run("Edge Cases: Level ID clamping", func(t *testing.T) {
		// Negative levelID should be clamped to 0
		negResult := GridSizeForLevel(-5, "Seedling")
		zeroResult := GridSizeForLevel(0, "Seedling")
		if negResult != zeroResult {
			t.Errorf(
				"Expected levelID < 0 to be clamped to 0. Got %v, expected %v",
				negResult,
				zeroResult,
			)
		}

		// Extremely large levelID should be clamped or wrapped without panicking
		// Note: The implementation clamps to math.MaxUint32 (if it exceeds int(math.MaxUint32), which happens on 64-bit systems)
		// On 32-bit systems, int is 32-bit, so it might not exceed int(math.MaxUint32).
		// We'll test with math.MaxInt and math.MaxInt32 and ensure they don't panic and return valid sizes
		largeResults := [][2]int{
			GridSizeForLevel(math.MaxInt32, "Sprout"),
			GridSizeForLevel(math.MaxInt, "Sprout"),
		}

		ranges := GridSizeRanges["Sprout"]
		for _, result := range largeResults {
			width, height := result[0], result[1]
			if width < ranges.MinW || width > ranges.MaxW || height < ranges.MinH ||
				height > ranges.MaxH {
				t.Errorf(
					"Edge case result out of bounds. Result: %v, Bounds: [%d-%d, %d-%d]",
					result,
					ranges.MinW,
					ranges.MaxW,
					ranges.MinH,
					ranges.MaxH,
				)
			}
		}
	})
}
