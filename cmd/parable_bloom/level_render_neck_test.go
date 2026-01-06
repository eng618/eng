package parable_bloom

import (
	"bytes"
	"strings"
	"testing"
)

func TestRenderNeckMarkerUnicode(t *testing.T) {
	level := &Level{
		ID:       101,
		Name:     "Neck Test",
		GridSize: [2]int{5, 3},
		Vines: []Vine{
			{ID: "vine_0", HeadDirection: "right", OrderedPath: []Point{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}}},
		},
	}

	var buf bytes.Buffer
	renderLevelToWriter(&buf, level, "unicode", false)
	out := buf.String()
	// Neck should appear as a regular horizontal segment '─' in unicode mode
	if !strings.Contains(out, "─") {
		t.Fatalf("expected horizontal '─' in output:\n%s", out)
	}
}

func TestRenderNeckMarkerAscii(t *testing.T) {
	level := &Level{
		ID:       102,
		Name:     "Neck Test ASCII",
		GridSize: [2]int{5, 3},
		Vines: []Vine{
			{ID: "vine_0", HeadDirection: "right", OrderedPath: []Point{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}}},
		},
	}

	var buf bytes.Buffer
	renderLevelToWriter(&buf, level, "ascii", false)
	out := buf.String()
	// Neck should appear as a regular horizontal segment '-' in ascii mode
	if !strings.Contains(out, "-") {
		t.Fatalf("expected horizontal '-' in output:\n%s", out)
	}
}
