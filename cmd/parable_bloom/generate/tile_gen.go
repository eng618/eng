package generate

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/eng618/eng/cmd/parable_bloom/common"
)

// calculateVineLengths computes the initial vine count and their lengths based on constraints and profile.
func calculateVineLengths(
	gridSize [2]int,
	constraints common.DifficultySpec,
	profile common.VarietyProfile,
	rng *rand.Rand,
) (int, []int) {
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

	// Distribute lengths to exactly fill the grid, but consider profile.LengthMix
	lengths := make([]int, vineCount)
	for i := 0; i < vineCount; i++ {
		bucket := chooseLengthBucket(profile, rng)
		switch bucket {
		case "short":
			lengths[i] = maxInt(1, avgLen-2)
		case "medium":
			lengths[i] = maxInt(1, avgLen)
		case "long":
			lengths[i] = maxInt(1, avgLen+2)
		default:
			lengths[i] = avgLen
		}
	}

	// Adjust to exactly fill total cells
	cur := 0
	for _, l := range lengths {
		cur += l
	}
	if cur != total {
		delta := total - cur
		for i := 0; i < int(math.Abs(float64(delta))); i++ {
			idx := i % vineCount
			if delta > 0 {
				lengths[idx]++
			} else if lengths[idx] > 1 {
				lengths[idx]--
			}
		}
	}

	return vineCount, lengths
}

// growVines attempts to grow vines based on the given lengths, returning the vines and occupied map.
func growVines(
	gridSize [2]int,
	lengths []int,
	profile common.VarietyProfile,
	cfg common.GeneratorConfig,
	rng *rand.Rand,
) ([]common.Vine, map[string]bool, error) {
	w := gridSize[0]
	h := gridSize[1]

	occupied := make(map[string]bool)
	vines := make([]common.Vine, 0, len(lengths))

	for i := 0; i < len(lengths); i++ {
		target := lengths[i]
		// Try to grow a vine with several seed attempts
		var grown common.Vine
		var err error
		for attempt := 0; attempt < cfg.MaxSeedRetries; attempt++ {
			seed := pickSeedWithRegionBias(w, h, occupied, profile, rng)
			if seed == nil {
				break
			}
			v, newOcc, e := GrowFromSeed(*seed, occupied, gridSize, target, profile, cfg, rng)
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
			s := pickSeedWithRegionBias(w, h, occupied, profile, rng)
			if s == nil {
				return nil, nil, fmt.Errorf("unable to find empty cell for fallback: %w", err)
			}
			id := fmt.Sprintf("v%d", len(vines)+1)
			v := common.Vine{ID: id, HeadDirection: "up", OrderedPath: []common.Point{*s}}
			vines = append(vines, v)
			occupied[fmt.Sprintf("%d,%d", s.X, s.Y)] = true
		} else {
			grown.ID = fmt.Sprintf("v%d", len(vines)+1)
			vines = append(vines, grown)
		}
	}

	return vines, occupied, nil
}

// fillEmptyCells adds single-cell vines for any remaining empty cells.
func fillEmptyCells(gridSize [2]int, vines []common.Vine, occupied map[string]bool) []common.Vine {
	w := gridSize[0]
	h := gridSize[1]

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			key := fmt.Sprintf("%d,%d", x, y)
			if !occupied[key] {
				id := fmt.Sprintf("v%d", len(vines)+1)
				v := common.Vine{ID: id, HeadDirection: "up", OrderedPath: []common.Point{{X: x, Y: y}}}
				vines = append(vines, v)
				occupied[key] = true
			}
		}
	}

	return vines
}

