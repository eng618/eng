package common

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadModules_Fallback(t *testing.T) {
	// Call LoadModules with a path that does not exist.
	// It should fallback to defaultModuleRanges()

	// Create a temporary directory and change working directory to it
	// to ensure that default locations (assets/data/modules.json, etc.)
	// are not inadvertently hit if they exist in the real tree.
	tempDir := t.TempDir()
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer func() {
		os.Chdir(originalWD)
	}()

	ranges, err := LoadModules("non_existent_path.json")
	if err != nil {
		t.Fatalf("LoadModules returned unexpected error: %v", err)
	}

	expected := defaultModuleRanges()

	if !reflect.DeepEqual(ranges, expected) {
		t.Errorf("LoadModules returned %v, expected %v", ranges, expected)
	}
}

func TestLoadModules_Success(t *testing.T) {
	// Test the happy path by writing a valid JSON to a temp file and loading it.
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "modules.json")

	validJSON := `{
		"version": "1.0",
		"modules": [
			{
				"id": 1,
				"name": "Test Module 1",
				"level_range": [1, 10]
			},
			{
				"id": 2,
				"name": "Test Module 2",
				"level_range": [11, 20]
			}
		]
	}`

	err := os.WriteFile(tempFile, []byte(validJSON), 0644)
	if err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	ranges, err := LoadModules(tempFile)
	if err != nil {
		t.Fatalf("LoadModules returned unexpected error: %v", err)
	}

	expected := []ModuleRange{
		{ID: 1, Name: "Test Module 1", Start: 1, End: 10},
		{ID: 2, Name: "Test Module 2", Start: 11, End: 20},
	}

	if !reflect.DeepEqual(ranges, expected) {
		t.Errorf("LoadModules returned %v, expected %v", ranges, expected)
	}
}
