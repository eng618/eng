//go:build ignore
// +build ignore

package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/eng618/eng/cmd/parable_bloom/common"
)

func TestParableBloomTestsRemoved(t *testing.T) {
	t.Skip("Parable Bloom eng CLI tests removed; use tools/level-builder tests instead.")
}

func TestRenderLevelAscii(t *testing.T) {
	level := &common.Level{
		ID:       42,
		Name:     "Test Level",
		GridSize: [2]int{5, 4},
		Vines: []common.Vine{
			{
				ID:            "vine_0",
				HeadDirection: "right",
				OrderedPath:   []common.Point{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}},
			},
			{
				ID:            "vine_1",
				HeadDirection: "up",
				OrderedPath:   []common.Point{{X: 4, Y: 3}, {X: 4, Y: 2}},
			},
		},
	}

	var buf bytes.Buffer
	RenderLevelToWriter(&buf, level, "ascii", true)
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
	level := &common.Level{
		ID:       99,
		Name:     "Unicode Level",
		GridSize: [2]int{3, 3},
		Vines: []common.Vine{
			{ID: "vine_0", HeadDirection: "left", OrderedPath: []common.Point{{X: 2, Y: 0}, {X: 1, Y: 0}}},
		},
	}

	var buf bytes.Buffer
	RenderLevelToWriter(&buf, level, "unicode", false)
	out := buf.String()
	if !strings.Contains(out, "←") {
		t.Fatalf("expected unicode left arrow in output:\n%s", out)
	}
	// Default (connectors=true) should show box-drawing or tail marker
	// For tail marker, compute expected glyph via tailGlyph for the tail segment and assert it's present
	expectedTail, _ := tailGlyph("unicode", &common.Point{X: 2, Y: 0}, nil)
	found := strings.Contains(out, expectedTail)
	if !found {
		for _, ch := range []string{"─", "│", "┌", "└", "┐", "┘"} {
			if strings.Contains(out, ch) {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatalf("expected box-drawing or tail marker in output:\n%s", out)
	}
}

func TestRenderLevelAsciiWithConnectors(t *testing.T) {
	level := &common.Level{
		ID:       42,
		Name:     "Test Level",
		GridSize: [2]int{5, 4},
		Vines: []common.Vine{
			{
				ID:            "vine_0",
				HeadDirection: "right",
				OrderedPath:   []common.Point{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}},
			},
			{
				ID:            "vine_1",
				HeadDirection: "up",
				OrderedPath:   []common.Point{{X: 4, Y: 3}, {X: 4, Y: 2}},
			},
		},
	}

	var buf bytes.Buffer
	RenderLevelToWriter(&buf, level, "ascii", true)
	out := buf.String()
	if !strings.Contains(out, "-") && !strings.Contains(out, "|") && !strings.Contains(out, "+") {
		t.Fatalf("expected ASCII connector ('-', '|', '+') in output:\n%s", out)
	}
	// Ensure there's an ASCII tail marker for tails/ends
	if !strings.Contains(out, "o") {
		t.Fatalf("expected tail marker 'o' in ASCII connectors output:\n%s", out)
	}
}

func TestRenderLevelUnicode_DebugTail(t *testing.T) {
	level := &common.Level{
		ID:       99,
		Name:     "Unicode Level",
		GridSize: [2]int{3, 3},
		Vines: []common.Vine{
			{ID: "vine_0", HeadDirection: "left", OrderedPath: []common.Point{{X: 2, Y: 0}, {X: 1, Y: 0}}},
		},
	}
	var buf bytes.Buffer
	RenderLevelToWriter(&buf, level, "unicode", false)
	out := buf.String()
	expectedTail, _ := tailGlyph("unicode", &common.Point{X: 2, Y: 0}, nil)
	if expectedTail == "" {
		t.Fatalf("expectedTail was empty; unexpected")
	}
	if !strings.Contains(out, expectedTail) {
		t.Fatalf("expected tail %q not found in output:\n%s", expectedTail, out)
	}
}

func TestRenderLevelUnicodeConnectorsPresence(t *testing.T) {
	level := &common.Level{
		ID:       99,
		Name:     "Unicode Level",
		GridSize: [2]int{3, 3},
		Vines: []common.Vine{
			{ID: "vine_0", HeadDirection: "left", OrderedPath: []common.Point{{X: 2, Y: 0}, {X: 1, Y: 0}}},
		},
	}

	var buf bytes.Buffer
	RenderLevelToWriter(&buf, level, "unicode", false)
	out := buf.String()
	if !containsAny(out, []string{"─", "│", "┌", "└", "┐", "┘", "▪"}) {
		t.Fatalf("expected box-drawing or tail marker in output:\n%s", out)
	}
	// Ensure tail marker present for left-moving short vine
	expectedTail2, _ := tailGlyph("unicode", &common.Point{X: 2, Y: 0}, nil)
	if !strings.Contains(out, expectedTail2) {
		t.Fatalf("expected tail marker '%s' in output:\n%s", expectedTail2, out)
	}
}

func TestRenderLevelUnicodeHeadPresent(t *testing.T) {
	level := &common.Level{
		ID:       99,
		Name:     "Unicode Level",
		GridSize: [2]int{3, 3},
		Vines: []common.Vine{
			{ID: "vine_0", HeadDirection: "left", OrderedPath: []common.Point{{X: 2, Y: 0}, {X: 1, Y: 0}}},
		},
	}

	var buf bytes.Buffer
	RenderLevelToWriter(&buf, level, "unicode", false)
	out := buf.String()
	if !strings.Contains(out, "←") {
		t.Fatalf("expected head arrow '←' in output:\n%s", out)
	}
}
