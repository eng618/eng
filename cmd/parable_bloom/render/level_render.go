package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/cmd/parable_bloom/common"
	"github.com/eng618/eng/internal/utils"
	"github.com/eng618/eng/internal/utils/log"
)

// LevelRenderCmd prints a visual representation of a level to the terminal.
var LevelRenderCmd = &cobra.Command{
	Use:   "level-render",
	Short: "Render a level to the terminal (ASCII/Unicode)",
	Long: `Render a level to the terminal for quick visual inspection.

You can supply a file path with --file (-f) or a level id with --id (-i) (looks in assets/levels).
`,
	Run: func(cmd *cobra.Command, args []string) {
		isVerbose := utils.IsVerbose(cmd)
		log.Start("Rendering level")

		file, _ := cmd.Flags().GetString("file")
		id, _ := cmd.Flags().GetInt("id")
		style, _ := cmd.Flags().GetString("style")
		showCoords, _ := cmd.Flags().GetBool("coords")

		var level *common.Level
		var err error
		if file != "" {
			level, err = common.ReadLevel(file)
			if err != nil {
				log.Error("Failed to read level file: %v", err)
				return
			}
		} else if id != 0 {
			path := common.GetLevelFilePath(id, "assets/levels")
			if !common.FileExists(path) {
				log.Error("Level file not found: %s", path)
				return
			}
			level, err = common.ReadLevel(path)
			if err != nil {
				log.Error("Failed to read level file: %v", err)
				return
			}
		} else {
			log.Error("Please provide either --file or --id to render a level")
			return
		}

		if style == "" {
			style = "unicode"
		}

		RenderLevelToWriter(cmd.OutOrStdout(), level, style, showCoords)
		log.Verbose(isVerbose, "Render complete")
	},
}

func init() {
	LevelRenderCmd.Flags().StringP("file", "f", "", "Path to a level JSON file to render")
	LevelRenderCmd.Flags().IntP("id", "i", 0, "Level ID to render (uses assets/levels/level_<id>.json)")
	LevelRenderCmd.Flags().StringP("style", "s", "unicode", "Render style: ascii or unicode")
	LevelRenderCmd.Flags().BoolP("coords", "c", false, "Show axis coordinates")
}

// RenderLevelToWriter prints a visual representation of a level to the given writer.
// style can be "ascii" or "unicode".
func RenderLevelToWriter(w io.Writer, level *common.Level, style string, showCoords bool) {
	width := level.GetGridWidth()
	height := level.GetGridHeight()

	if width <= 0 || height <= 0 {
		fmt.Fprintf(w, "invalid grid size: %dx%d\n", width, height)
		return
	}

	// Default cell filler
	var emptyCell string
	var headMap map[string]string
	if strings.ToLower(style) == "ascii" {
		emptyCell = "."
		headMap = map[string]string{"up": "^", "down": "v", "left": "<", "right": ">"}
	} else {
		emptyCell = "·"
		headMap = map[string]string{"up": "↑", "down": "↓", "left": "←", "right": "→"}
	}

	// Build grid of strings
	grid := make([][]string, height)
	for y := 0; y < height; y++ {
		row := make([]string, width)
		for x := 0; x < width; x++ {
			row[x] = emptyCell
		}
		grid[y] = row
	}

	// Build occupancy map and compute glyphs via small helpers to lower complexity
	occ := buildOccupancy(level, width, height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			grid[y][x] = computeCellGlyph(level, occ, x, y, style, emptyCell, headMap)
		}
	}
	// Print header
	fmt.Fprintf(w, "Level %d: %s (grid %dx%d)\n", level.ID, level.Name, width, height)

	// Print rows from top (height-1) to 0 — origin lower-left
	for y := height - 1; y >= 0; y-- {
		if showCoords {
			fmt.Fprintf(w, "%2d ", y)
		} else {
			fmt.Fprint(w, "   ")
		}
		for x := 0; x < width; x++ {
			cell := grid[y][x]
			fmt.Fprintf(w, "%2s ", cell)
		}
		fmt.Fprint(w, "\n")
	}

	// Optionally print X coords at the bottom
	if showCoords {
		fmt.Fprint(w, "   ")
		for x := 0; x < width; x++ {
			fmt.Fprintf(w, "%2d ", x%100)
		}
		fmt.Fprint(w, "\n")
	}

	// Legend
	fmt.Fprintln(w, "\nLegend: each non-empty symbol represents a vine; head shown as arrow; '*' indicates collision of vines.")
}

