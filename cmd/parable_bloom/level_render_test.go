package parable_bloom

import (
	"bytes"
	"strings"
	"testing"
)

func TestRenderLevelAscii(t *testing.T) {
	level := &Level{
		ID:       42,
		Name:     "Test Level",
		GridSize: [2]int{5, 4},
		Vines: []Vine{
			{
				ID:            "vine_0",
				HeadDirection: "right",
				OrderedPath:   []Point{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}},
			},
			{
				ID:            "vine_1",
				HeadDirection: "up",
				OrderedPath:   []Point{{X: 4, Y: 3}, {X: 4, Y: 2}},
			},
		},
	}

	var buf bytes.Buffer
	renderLevelToWriter(&buf, level, "ascii", true)
	out := buf.String()
	if !strings.Contains(out, "Level 42: Test Level (grid 5x4)") {
		t.Fatalf("missing header in output:\n%s", out)
	}
	if !strings.Contains(out, ">") {
		t.Fatalf("expected head arrow '>' for right heading vine: \n%s", out)
	}
	if !strings.Contains(out, "^") {
		t.Fatalf("expected head arrow '^' or '^' equivalent for up heading vine: \n%s", out)
	}
	// Default (connectors=true) should show ASCII connectors
	if !strings.Contains(out, "-") && !strings.Contains(out, "|") && !strings.Contains(out, "+") {
		t.Fatalf("expected ASCII connector ('-', '|', '+') in output:\n%s", out)
	}
}

func TestRenderLevelUnicode(t *testing.T) {
	level := &Level{
		ID:       99,
		Name:     "Unicode Level",
		GridSize: [2]int{3, 3},
		Vines: []Vine{
			{ID: "vine_0", HeadDirection: "left", OrderedPath: []Point{{X: 2, Y: 0}, {X: 1, Y: 0}}},
		},
	}

	var buf bytes.Buffer
	renderLevelToWriter(&buf, level, "unicode", false)
	out := buf.String()
	if !strings.Contains(out, "←") {
		t.Fatalf("expected unicode left arrow in output:\n%s", out)
	}
	// Default (connectors=true) should show box-drawing or tail marker
	if !strings.Contains(out, "─") && !strings.Contains(out, "│") && !strings.Contains(out, "┌") && !strings.Contains(out, "└") && !strings.Contains(out, "┐") && !strings.Contains(out, "┘") && !strings.Contains(out, "●") {
		t.Fatalf("expected box-drawing or tail marker in output:\n%s", out)
	}
}

func TestRenderLevelAsciiWithConnectors(t *testing.T) {
	level := &Level{
		ID:       42,
		Name:     "Test Level",
		GridSize: [2]int{5, 4},
		Vines: []Vine{
			{
				ID:            "vine_0",
				HeadDirection: "right",
				OrderedPath:   []Point{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}},
			},
			{
				ID:            "vine_1",
				HeadDirection: "up",
				OrderedPath:   []Point{{X: 4, Y: 3}, {X: 4, Y: 2}},
			},
		},
	}

	var buf bytes.Buffer
	renderLevelToWriter(&buf, level, "ascii", true)
	out := buf.String()
	if !strings.Contains(out, "-") && !strings.Contains(out, "|") && !strings.Contains(out, "+") {
		t.Fatalf("expected ASCII connector ('-', '|', '+') in output:\n%s", out)
	}
	// Ensure there's an ASCII tail marker for tails/ends
	if !strings.Contains(out, "o") {
		t.Fatalf("expected tail marker 'o' in ASCII connectors output:\n%s", out)
	}
}

func TestRenderLevelUnicodeWithConnectors(t *testing.T) {
	level := &Level{
		ID:       99,
		Name:     "Unicode Level",
		GridSize: [2]int{3, 3},
		Vines: []Vine{
			{ID: "vine_0", HeadDirection: "left", OrderedPath: []Point{{X: 2, Y: 0}, {X: 1, Y: 0}}},
		},
	}

	var buf bytes.Buffer
	renderLevelToWriter(&buf, level, "unicode", false)
	out := buf.String()
	if !strings.Contains(out, "─") && !strings.Contains(out, "│") && !strings.Contains(out, "┌") && !strings.Contains(out, "└") && !strings.Contains(out, "┐") && !strings.Contains(out, "┘") && !strings.Contains(out, "●") {
		t.Fatalf("expected box-drawing or tail marker in output:\n%s", out)
	}
	// Ensure tail marker present for left-moving short vine
	if !strings.Contains(out, "●") {
		t.Fatalf("expected tail marker '●' in output:\n%s", out)
	}

	// And ensure the head arrow is still present
	if !strings.Contains(out, "←") {
		t.Fatalf("expected head arrow '←' in output:\n%s", out)
	}
}
