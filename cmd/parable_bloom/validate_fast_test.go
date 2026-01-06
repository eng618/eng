package parable_bloom

import (
	"testing"
)

func TestFastValidateLevelCoverage_Succeeds(t *testing.T) {
	level := &Level{
		GridSize: [2]int{3, 3},
		Vines: []Vine{
			{ID: "a", HeadDirection: "left", OrderedPath: []Point{{0, 0}, {1, 0}, {2, 0}}},
			{ID: "b", HeadDirection: "left", OrderedPath: []Point{{0, 1}, {1, 1}, {2, 1}}},
			{ID: "c", HeadDirection: "left", OrderedPath: []Point{{0, 2}, {1, 2}, {2, 2}}},
		},
	}

	if err := FastValidateLevelCoverage(level); err != nil {
		t.Fatalf("expected valid level, got error: %v", err)
	}
}

func TestFastValidateLevelCoverage_FailsOverlap(t *testing.T) {
	level := &Level{
		GridSize: [2]int{2, 2},
		Vines: []Vine{
			{ID: "a", HeadDirection: "left", OrderedPath: []Point{{0, 0}, {1, 0}}},
			{ID: "b", HeadDirection: "left", OrderedPath: []Point{{0, 0}, {0, 1}}},
		},
	}

	if err := FastValidateLevelCoverage(level); err == nil {
		t.Fatalf("expected overlap error, got nil")
	}
}

func TestFastValidateLevelCoverage_FailsContiguity(t *testing.T) {
	level := &Level{
		GridSize: [2]int{3, 1},
		Vines: []Vine{
			{ID: "a", HeadDirection: "left", OrderedPath: []Point{{0, 0}, {2, 0}}},
		},
	}

	if err := FastValidateLevelCoverage(level); err == nil {
		t.Fatalf("expected contiguity error, got nil")
	}
}
