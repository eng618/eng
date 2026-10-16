package ui

import (
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

var diffANSI = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// stripDiffANSI removes ANSI escape codes for content assertions.
func stripDiffANSI(s string) string {
	return diffANSI.ReplaceAllString(s, "")
}

func TestRenderDiffLine(t *testing.T) {
	// Force color output so style assertions are deterministic.
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(prev)

	cases := []struct {
		name   string
		line   string
		plain  string
		styled bool
	}{
		{name: "addition styled", line: "+dart 3.13.5", plain: "+dart 3.13.5", styled: true},
		{name: "deletion styled", line: "-dart 3.13.3", plain: "-dart 3.13.3", styled: true},
		{name: "hunk styled", line: "@@ -2,3 +2,3 @@", plain: "@@ -2,3 +2,3 @@", styled: true},
		{name: "file header plus not add", line: "+++ b/.tool-versions", plain: "+++ b/.tool-versions", styled: true},
		{name: "file header minus not del", line: "--- a/.tool-versions", plain: "--- a/.tool-versions", styled: true},
		{
			name:   "git metadata",
			line:   "diff --git a/.tool-versions b/.tool-versions",
			plain:  "diff --git a/.tool-versions b/.tool-versions",
			styled: true,
		},
		{name: "context plain", line: " bun 1.4.2", plain: " bun 1.4.2", styled: false},
		{name: "empty plain", line: "", plain: "", styled: false},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderDiffLine(tt.line)
			if stripped := stripDiffANSI(got); stripped != tt.plain {
				t.Fatalf("stripped = %q, want %q (raw %q)", stripped, tt.plain, got)
			}
			if tt.styled && got == tt.line {
				t.Fatalf("expected styled output to differ from input for %q", tt.line)
			}
			if !tt.styled && got != tt.line {
				t.Fatalf("expected plain passthrough, got %q", got)
			}
		})
	}
}

func TestRenderDiffLineAddDelDistinct(t *testing.T) {
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(prev)

	add := RenderDiffLine("+x")
	del := RenderDiffLine("-x")
	if add == del {
		t.Fatal("expected add and del styles to differ")
	}
	if stripDiffANSI(add) != "+x" || stripDiffANSI(del) != "-x" {
		t.Fatalf("content not preserved: %q %q", add, del)
	}
}

func TestRenderDiffPreservesContent(t *testing.T) {
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(prev)

	in := "diff --git a/f b/f\n--- a/f\n+++ b/f\n@@ -1 +1 @@\n-old\n+new\n ctx\n"
	got := RenderDiff(in)
	if stripDiffANSI(got) != strings.TrimSuffix(in, "\n")+"\n" {
		t.Fatalf("content mismatch:\n%q", got)
	}
	if got == in {
		t.Fatal("expected styled diff to differ from plain input")
	}
	if out := RenderDiff(""); out != "" {
		t.Fatalf("expected empty passthrough, got %q", out)
	}
}

func TestRenderDiffStatColorsMarkers(t *testing.T) {
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(prev)

	in := " .tool-versions | 8 ++++----\n 1 file changed, 4 insertions(+), 4 deletions(-)\n"
	got := RenderDiffStat(in)
	if stripDiffANSI(got) != in {
		t.Fatalf("stat content mismatch:\n%q", got)
	}
	if got == in {
		t.Fatal("expected styled stat to differ from plain input")
	}
	// Lines without markers still pass through with content intact.
	plain := " 1 file changed\n"
	if stripDiffANSI(RenderDiffStat(plain)) != plain {
		t.Fatalf("expected stat passthrough, got %q", RenderDiffStat(plain))
	}
}