// TileGridIntoVines partitions the grid into vines according to the provided
// difficulty constraints and a variety profile. It returns a full-coverage set
// of vines (no overlaps, each cell assigned exactly once) or an error.
func TileGridIntoVines(
	gridSize [2]int,
	constraints common.DifficultySpec,
	profile common.VarietyProfile,
	cfg common.GeneratorConfig,
	rng *rand.Rand,
) ([]common.Vine, error) {
	_, lengths := calculateVineLengths(gridSize, constraints, profile, rng)

	vines, occupied, err := growVines(gridSize, lengths, profile, cfg, rng)
	if err != nil {
		return nil, err
	}

	vines = fillEmptyCells(gridSize, vines, occupied)

	// Sanity: ensure coverage and no overlaps
	level := &common.Level{GridSize: gridSize, Vines: vines}
	if err := common.FastValidateLevelCoverage(level); err != nil {
		return nil, fmt.Errorf("tiling final validation failed: %w", err)
	}

	return vines, nil
}

// GrowFromSeed attempts to grow a vine starting from seed, avoiding occupied cells.
// It returns the vine and the updated occupancy map on success.
func GrowFromSeed(
	seed common.Point,
	occupied map[string]bool,
	gridSize [2]int,
	targetLen int,
	profile common.VarietyProfile,
	_ common.GeneratorConfig,
	rng *rand.Rand,
) (common.Vine, map[string]bool, error) {
	w := gridSize[0]
	h := gridSize[1]

	path := []common.Point{seed}
	seen := map[string]bool{fmt.Sprintf("%d,%d", seed.X, seed.Y): true}
	occ := make(map[string]bool)
	for k, v := range occupied {
		occ[k] = v
	}
	occ[fmt.Sprintf("%d,%d", seed.X, seed.Y)] = true

	// choose first step biased by DirBalance if possible
	for len(path) < targetLen {
		head := path[len(path)-1]
		neighbors := availableNeighbors(head, w, h, occ)
		if len(neighbors) == 0 {
			// Stuck
			return common.Vine{}, nil, fmt.Errorf("stuck at length %d, target %d", len(path), targetLen)
		}

		var chosen common.Point
		if len(path) == 1 && len(profile.DirBalance) > 0 {
			// try to bias initial direction
			chosen = chooseNeighborByDirBias(head, neighbors, profile.DirBalance, rng)
			// fall back if chosen is zero
			if chosen == (common.Point{}) {
				chosen = neighbors[rng.Intn(len(neighbors))]
			}
		} else if len(path) >= 2 {
			// prefer continuing straight vs turning based on TurnMix
			prev := path[len(path)-2]
			dx := head.X - prev.X
			dy := head.Y - prev.Y
			straight := common.Point{X: head.X + dx, Y: head.Y + dy}
			// find if straight is available
			straightAvailable := false
			for _, n := range neighbors {
				if n.X == straight.X && n.Y == straight.Y {
					straightAvailable = true
					break
				}
			}
			if straightAvailable && rng.Float64() > profile.TurnMix {
				chosen = straight
			} else {
				chosen = neighbors[rng.Intn(len(neighbors))]
			}
		} else {
			// default random pick
			chosen = neighbors[rng.Intn(len(neighbors))]
		}

		k := fmt.Sprintf("%d,%d", chosen.X, chosen.Y)
		path = append(path, chosen)
		seen[k] = true
		occ[k] = true
	}

	// Heuristic head direction: head is path[0] -> path[1]
	headDir := "up"
	if len(path) >= 2 {
		dx := path[1].X - path[0].X
		dy := path[1].Y - path[0].Y
		// HeadDirection should be the moving direction; neck-head is opposite of movement
		for dir, d := range common.HeadDirections {
			if d[0] == -dx && d[1] == -dy {
				headDir = dir
				break
			}
		}
	}

	return common.Vine{HeadDirection: headDir, OrderedPath: path}, occ, nil
}

