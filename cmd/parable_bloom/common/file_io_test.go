package common

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func newSampleLevel(id int) *Level {
	return &Level{
		ID:         id,
		Name:       fmt.Sprintf("Level %d", id),
		Difficulty: "Seedling",
		GridSize:   [2]int{8, 8},
		Mask: &Mask{
			Mode:   "show-all",
			Points: []any{},
		},
		Vines: []Vine{
			{
				ID:            "vine_0",
				HeadDirection: "down",
				OrderedPath:   []Point{{X: 7, Y: 1}, {X: 7, Y: 2}},
				VineColor:     "moss_green",
			},
		},
		MaxMoves:   8,
		MinMoves:   4,
		Complexity: "low",
		Grace:      3,
	}
}

func TestAtomicWriteLevelConcurrent(t *testing.T) {
	dir := t.TempDir()
	lvl := newSampleLevel(1)
	var wg sync.WaitGroup
	errors := make(chan error, 20)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			path := filepath.Join(dir, fmt.Sprintf("level_%d.json", i))
			if err := WriteLevel(path, lvl, true); err != nil {
				errors <- err
				return
			}
			data, err := os.ReadFile(path)
			if err != nil {
				errors <- err
				return
			}
			var m map[string]interface{}
			if err := json.Unmarshal(data, &m); err != nil {
				errors <- err
				return
			}
			// Ensure runtime-only fields are not present
			if _, ok := m["occupancy_percent"]; ok {
				errors <- fmt.Errorf("occupancy_percent should not be persisted")
				return
			}
			if _, ok := m["color_distribution"]; ok {
				errors <- fmt.Errorf("color_distribution should not be persisted")
				return
			}
			if _, ok := m["blocking_graph"]; ok {
				errors <- fmt.Errorf("blocking_graph should not be persisted")
				return
			}
		}(i)
	}

	wg.Wait()
	close(errors)
	for err := range errors {
		t.Fatalf("error during concurrent write test: %v", err)
	}
}

func TestWriteLevelSanityCheck(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "level_test.json")
	lvl := newSampleLevel(42)
	if err := WriteLevel(path, lvl, true); err != nil {
		t.Fatalf("WriteLevel failed: %v", err)
	}
	// Verify file exists and is valid JSON
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("written file is invalid JSON: %v", err)
	}
	if m["id"].(float64) != 42 {
		t.Fatalf("unexpected id in persisted file")
	}
}
