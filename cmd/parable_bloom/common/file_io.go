package common

// TODO: DEPRECATED - This file has been migrated to parable-bloom/tools/level-builder/pkg/common/file_io.go
// All file I/O functions (ReadLevel, WriteLevel, atomic writes) have been moved to the new location.
// This file can be safely deleted once the old parable-bloom command is fully deprecated.
// Migration completed: 2026-01-13

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReadLevel reads a single level from a JSON file.
func ReadLevel(filePath string) (*Level, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read level file %s: %w", filePath, err)
	}

	var level Level
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	err = decoder.Decode(&level)
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

	// Prepare a sanitized level for persistence (exclude runtime-only fields)
	type persistLevel struct {
		ID                  int     `json:"id"`
		Name                string  `json:"name"`
		Difficulty          string  `json:"difficulty"`
		GridSize            [2]int  `json:"grid_size"`
		Mask                *Mask   `json:"mask"`
		Vines               []Vine  `json:"vines"`
		MaxMoves            int     `json:"max_moves"`
		MinMoves            int     `json:"min_moves"`
		Complexity          string  `json:"complexity"`
		Grace               int     `json:"grace"`
		GenerationSeed      int64   `json:"generation_seed,omitempty"`
		GenerationAttempts  int     `json:"generation_attempts,omitempty"`
		GenerationElapsedMS int64   `json:"generation_elapsed_ms,omitempty"`
		GenerationScore     float64 `json:"generation_score,omitempty"`
	}

	pLevel := persistLevel{
		ID:                  level.ID,
		Name:                level.Name,
		Difficulty:          level.Difficulty,
		GridSize:            level.GridSize,
		Mask:                level.Mask,
		Vines:               level.Vines,
		MaxMoves:            level.MaxMoves,
		MinMoves:            level.MinMoves,
		Complexity:          level.Complexity,
		Grace:               level.Grace,
		GenerationSeed:      level.GenerationSeed,
		GenerationAttempts:  level.GenerationAttempts,
		GenerationElapsedMS: level.GenerationElapsedMS,
		GenerationScore:     level.GenerationScore,
	}

	// Marshal sanitized level
	data, err := json.MarshalIndent(pLevel, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal level to JSON: %w", err)
	}

	// Sanity check: ensure marshaled bytes decode correctly
	var tmp interface{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return fmt.Errorf("sanity check failed: marshaled JSON invalid: %w", err)
	}

	// Write atomically
	if err := atomicWriteFile(filePath, data, 0o644); err != nil {
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
	base := filepath.Base(name)
	if filepath.Ext(name) != ".json" {
		return false
	}
	return strings.HasPrefix(base, "level") || strings.HasPrefix(base, "test_")
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

// atomicWriteFile writes data to a temporary file and renames it into place to ensure atomicity.
func atomicWriteFile(filePath string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(filePath)
	tmpFile, err := os.CreateTemp(dir, "tmplevel-*.json")
	if err != nil {
		return err
	}
	tmpName := tmpFile.Name()
	// Ensure tmp file cleanup on any failure
	defer func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpName)
	}()

	if _, err := tmpFile.Write(data); err != nil {
		return err
	}
	if err := tmpFile.Sync(); err != nil {
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	if err := os.Chmod(tmpName, perm); err != nil {
		return err
	}

	// Rename is atomic on POSIX filesystems
	if err := os.Rename(tmpName, filePath); err != nil {
		return err
	}

	return nil
}
