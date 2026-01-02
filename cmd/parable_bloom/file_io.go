package parable_bloom

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ReadLevel reads a single level from a JSON file.
func ReadLevel(filePath string) (*Level, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read level file %s: %w", filePath, err)
	}

	var level Level
	decoder := json.NewDecoder(nil)
	decoder.DisallowUnknownFields()

	err = json.Unmarshal(data, &level)
	if err != nil {
		return nil, fmt.Errorf("failed to parse level file %s: %w", filePath, err)
	}

	return &level, nil
}

// WriteLevel writes a level to a JSON file.
// Returns error if file exists and overwrite is false.
func WriteLevel(filePath string, level *Level, overwrite bool) error {
	// Check if file exists
	_, err := os.Stat(filePath)
	fileExists := err == nil

	if fileExists && !overwrite {
		return fmt.Errorf("file already exists: %s (use --overwrite to replace)", filePath)
	}

	// Create directory if needed
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Marshal level to JSON with indentation
	data, err := json.MarshalIndent(level, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal level to JSON: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write level file %s: %w", filePath, err)
	}

	return nil
}

// ReadLevelsFromDir reads all level files from a directory.
func ReadLevelsFromDir(dirPath string) ([]*Level, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	var levels []*Level
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !isLevelFile(entry.Name()) {
			continue
		}

		filePath := filepath.Join(dirPath, entry.Name())
		level, err := ReadLevel(filePath)
		if err != nil {
			// Log error but continue processing other files
			fmt.Fprintf(os.Stderr, "warning: %v\n", err)
			continue
		}

		levels = append(levels, level)
	}

	return levels, nil
}

// isLevelFile checks if a filename is a level file.
func isLevelFile(name string) bool {
	return filepath.Ext(name) == ".json" &&
		(filepath.Base(name)[:5] == "level" || filepath.Base(name)[:5] == "test_")
}

// FileExists checks if a file exists.
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

// EnsureDir creates a directory if it doesn't exist.
func EnsureDir(dirPath string) error {
	return os.MkdirAll(dirPath, 0o755)
}

// GetLevelFilePath returns the expected file path for a level.
func GetLevelFilePath(levelID int, baseDir string) string {
	return filepath.Join(baseDir, fmt.Sprintf("level_%d.json", levelID))
}

// marshalJSON marshals a value to indented JSON.
func marshalJSON(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}