// buildOccupancy creates a map of cell -> list of segment entries.
func buildOccupancy(level *common.Level, width, height int) map[string][]struct{ vineIdx, segIdx int } {
	occ := make(map[string][]struct{ vineIdx, segIdx int })
	for i, vine := range level.Vines {
		for j, pt := range vine.OrderedPath {
			if pt.X < 0 || pt.X >= width || pt.Y < 0 || pt.Y >= height {
				continue
			}
			key := fmt.Sprintf("%d,%d", pt.X, pt.Y)
			occ[key] = append(occ[key], struct{ vineIdx, segIdx int }{vineIdx: i, segIdx: j})
		}
	}
	return occ
}

// computeCellGlyph returns the glyph to draw at the given cell.
func computeCellGlyph(level *common.Level, occ map[string][]struct{ vineIdx, segIdx int }, x, y int, style, emptyCell string, headMap map[string]string) string {
	key := fmt.Sprintf("%d,%d", x, y)
	entries := occ[key]
	if len(entries) == 0 {
		return emptyCell
	}
	if len(entries) > 1 {
		return "*"
	}
	entry := entries[0]
	vine := level.Vines[entry.vineIdx]
	j := entry.segIdx
	// head
	if glyph, ok := headGlyph(vine, j, headMap); ok {
		return glyph
	}
	curr := vine.OrderedPath[j]
	var prev, next *common.Point
	if j > 0 {
		p := vine.OrderedPath[j-1]
		prev = &p
	}
	if j < len(vine.OrderedPath)-1 {
		n := vine.OrderedPath[j+1]
		next = &n
	}
	// neighbor directions
	h := (prev != nil && prev.X == x && prev.Y == y+1) || (next != nil && next.X == x && next.Y == y+1)
	r := (prev != nil && prev.X == x+1 && prev.Y == y) || (next != nil && next.X == x+1 && next.Y == y)
	d := (prev != nil && prev.X == x && prev.Y == y-1) || (next != nil && next.X == x && next.Y == y-1)
	l := (prev != nil && prev.X == x-1 && prev.Y == y) || (next != nil && next.X == x-1 && next.Y == y)

	if glyph, ok := straightGlyph(style, &curr, prev, next); ok {
		return glyph
	}
	if glyph, ok := tailGlyph(style, prev, next); ok {
		return glyph
	}
	return connectorGlyph(style, h, r, d, l)
}

func headGlyph(vine common.Vine, j int, headMap map[string]string) (string, bool) {
	if j == 0 {
		arrow, ok := headMap[vine.HeadDirection]
		if ok {
			return arrow, true
		}
	}
	return "", false
}

func straightGlyph(style string, curr, prev, next *common.Point) (string, bool) {
	if prev != nil && next != nil && prev.Y == next.Y && curr.Y == prev.Y {
		if strings.ToLower(style) == "ascii" {
			return "-", true
		}
		return "─", true
	}
	return "", false
}

func tailGlyph(style string, prev, next *common.Point) (string, bool) {
	if next == nil && prev != nil {
		if strings.ToLower(style) == "ascii" {
			return "o", true
		}
		// Use a smaller square for tail to better match segment proportions
		return "▪", true
	}
	return "", false
}

func connectorGlyph(style string, h, r, d, l bool) string {
	if strings.ToLower(style) == "ascii" {
		if (h || d) && (l || r) {
			return "+"
		}
		if h || d {
			return "|"
		}
		if l || r {
			return "-"
		}
		return "o"
	}
	// Unicode mapping via bitmask to reduce branching.
	bits := 0
	if h {
		bits |= 1
	}
	if r {
		bits |= 2
	}
	if d {
		bits |= 4
	}
	if l {
		bits |= 8
	}
	m := map[int]string{
		15: "┼",
		11: "┴",
		14: "┬",
		13: "┤",
		7:  "├",
		3:  "└",
		6:  "┌",
		12: "┐",
		9:  "┘",
		5:  "│",
		10: "─",
		1:  "│",
		4:  "│",
		2:  "─",
		8:  "─",
	}
	if v, ok := m[bits]; ok {
		return v
	}
	return "·"
}
