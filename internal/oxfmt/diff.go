package oxfmt

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sergi/go-diff/diffmatchpatch"

	"github.com/eng618/eng/internal/ui/theme"
)

// diffContextLines is the number of unchanged lines shown around each hunk.
const diffContextLines = 3

var (
	diffAddStyle    = lipgloss.NewStyle().Foreground(theme.Success)         //nolint:gochecknoglobals
	diffDelStyle    = lipgloss.NewStyle().Foreground(theme.Destructive)     //nolint:gochecknoglobals
	diffHunkStyle   = lipgloss.NewStyle().Foreground(theme.MutedForeground) //nolint:gochecknoglobals
	diffHeaderStyle = lipgloss.NewStyle().Bold(true)                        //nolint:gochecknoglobals
)

// diffOp is a single rendered line: ' ' context, '-' removal, '+' addition.
type diffOp struct {
	kind rune
	text string
}

// RenderDiff returns a colored unified-style diff between oldText and newText
// for path. It returns an empty string when the inputs are identical.
func RenderDiff(path, oldText, newText string) string {
	oldText = strings.ReplaceAll(oldText, "\r\n", "\n")
	newText = strings.ReplaceAll(newText, "\r\n", "\n")
	if oldText == newText {
		return ""
	}
	ops := lineOps(oldText, newText)
	ranges := hunkRanges(ops, diffContextLines)
	if len(ranges) == 0 {
		ranges = [][2]int{{0, len(ops)}}
	}
	display := strings.TrimPrefix(strings.TrimPrefix(path, "./"), "/")
	var sb strings.Builder
	sb.WriteString(diffHeaderStyle.Render("--- a/"+display) + "\n")
	sb.WriteString(diffHeaderStyle.Render("+++ b/"+display) + "\n")
	oldLine, newLine := 1, 1
	pos := 0
	for _, r := range ranges {
		for _, op := range ops[pos:r[0]] {
			oldLine, newLine = advance(op.kind, oldLine, newLine)
		}
		oldCount, newCount := counts(ops[r[0]:r[1]])
		fmt.Fprintf(&sb, "%s\n", diffHunkStyle.Render(
			fmt.Sprintf("@@ -%d,%d +%d,%d @@", oldLine, oldCount, newLine, newCount),
		))
		for _, op := range ops[r[0]:r[1]] {
			sb.WriteString(renderOp(op) + "\n")
			oldLine, newLine = advance(op.kind, oldLine, newLine)
		}
		pos = r[1]
	}
	return strings.TrimSuffix(sb.String(), "\n")
}

// lineOps diffs oldText and newText at line granularity.
func lineOps(oldText, newText string) []diffOp {
	dmp := diffmatchpatch.New()
	a, b, lines := dmp.DiffLinesToChars(oldText, newText)
	diffs := dmp.DiffMain(a, b, false)
	diffs = dmp.DiffCharsToLines(diffs, lines)
	var ops []diffOp
	for _, d := range diffs {
		kind := ' '
		switch d.Type {
		case diffmatchpatch.DiffDelete:
			kind = '-'
		case diffmatchpatch.DiffInsert:
			kind = '+'
		}
		chunks := strings.SplitAfter(d.Text, "\n")
		for i, chunk := range chunks {
			if i == len(chunks)-1 && chunk == "" {
				continue
			}
			ops = append(ops, diffOp{kind: kind, text: strings.TrimSuffix(chunk, "\n")})
		}
	}
	return ops
}

// hunkRanges merges changed-line windows (with context radius) into [from, to)
// ranges over ops.
func hunkRanges(ops []diffOp, contextLines int) [][2]int {
	var ranges [][2]int
	for i, op := range ops {
		if op.kind == ' ' {
			continue
		}
		from := max(0, i-contextLines)
		to := min(len(ops), i+contextLines+1)
		if n := len(ranges); n > 0 && from <= ranges[n-1][1] {
			ranges[n-1][1] = max(ranges[n-1][1], to)
			continue
		}
		ranges = append(ranges, [2]int{from, to})
	}
	return ranges
}

// counts returns the old/new line spans of a hunk.
func counts(ops []diffOp) (int, int) {
	oldCount, newCount := 0, 0
	for _, op := range ops {
		switch op.kind {
		case '-':
			oldCount++
		case '+':
			newCount++
		default:
			oldCount++
			newCount++
		}
	}
	return oldCount, newCount
}

// advance moves line counters past one op.
func advance(kind rune, oldLine, newLine int) (int, int) {
	switch kind {
	case '-':
		return oldLine + 1, newLine
	case '+':
		return oldLine, newLine + 1
	default:
		return oldLine + 1, newLine + 1
	}
}

// renderOp styles one diff line.
func renderOp(op diffOp) string {
	switch op.kind {
	case '-':
		return diffDelStyle.Render("-" + op.text)
	case '+':
		return diffAddStyle.Render("+" + op.text)
	default:
		return " " + op.text
	}
}
