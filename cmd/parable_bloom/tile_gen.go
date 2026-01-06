package parable_bloom

import (
	"fmt"
	"math/rand"
)

// TileGridIntoVines partitions the grid into vines according to the provided
// difficulty constraints and a variety profile. It returns a full-coverage set
// of vines (no overlaps, each cell assigned exactly once) or an error.
func TileGridIntoVines(gridSize [2]int, constraints DifficultySpec, profile VarietyProfile, cfg GeneratorConfig, rng *rand.Rand) ([]Vine, error) {
	w := gridSize[0]
	h := gridSize[1]
	total := w * h

	// Choose a target average length (middle of range)
	minLen := constraints.AvgLengthRange[0]
	maxLen := constraints.AvgLengthRange[1]
	avgLen := (minLen + maxLen) / 2
	if avgLen <= 0 {
		avgLen = 3
	}

	// Initial vine count (rounded)
	vineCount := total / avgLen
	if vineCount < constraints.VineCountRange[0] {
		vineCount = constraints.VineCountRange[0]
	}
	if vineCount > constraints.VineCountRange[1] {
		vineCount = constraints.VineCountRange[1]
	}

	// Distribute lengths to exactly fill the grid
	lengths := make([]int, vineCount)
	for i := 0; i < vineCount; i++ {
		lengths[i] = total / vineCount
	}
	remainder := total - (lengths[0] * vineCount)
	for i := 0; i < remainder; i++ {
		lengths[i%vineCount]++
	}

	occupied := make(map[string]bool)
	vines := make([]Vine, 0, vineCount)

	for i := 0; i < vineCount; i++ {
		target := lengths[i]
		// Try to grow a vine with several seed attempts
		var grown Vine
		var err error
		for attempt := 0; attempt < cfg.MaxSeedRetries; attempt++ {
			seed := randomEmptyCell(w, h, occupied, rng)
			if seed == nil {
				break
			}
			v, newOcc, e := GrowFromSeed(*seed, occupied, gridSize, target, rng)
			if e == nil {
				grown = v
				for k := range newOcc {
					occupied[k] = true
				}
				break
			}
			err = e
		}

		if grown.Length() == 0 {
			// fallback: create small single-cell vine at any empty cell
			s := randomEmptyCell(w, h, occupied, rng)
			if s == nil {
				return nil, fmt.Errorf("unable to find empty cell for fallback: %w", err)
			}
			id := fmt.Sprintf("v%d", len(vines)+1)
			v := Vine{ID: id, HeadDirection: "up", OrderedPath: []Point{*s}}
			vines = append(vines, v)
			occupied[fmt.Sprintf("%d,%d", s.X, s.Y)] = true
		} else {
			grown.ID = fmt.Sprintf("v%d", len(vines)+1)
			vines = append(vines, grown)
		}
	}

	// Final sweep: any remaining empty cells become single-segment vines
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			key := fmt.Sprintf("%d,%d", x, y)
			if !occupied[key] {
				id := fmt.Sprintf("v%d", len(vines)+1)
				v := Vine{ID: id, HeadDirection: "up", OrderedPath: []Point{{x, y}}}
				vines = append(vines, v)
				occupied[key] = true
			}
		}
	}

	// Sanity: ensure coverage and no overlaps
	level := &Level{GridSize: gridSize, Vines: vines}
	if err := FastValidateLevelCoverage(level); err != nil {
		return nil, fmt.Errorf("tiling final validation failed: %w", err)
	}

	return vines, nil
}

// GrowFromSeed attempts to grow a vine starting from seed, avoiding occupied cells.
// It returns the vine and the updated occupancy map on success.
func GrowFromSeed(seed Point, occupied map[string]bool, gridSize [2]int, targetLen int, rng *rand.Rand) (Vine, map[string]bool, error) {
	w := gridSize[0]
	h := gridSize[1]

	path := []Point{seed}
	seen := map[string]bool{fmt.Sprintf("%d,%d", seed.X, seed.Y): true}
	occ := make(map[string]bool)
	for k, v := range occupied {
		occ[k] = v
	}
	occ[fmt.Sprintf("%d,%d", seed.X, seed.Y)] = true

	for len(path) < targetLen {
		head := path[len(path)-1]
		neighbors := availableNeighbors(head, w, h, occ)
		if len(neighbors) == 0 {
			// Stuck
			return Vine{}, nil, fmt.Errorf("stuck at length %d, target %d", len(path), targetLen)
		}
		// pick a neighbor randomly
		n := neighbors[rng.Intn(len(neighbors))]
		k := fmt.Sprintf("%d,%d", n.X, n.Y)
		path = append(path, n)
		seen[k] = true
		occ[k] = true
	}

	// Heuristic head direction: head is path[0] -> path[1]
	headDir := "up"
	if len(path) >= 2 {
		dx := path[1].X - path[0].X
		dy := path[1].Y - path[0].Y
		// HeadDirection should be the moving direction; neck-head is opposite of movement
		for dir, d := range HeadDirections {
			if d[0] == -dx && d[1] == -dy {
				headDir = dir
				break
			}
		}
	}

	return Vine{HeadDirection: headDir, OrderedPath: path}, occ, nil
}

// availableNeighbors lists unoccupied Manhattan neighbors within grid.
func availableNeighbors(p Point, w, h int, occ map[string]bool) []Point {
	candidates := []Point{{p.X + 1, p.Y}, {p.X - 1, p.Y}, {p.X, p.Y + 1}, {p.X, p.Y - 1}}
	out := make([]Point, 0, 4)
	for _, c := range candidates {
		if c.X < 0 || c.X >= w || c.Y < 0 || c.Y >= h {
			continue
		}
		if occ[fmt.Sprintf("%d,%d", c.X, c.Y)] {
			continue
		}
		out = append(out, c)
	}
	return out
}

// randomEmptyCell picks a random empty cell from the grid; returns nil if none.
func randomEmptyCell(w, h int, occ map[string]bool, rng *rand.Rand) *Point {
	empty := make([]Point, 0)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			key := fmt.Sprintf("%d,%d", x, y)
			if !occ[key] {
				empty = append(empty, Point{X: x, Y: y})
			}
		}
	}
	if len(empty) == 0 {
		return nil
	}
	p := empty[rng.Intn(len(empty))]
	return &p
}
