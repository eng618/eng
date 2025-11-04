package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestFiles creates a temporary directory with test files of various types and sizes
func setupTestFiles(t *testing.T) (string, map[string]int) {
	tmpDir := t.TempDir()
	files := map[string]int{
		"a.json":           100,
		"b.txt":            200,
		"sub/c.json":       300,
		"sub/d.mp4":        400,
		"test.bak":         500,
		".hidden.txt":      150, // hidden file
		"invalid*name.txt": 250, // file with special characters
		"sub/e.jpg":        350,
		"sub/f.png":        450,
		"sub/g.mov":        550,
		"sub/h.avi":        650,
		"sub/i.mkv":        750,
		"sub/j.jpeg":       850,
		"sub/k.gif":        950,
		"sub/l.ds_store":   50,
	}

	for f, size := range files {
		full := filepath.Join(tmpDir, f)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("mkdirall: %v", err)
		}
		content := make([]byte, size)
		if err := os.WriteFile(full, content, 0o644); err != nil {
			t.Fatalf("write file: %v", err)
		}
	}

	return tmpDir, files
}

// TestBuildMatchFunction tests the pattern matching function builder
func TestBuildMatchFunction(t *testing.T) {
	tests := []struct {
		name        string
		glob        string
		ext         string
		testFile    string
		shouldMatch bool
		wantErr     bool
	}{
		{
			name:        "valid glob",
			glob:        "*.json",
			testFile:    "test.json",
			shouldMatch: true,
		},
		{
			name:    "invalid glob",
			glob:    "[",
			wantErr: true,
		},
		{
			name:        "valid extension",
			ext:         ".txt",
			testFile:    "test.txt",
			shouldMatch: true,
		},
		{
			name:        "extension without dot",
			ext:         "json",
			testFile:    "test.json",
			shouldMatch: true,
		},
		{
			name:        "no pattern",
			testFile:    "test.txt",
			shouldMatch: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			matchFn, err := buildMatchFunction(tc.glob, tc.ext)
			if tc.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if matchFn == nil {
				if tc.shouldMatch {
					t.Error("expected match function, got nil")
				}
				return
			}
			if got := matchFn(tc.testFile); got != tc.shouldMatch {
				t.Errorf("match result = %v, want %v", got, tc.shouldMatch)
			}
		})
	}
}

// TestDeleteFiles tests the parallel file deletion
func TestDeleteFiles(t *testing.T) {
	tmpDir, files := setupTestFiles(t)

	// Create a list of files to delete (only .txt files)
	var toDelete []string
	var expectedDeleted int
	for f := range files {
		if strings.HasSuffix(f, ".txt") {
			toDelete = append(toDelete, filepath.Join(tmpDir, f))
			expectedDeleted++
		}
	}

	deleted, errors := deleteFiles(toDelete, false)

	if deleted != int64(expectedDeleted) {
		t.Errorf("expected %d deletions, got %d", expectedDeleted, deleted)
	}
	if errors != 0 {
		t.Errorf("expected 0 errors, got %d", errors)
	}

	// Verify files were actually deleted
	for _, f := range toDelete {
		if _, err := os.Stat(f); !os.IsNotExist(err) {
			t.Errorf("file %s should be deleted but still exists", f)
		}
	}
}

// TestScanFiles creates a temp directory with a mix of files and tests different matching strategies
func TestScanFiles(t *testing.T) {
	tmpDir, _ := setupTestFiles(t)

	tests := []struct {
		name          string
		matchFn       func(name string) bool
		expectedFiles []string
		expectedSize  int64
	}{
		{
			name: "extension match",
			matchFn: func(name string) bool {
				return strings.ToLower(filepath.Ext(name)) == ".json"
			},
			expectedFiles: []string{"a.json", "sub/c.json"},
			expectedSize:  400, // 100 + 300
		},
		{
			name: "glob match",
			matchFn: func(name string) bool {
				matched, _ := filepath.Match("*.bak", name)
				return matched
			},
			expectedFiles: []string{"test.bak"},
			expectedSize:  500,
		},
		{
			name: "txt files match",
			matchFn: func(name string) bool {
				return strings.ToLower(filepath.Ext(name)) == ".txt"
			},
			expectedFiles: []string{".hidden.txt", "b.txt", "invalid*name.txt"},
			expectedSize:  600, // 150 + 200 + 250
		},
		{
			name: "no match",
			matchFn: func(name string) bool {
				return strings.HasSuffix(name, ".nonexistent")
			},
			expectedFiles: []string{},
			expectedSize:  0,
		},
		{
			name: "video files match",
			matchFn: func(name string) bool {
				ext := strings.ToLower(filepath.Ext(name))
				return ext == ".mp4" || ext == ".mov" || ext == ".avi" || ext == ".mkv"
			},
			expectedFiles: []string{"sub/d.mp4", "sub/g.mov", "sub/h.avi", "sub/i.mkv"},
			expectedSize:  2350, // 400 + 550 + 650 + 750
		},
		{
			name: "image files match",
			matchFn: func(name string) bool {
				ext := strings.ToLower(filepath.Ext(name))
				return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif"
			},
			expectedFiles: []string{"sub/e.jpg", "sub/f.png", "sub/j.jpeg", "sub/k.gif"},
			expectedSize:  2600, // 350 + 450 + 850 + 950
		},
		{
			name: "system files match",
			matchFn: func(name string) bool {
				return strings.ToLower(filepath.Ext(name)) == ".ds_store"
			},
			expectedFiles: []string{"sub/l.ds_store"},
			expectedSize:  50,
		},
		{
			name: "combined json and video files match",
			matchFn: func(name string) bool {
				ext := strings.ToLower(filepath.Ext(name))
				return ext == ".json" || ext == ".mp4" || ext == ".mov" || ext == ".avi" || ext == ".mkv"
			},
			expectedFiles: []string{"a.json", "sub/c.json", "sub/d.mp4", "sub/g.mov", "sub/h.avi", "sub/i.mkv"},
			expectedSize:  2750, // 100 + 300 + 400 + 550 + 650 + 750
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Don't use spinner in tests
			matches, totalSize, err := ScanFiles(tmpDir, tc.matchFn, nil)
			if err != nil {
				t.Fatalf("ScanFiles error: %v", err)
			}

			if totalSize != tc.expectedSize {
				t.Errorf("expected total size %d, got %d", tc.expectedSize, totalSize)
			}

			if len(matches) != len(tc.expectedFiles) {
				t.Fatalf("expected %d files, got %d (%v)", len(tc.expectedFiles), len(matches), matches)
			}

			// ensure returned files are the expected ones
			foundPaths := make(map[string]bool)
			for _, m := range matches {
				rel, err := filepath.Rel(tmpDir, m)
				if err != nil {
					t.Fatalf("failed to get relative path: %v", err)
				}
				foundPaths[rel] = true
			}

			for _, expected := range tc.expectedFiles {
				if !foundPaths[expected] {
					t.Errorf("missing expected file %s", expected)
				}
			}

			// Check we haven't found any unexpected files
			for found := range foundPaths {
				expectedFile := false
				for _, e := range tc.expectedFiles {
					if e == found {
						expectedFile = true
						break
					}
				}
				if !expectedFile {
					t.Errorf("found unexpected file %s", found)
				}
			}
		})
	}
}