// availableNeighbors lists unoccupied Manhattan neighbors within grid.
func availableNeighbors(p common.Point, w, h int, occ map[string]bool) []common.Point {
	candidates := []common.Point{
		{X: p.X + 1, Y: p.Y},
		{X: p.X - 1, Y: p.Y},
		{X: p.X, Y: p.Y + 1},
		{X: p.X, Y: p.Y - 1},
	}
	out := make([]common.Point, 0, 4)
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

// chooseNeighborByDirBias picks a neighbor closest to a desired direction distribution.
func chooseNeighborByDirBias(
	origin common.Point,
	neighbors []common.Point,
	dirBalance map[string]float64,
	rng *rand.Rand,
) common.Point {
	if len(neighbors) == 0 || len(dirBalance) == 0 {
		return common.Point{}
	}
	// Score neighbors by their direction
	scores := make([]float64, len(neighbors))
	sum := 0.0
	for i, n := range neighbors {
		dx := n.X - origin.X
		dy := n.Y - origin.Y
		var dir string
		for k, v := range common.HeadDirections {
			if v[0] == dx && v[1] == dy {
				dir = k
				break
			}
		}
		s := dirBalance[dir]
		scores[i] = s
		sum += s
	}
	if sum == 0 {
		// fallback random
		return neighbors[rng.Intn(len(neighbors))]
	}
	// pick weighted
	r := rng.Float64() * sum
	acc := 0.0
	for i, s := range scores {
		acc += s
		if r <= acc {
			return neighbors[i]
		}
	}
	return neighbors[len(neighbors)-1]
}

// pickSeedWithRegionBias picks a seed based on profile.RegionBias and emptiness.
func pickSeedWithRegionBias(
	w, h int,
	occ map[string]bool,
	profile common.VarietyProfile,
	rng *rand.Rand,
) *common.Point {
	// if no profile fields set, fall back to uniform
	if profile.LengthMix == nil && profile.DirBalance == nil && profile.RegionBias == "" {
		return randomEmptyCell(w, h, occ, rng)
	}
	empty := make([]common.Point, 0)
	weights := make([]float64, 0)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			k := fmt.Sprintf("%d,%d", x, y)
			if occ[k] {
				continue
			}
			p := common.Point{X: x, Y: y}
			empty = append(empty, p)
			// base weight
			var wgt float64
			switch profile.RegionBias {
			case "edge":
				// favor distance to nearest edge
				d := minInt(minInt(x, w-1-x), minInt(y, h-1-y))
				wgt = float64(1 + (w/2 - d))
			case "center":
				cx := float64(w-1) / 2.0
				cy := float64(h-1) / 2.0
				dx := float64(x) - cx
				dy := float64(y) - cy
				dist := dx*dx + dy*dy
				wgt = 1.0 / (0.1 + dist)
			default:
				wgt = 1.0
			}
			weights = append(weights, wgt)
		}
	}
	if len(empty) == 0 {
		return nil
	}
	// weighted pick
	total := 0.0
	for _, v := range weights {
		total += v
	}
	r := rng.Float64() * total
	acc := 0.0
	for i, v := range weights {
		acc += v
		if r <= acc {
			return &empty[i]
		}
	}
	return &empty[len(empty)-1]
}

// chooseLengthBucket picks short/medium/long based on LengthMix weights.
func chooseLengthBucket(profile common.VarietyProfile, rng *rand.Rand) string {
	if len(profile.LengthMix) == 0 {
		return "medium"
	}
	wShort := profile.LengthMix["short"]
	wMed := profile.LengthMix["medium"]
	wLong := profile.LengthMix["long"]
	total := wShort + wMed + wLong
	if total <= 0 {
		return "medium"
	}
	r := rng.Float64() * total
	if r <= wShort {
		return "short"
	}
	r -= wShort
	if r <= wMed {
		return "medium"
	}
	return "long"
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// randomEmptyCell picks a random empty cell from the grid; returns nil if none.
func randomEmptyCell(w, h int, occ map[string]bool, rng *rand.Rand) *common.Point {
	empty := make([]common.Point, 0)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			key := fmt.Sprintf("%d,%d", x, y)
			if !occ[key] {
				empty = append(empty, common.Point{X: x, Y: y})
			}
		}
	}
	if len(empty) == 0 {
		return nil
	}
	p := empty[rng.Intn(len(empty))]
	return &p
}
