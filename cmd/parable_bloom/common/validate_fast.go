package common

import (
	"fmt"
)

// FastValidateLevelCoverage performs fast structural checks on a level.
// It verifies:
//   - grid size is positive
//   - each vine path is contiguous (Manhattan adjacency)
//   - no self-intersections within a vine
//   - head direction matches the first segment (head -> neck)
//   - all coordinates are in-bounds
//   - every grid cell is assigned to exactly one vine or masked (full coverage)
func FastValidateLevelCoverage(level *Level) error {
	w := level.GetGridWidth()
	h := level.GetGridHeight()
	if w <= 0 || h <= 0 {
		return fmt.Errorf("invalid grid size: %dx%d", w, h)
	}

	total := w * h
	occupied := make(map[string]int)

	for _, vine := range level.Vines {
		if len(vine.OrderedPath) == 0 {
			return fmt.Errorf("vine %q has empty path", vine.ID)
		}

		seen := make(map[string]bool)
		for i, pt := range vine.OrderedPath {
			// bounds check
			if pt.X < 0 || pt.X >= w || pt.Y < 0 || pt.Y >= h {
				return fmt.Errorf("vine %q has out-of-bounds point %v", vine.ID, pt)
			}

			key := fmt.Sprintf("%d,%d", pt.X, pt.Y)
			occupied[key]++
			if seen[key] {
				return fmt.Errorf("vine %q self-intersects at %v", vine.ID, pt)
			}
			seen[key] = true

			// contiguity check with previous
			if i > 0 {
				prev := vine.OrderedPath[i-1]
				dx := prev.X - pt.X
				dy := prev.Y - pt.Y
				if abs(dx)+abs(dy) != 1 {
					return fmt.Errorf("vine %q has non-contiguous segment between %v and %v", vine.ID, prev, pt)
				}
			}
		}

		// head direction check
		if len(vine.OrderedPath) >= 2 {
			head := vine.GetHead()
			neck := vine.GetNeck()
			// neck - head should be the opposite of HeadDirections[headdir]
			exp, ok := HeadDirections[vine.HeadDirection]
			if !ok {
				return fmt.Errorf("vine %q has invalid head_direction %q", vine.ID, vine.HeadDirection)
			}
			dx := neck.X - head.X
			dy := neck.Y - head.Y
			if dx != -exp[0] || dy != -exp[1] {
				return fmt.Errorf(
					"vine %q head_direction %q does not match first segment (head=%v neck=%v)",
					vine.ID,
					vine.HeadDirection,
					head,
					neck,
				)
			}
		}
	}

	// Count masked cells
	maskedCount := 0
	if level.Mask != nil && level.Mask.Mode == "hide" {
		for _, p := range level.Mask.Points {
			if _, ok := p.(Point); ok {
				maskedCount++
			} else if _, ok := p.([]interface{}); ok && len(p.([]interface{})) == 2 {
				// assume [x,y]
				maskedCount++
			}
		}
	}

	expected := total - maskedCount
	if len(occupied) != expected {
		return fmt.Errorf("grid coverage mismatch: occupied %d cells, expected %d (total %d - masked %d)", len(occupied), expected, total, maskedCount)
	}

	// ensure no overlaps (counts > 1)
	for k, c := range occupied {
		if c != 1 {
			return fmt.Errorf("cell %s covered %d times", k, c)
		}
	}

	return nil
}
