package oxfmt

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderDiffIdentical(t *testing.T) {
	assert.Empty(t, RenderDiff("a.ts", "const x = 1;\n", "const x = 1;\n"))
	assert.Empty(t, RenderDiff("a.ts", "", ""))
}

func TestRenderDiffSimpleChange(t *testing.T) {
	got := RenderDiff("a.ts", "const   x=1\n", "const x = 1;\n")
	assert.Contains(t, got, "--- a/a.ts")
	assert.Contains(t, got, "+++ b/a.ts")
	assert.Contains(t, got, "@@")
	assert.Contains(t, got, "-const   x=1")
	assert.Contains(t, got, "+const x = 1;")
}

func TestRenderDiffHeaderStripsAbsolutePrefix(t *testing.T) {
	got := RenderDiff("/tmp/work/a.ts", "a\n", "b\n")
	assert.Contains(t, got, "--- a/tmp/work/a.ts")
	assert.NotContains(t, got, "a//tmp")
}

func TestRenderDiffCollapsesContext(t *testing.T) {
	var oldLines, newLines []string
	for i := range 30 {
		oldLines = append(oldLines, fmt.Sprintf("unchanged-%02d", i))
		newLines = append(newLines, oldLines[i])
	}
	oldLines[2] = "before"
	newLines[2] = "after"
	oldLines[25] = "before"
	newLines[25] = "after"
	got := RenderDiff("big.ts", strings.Join(oldLines, "\n")+"\n", strings.Join(newLines, "\n")+"\n")
	assert.Equal(t, 2, strings.Count(got, "@@ -"), "expected two hunks, got:\n%s", got)
	assert.Contains(t, got, "-before")
	assert.Contains(t, got, "+after")
	assert.Contains(t, got, " "+oldLines[0], "near context should render")
	assert.NotContains(t, got, oldLines[15], "far context should collapse")
}
