package parable_bloom

import (
	"fmt"
	"time"
)

// Point represents a coordinate in the game grid.
type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// String returns a string representation of the point.
func (p Point) String() string {
	return fmt.Sprintf("(%d,%d)", p.X, p.Y)
}

// Mask represents the visual masking configuration for a level.
type Mask struct {
	Mode   string `json:"mode"`   // "show-all", "hide", "show"
	Points []any  `json:"points"` // [][]int or []Point
}

// Vine represents a single vine in the level.
type Vine struct {
	ID            string   `json:"id"`
	HeadDirection string   `json:"head_direction"` // "up", "down", "left", "right"
	OrderedPath   []Point  `json:"ordered_path"`
	VineColor     string   `json:"vine_color"`
	Blocks        []string `json:"blocks"`
}

// GetHead returns the head segment (first point) of the vine.
func (v *Vine) GetHead() Point {
	if len(v.OrderedPath) > 0 {
		return v.OrderedPath[0]
	}
	return Point{X: -1, Y: -1}
}

// GetNeck returns the neck segment (second point) of the vine.
func (v *Vine) GetNeck() Point {
	if len(v.OrderedPath) > 1 {
		return v.OrderedPath[1]
	}
	return Point{X: -1, Y: -1}
}

// GetTail returns the tail segment (last point) of the vine.
func (v *Vine) GetTail() Point {
	if len(v.OrderedPath) > 0 {
		return v.OrderedPath[len(v.OrderedPath)-1]
	}
	return Point{X: -1, Y: -1}
}

// Length returns the number of segments in the vine.
func (v *Vine) Length() int {
	return len(v.OrderedPath)
}

// Level represents a complete game level.
type Level struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Difficulty string `json:"difficulty"` // "Tutorial", "Seedling", "Sprout", "Nurturing", "Flourishing", "Transcendent"
	GridSize   [2]int `json:"grid_size"`  // [width, height]
	Mask       *Mask  `json:"mask"`
	Vines      []Vine `json:"vines"`
	MaxMoves   int    `json:"max_moves"`
	MinMoves   int    `json:"min_moves"`
	Complexity string `json:"complexity"` // "tutorial", "low", "medium", "high", "extreme"
	Grace      int    `json:"grace"`      // 3 or 4
	// Generation metadata persisted for reproducibility & diagnostics
	GenerationSeed       int64   `json:"generation_seed,omitempty"`
	GenerationAttempts   int     `json:"generation_attempts,omitempty"`
	GenerationElapsedMS  int64   `json:"generation_elapsed_ms,omitempty"`
	GenerationScore      float64 `json:"generation_score,omitempty"`
	// These are populated during validation but not persisted
	OccupancyPercent  float64             `json:"-"`
	ColorDistribution map[string]float64  `json:"-"`
	BlockingGraph     map[string][]string `json:"-"`
}

// GetGridWidth returns the width of the grid.
func (l *Level) GetGridWidth() int {
	if len(l.GridSize) > 0 {
		return l.GridSize[0]
	}
	return 0
}

// GetGridHeight returns the height of the grid.
func (l *Level) GetGridHeight() int {
	if len(l.GridSize) > 1 {
		return l.GridSize[1]
	}
	return 0
}

// GetTotalCells returns the total number of cells in the grid.
func (l *Level) GetTotalCells() int {
	return l.GetGridWidth() * l.GetGridHeight()
}

// GetOccupiedCells returns the total number of cells occupied by vines.
func (l *Level) GetOccupiedCells() int {
	total := 0
	for _, vine := range l.Vines {
		total += vine.Length()
	}
	return total
}

// Module represents a module with parable content and level range.
type Module struct {
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	LevelRange    [2]int   `json:"level_range"`
	Parable       *Parable `json:"parable"`
	UnlockMessage string   `json:"unlock_message"`
}

// Parable represents the parable content for a module.
type Parable struct {
	Title           string `json:"title"`
	Scripture       string `json:"scripture"`
	Content         string `json:"content"`
	Reflection      string `json:"reflection"`
	BackgroundImage string `json:"background_image"`
}

// ModuleData represents the complete modules.json structure.
type ModuleData struct {
	Version           string                 `json:"version"`
	Modules           []Module               `json:"modules"`
	ColorDistribution map[string]interface{} `json:"color_distribution"`
	BlockingGraph     map[string]interface{} `json:"blocking_graph"`
}

// ModuleRange represents a difficulty tier range with module context.
type ModuleRange struct {
	ID    int
	Name  string
	Start int
	End   int
}

