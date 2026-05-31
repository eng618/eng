package common

import (
	"testing"
)

func TestDifficultyForLevel(t *testing.T) {
	// A small set of modules to test all boundaries and special cases.
	modules := []ModuleRange{
		{ID: 1, Name: "Tutorial", Start: 1, End: 5},
		{ID: 2, Name: "ModuleA", Start: 6, End: 20},    // 15 levels (remaining = 14)
		{ID: 3, Name: "ModuleB", Start: 21, End: 25},   // 5 levels (remaining = 4)
		{ID: 4, Name: "ModuleShort", Start: 26, End: 26}, // 1 level (edge case)
	}

	tests := []struct {
		name     string
		levelID  int
		expected string
	}{
		// Tutorial Module
		{"Tutorial Start", 1, "Tutorial"},
		{"Tutorial Mid", 3, "Tutorial"},
		{"Tutorial End", 5, "Tutorial"},

		// Fallback (Level not in any module)
		{"Level Not In Any Module (Before)", 0, "Seedling"},
		{"Level Not In Any Module (After)", 100, "Seedling"},

		// ModuleA (15 levels, 6-20)
		// Total: 15. Remaining: 14.
		// Ratios:
		// level 6: pos=0, ratio=0.0 -> Seedling (<0.25)
		// level 7: pos=1, ratio=1/14 (~0.07) -> Seedling (<0.25)
		// level 8: pos=2, ratio=2/14 (~0.14) -> Seedling (<0.25)
		// level 9: pos=3, ratio=3/14 (~0.21) -> Seedling (<0.25)
		// level 10: pos=4, ratio=4/14 (~0.28) -> Sprout (<0.50)
		// level 13: pos=7, ratio=7/14 (0.50) -> Nurturing (<0.75)
		// level 16: pos=10, ratio=10/14 (~0.71) -> Nurturing (<0.75)
		// level 17: pos=11, ratio=11/14 (~0.78) -> Flourishing (>=0.75)
		// level 19: pos=13, ratio=13/14 (~0.92) -> Flourishing (>=0.75)
		// level 20: pos=14 (last) -> Transcendent
		{"ModuleA First Level (Seedling)", 6, "Seedling"},
		{"ModuleA Ratio < 0.25 (Seedling)", 9, "Seedling"},
		{"ModuleA Ratio >= 0.25, < 0.50 (Sprout)", 10, "Sprout"},
		{"ModuleA Ratio = 0.50 (Nurturing)", 13, "Nurturing"}, // 7/14 = 0.5
		{"ModuleA Ratio >= 0.50, < 0.75 (Nurturing)", 16, "Nurturing"},
		{"ModuleA Ratio >= 0.75 (Flourishing)", 17, "Flourishing"},
		{"ModuleA Almost Last Level (Flourishing)", 19, "Flourishing"},
		{"ModuleA Last Level (Transcendent)", 20, "Transcendent"},

		// ModuleB (5 levels, 21-25)
		// Total: 5. Remaining: 4.
		// level 21: pos=0, ratio=0/4=0 -> Seedling
		// level 22: pos=1, ratio=1/4=0.25 -> Sprout
		// level 23: pos=2, ratio=2/4=0.50 -> Nurturing
		// level 24: pos=3, ratio=3/4=0.75 -> Flourishing
		// level 25: pos=4 (last) -> Transcendent
		{"ModuleB Ratio = 0.00 (Seedling)", 21, "Seedling"},
		{"ModuleB Ratio = 0.25 (Sprout)", 22, "Sprout"},
		{"ModuleB Ratio = 0.50 (Nurturing)", 23, "Nurturing"},
		{"ModuleB Ratio = 0.75 (Flourishing)", 24, "Flourishing"},
		{"ModuleB Last Level (Transcendent)", 25, "Transcendent"},

		// ModuleShort (1 level, 26-26)
		// Total: 1. Remaining: 0. Last level is 26.
		{"ModuleShort Only Level (Transcendent)", 26, "Transcendent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DifficultyForLevel(tt.levelID, modules)
			if got != tt.expected {
				t.Errorf("DifficultyForLevel(%d) = %q; want %q", tt.levelID, got, tt.expected)
			}
		})
	}
}
