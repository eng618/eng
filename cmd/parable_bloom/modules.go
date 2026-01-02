package parable_bloom

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// LoadModules loads module definitions from a JSON file.
// Falls back to default ranges if file not found.
func LoadModules(path string) ([]ModuleRange, error) {
	// Try provided path first
	data, err := os.ReadFile(path)
	if err == nil {
		return parseModules(data)
	}

	// Try default locations
	defaultPaths := []string{
		"assets/data/modules.json",
		"./assets/data/modules.json",
		"../assets/data/modules.json",
	}

	for _, p := range defaultPaths {
		data, err := os.ReadFile(p)
		if err == nil {
			return parseModules(data)
		}
	}

	// Return defaults if no file found
	return defaultModuleRanges(), nil
}

func parseModules(data []byte) ([]ModuleRange, error) {
	var moduleData ModuleData
	decoder := json.NewDecoder(nil)
	decoder.DisallowUnknownFields()

	err := json.Unmarshal(data, &moduleData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse modules.json: %w", err)
	}

	var ranges []ModuleRange
	for _, mod := range moduleData.Modules {
		ranges = append(ranges, ModuleRange{
			ID:    mod.ID,
			Name:  mod.Name,
			Start: mod.LevelRange[0],
			End:   mod.LevelRange[1],
		})
	}

	return ranges, nil
}

// defaultModuleRanges returns sensible defaults if modules.json is not found.
func defaultModuleRanges() []ModuleRange {
	return []ModuleRange{
		{ID: 1, Name: "Tutorial", Start: 1, End: 5},
		{ID: 2, Name: "The Mustard Seed", Start: 6, End: 20},
		{ID: 3, Name: "The Sower", Start: 21, End: 35},
		{ID: 4, Name: "Wheat and Weeds", Start: 36, End: 50},
		{ID: 5, Name: "The Lost Sheep", Start: 51, End: 65},
		{ID: 6, Name: "The Prodigal Son", Start: 66, End: 80},
		{ID: 7, Name: "The Hidden Treasure", Start: 81, End: 90},
		{ID: 8, Name: "The Pearl of Great Price", Start: 91, End: 100},
	}
}

// DifficultyForLevel determines the difficulty tier for a level ID given modules.
func DifficultyForLevel(levelID int, modules []ModuleRange) string {
	// Tutorial module is special
	for _, m := range modules {
		if m.Name == "Tutorial" && levelID >= m.Start && levelID <= m.End {
			return "Tutorial"
		}
	}

	// Map other modules to difficulty tiers by module position
	var nonTutorial []ModuleRange
	for _, m := range modules {
		if m.Name != "Tutorial" {
			nonTutorial = append(nonTutorial, m)
		}
	}

	if len(nonTutorial) == 0 {
		return "Seedling" // Fallback
	}

	// Get module position for this level
	var modulePos int
	for i, m := range nonTutorial {
		if levelID >= m.Start && levelID <= m.End {
			modulePos = i
			break
		}
	}

	totalModules := len(nonTutorial)
	firstThirdCount := (totalModules + 2) / 3
	secondThirdCount := (totalModules*2 + 2) / 3

	switch {
	case modulePos < firstThirdCount:
		return "Seedling"
	case modulePos < secondThirdCount:
		return "Nurturing"
	case modulePos < totalModules-1:
		return "Flourishing"
	default:
		return "Transcendent"
	}
}

// GridSizeForLevel determines grid dimensions based on level ID and difficulty.
func GridSizeForLevel(levelID int, difficulty string) [2]int {
	ranges, ok := GridSizeRanges[difficulty]
	if !ok {
		ranges = GridSizeRanges["Seedling"]
	}

	// Use levelID as seed for deterministic variation
	seed := uint32(levelID) * 2654435761 // FNV-1a offset basis

	// Pseudo-random within ranges
	widthVar := int((seed % uint32(ranges.MaxW-ranges.MinW+1)))
	seed = seed*2654435761 ^ uint32(levelID)
	heightVar := int((seed % uint32(ranges.MaxH-ranges.MinH+1)))

	width := ranges.MinW + widthVar
	height := ranges.MinH + heightVar

	return [2]int{width, height}
}

// FindModuleForLevel finds the module containing the given level ID.
func FindModuleForLevel(levelID int, modules []ModuleRange) *ModuleRange {
	for i := range modules {
		if levelID >= modules[i].Start && levelID <= modules[i].End {
			return &modules[i]
		}
	}
	return nil
}

// GetModuleRange returns the difficulty range for a module.
func GetModuleRange(moduleName string, modules []ModuleRange) *ModuleRange {
	for i := range modules {
		if modules[i].Name == moduleName {
			return &modules[i]
		}
	}
	return nil
}

// ResolveModulesPath resolves the modules.json file path.
func ResolveModulesPath(configPath string) string {
	if configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	// Check standard locations
	stdPaths := []string{
		filepath.Join(os.Getenv("PWD"), "assets", "data", "modules.json"),
		filepath.Join(os.Getenv("HOME"), "parable-bloom", "assets", "data", "modules.json"),
		"assets/data/modules.json",
	}

	for _, p := range stdPaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return "assets/data/modules.json"
}