// ValidationResult holds the results of level validation.
type ValidationResult struct {
	Filename   string
	Violations []string
	Warnings   []string
	Valid      bool
	Timestamp  time.Time
}

// DifficultySpec defines constraints for a difficulty tier.
type DifficultySpec struct {
	VineCountRange   [2]int
	AvgLengthRange   [2]int
	MaxBlockingDepth int
	ColorCountRange  [2]int
	MinGridOccupancy float64
	DefaultGrace     int
}

var DifficultySpecs = map[string]DifficultySpec{
	"Tutorial": {
		VineCountRange:   [2]int{3, 8},
		AvgLengthRange:   [2]int{6, 8},
		MaxBlockingDepth: 0,
		ColorCountRange:  [2]int{1, 5},
		MinGridOccupancy: 0.30,
		DefaultGrace:     3,
	},
	"Seedling": {
		VineCountRange:   [2]int{4, 60},
		AvgLengthRange:   [2]int{6, 8},
		MaxBlockingDepth: 1,
		ColorCountRange:  [2]int{1, 5},
		MinGridOccupancy: 0.93,
		DefaultGrace:     3,
	},
	"Sprout": {
		VineCountRange:   [2]int{8, 80},
		AvgLengthRange:   [2]int{3, 8},
		MaxBlockingDepth: 2,
		ColorCountRange:  [2]int{1, 5},
		MinGridOccupancy: 0.93,
		DefaultGrace:     3,
	},
	"Nurturing": {
		VineCountRange:   [2]int{12, 100},
		AvgLengthRange:   [2]int{3, 8},
		MaxBlockingDepth: 3,
		ColorCountRange:  [2]int{1, 6},
		MinGridOccupancy: 0.93,
		DefaultGrace:     3,
	},
	"Flourishing": {
		VineCountRange:   [2]int{15, 150},
		AvgLengthRange:   [2]int{2, 6},
		MaxBlockingDepth: 4,
		ColorCountRange:  [2]int{1, 6},
		MinGridOccupancy: 0.93,
		DefaultGrace:     3,
	},
	"Transcendent": {
		VineCountRange:   [2]int{15, 200},
		AvgLengthRange:   [2]int{2, 6},
		MaxBlockingDepth: 4,
		ColorCountRange:  [2]int{1, 6},
		MinGridOccupancy: 0.93,
		DefaultGrace:     4,
	},
}

// VineColors defines the available vine colors.
var VineColors = map[string]string{
	"default":       "#888888", // Neutral gray for defaults
	"moss_green":    "#7CB342",
	"sunset_orange": "#FF9800",
	"golden_yellow": "#FFC107",
	"royal_purple":  "#7C4DFF",
	"sky_blue":      "#29B6F6",
	"coral_red":     "#FF6E40",
	"lime_green":    "#CDDC39",
}

// HeadDirections defines valid head directions and their deltas.
var HeadDirections = map[string][2]int{
	"right": {1, 0},
	"left":  {-1, 0},
	"up":    {0, 1},
	"down":  {0, -1},
}

// GridSizeRanges defines grid size ranges per difficulty tier.
var GridSizeRanges = map[string]struct {
	MinW, MinH, MaxW, MaxH int
}{
	"Tutorial":     {MinW: 5, MinH: 8, MaxW: 9, MaxH: 12},
	"Seedling":     {MinW: 6, MinH: 8, MaxW: 9, MaxH: 12},
	"Sprout":       {MinW: 9, MinH: 12, MaxW: 12, MaxH: 16},
	"Nurturing":    {MinW: 9, MinH: 16, MaxW: 12, MaxH: 20},
	"Flourishing":  {MinW: 12, MinH: 20, MaxW: 16, MaxH: 24},
	"Transcendent": {MinW: 16, MinH: 28, MaxW: 24, MaxH: 40},
}

// VarietyProfile controls shape and distribution characteristics for generated levels.
type VarietyProfile struct {
	LengthMix  map[string]float64 // keys: "short","medium","long" => relative weights
	TurnMix    float64            // 0..1 proportion of turns (bendiness)
	RegionBias string             // "edge","center","balanced"
	DirBalance map[string]float64 // desired head dir distribution (right,left,up,down)
}

// GeneratorConfig holds generation algorithm tuning parameters and safety caps.
type GeneratorConfig struct {
	MaxSeedRetries    int // retries to find a seed that can grow
	LocalRepairRadius int // radius for local repair tiles
	RepairRetries     int // number of local repair attempts per stuck region
}
