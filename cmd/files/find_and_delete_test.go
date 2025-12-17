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
		// Original test files
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

		// New category test files
		// Microsoft documents
		"document.doc":      1200,
		"document.docx":     1300,
		"spreadsheet.xls":   1400,
		"spreadsheet.xlsx":  1500,
		"presentation.ppt":  1600,
		"presentation.pptx": 1700,

		// Archive files
		"archive.zip": 1800,
		"archive.rar": 1900,
		"archive.7z":  2000,
		"archive.tar": 2100,
		"archive.gz":  2200,
		"archive.bz2": 2300,

		// Audio files
		"music.mp3":  2400,
		"sound.wav":  2500,
		"audio.flac": 2600,
		"song.aac":   2700,
		"track.ogg":  2800,

		// PDF documents
		"document.pdf": 2900,

		// Text files
		"readme.md": 300,
		"notes.rtf": 3100,

		// Log files
		"app.log": 3200,

		// Temporary files
		"temp.tmp":   3300,
		"cache.temp": 3400,
		"swap.swp":   3500,

		// Backup files
		"config.backup": 3600,
		"data.old":      3700,

		// Executable files
		"program.exe":   3800,
		"installer.msi": 3900,
		"app.dmg":       4000,
		"package.pkg":   4100,
		"software.deb":  4200,
		"rpm.rpm":       4300,

		// Additional video files for .m4v, .wmv, .3gp
		"video.m4v": 4400,
		"movie.wmv": 4500,
		"clip.3gp":  4600,

		// Edge case test files
		"my.file.txt":      100, // file with multiple dots
		"noextension":      200, // file without extension
		"001/DOCUMENT.PDF": 300, // uppercase extension in subdirectory to avoid case collision
		"002/config.BAK":   400, // uppercase backup extension in subdirectory to avoid case collision
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
		filename    string
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
			name:        "specific filename match",
			filename:    "package.json",
			testFile:    "package.json",
			shouldMatch: true,
		},
		{
			name:        "specific filename no match",
			filename:    "package.json",
			testFile:    "other.json",
			shouldMatch: false,
		},
		{
			name:        "no pattern",
			testFile:    "test.txt",
			shouldMatch: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			matchFn, err := buildMatchFunction(tc.glob, tc.ext, tc.filename)
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

// TestListExtensions tests the ListExtensions function
func TestListExtensions(t *testing.T) {
	tmpDir, _ := setupTestFiles(t)

	extensions, err := ListExtensions(tmpDir)
	if err != nil {
		t.Fatalf("ListExtensions error: %v", err)
	}

	expected := []string{
		".3gp", ".7z", ".aac", ".avi", ".backup", ".bak", ".bz2", ".deb", ".dmg", ".doc", ".docx",
		".ds_store", ".exe", ".flac", ".gif", ".gz", ".jpeg", ".jpg", ".json", ".log", ".m4v",
		".md", ".mkv", ".mov", ".mp3", ".mp4", ".msi", ".ogg", ".old", ".pdf", ".pkg", ".png",
		".ppt", ".pptx", ".rar", ".rpm", ".rtf", ".swp", ".tar", ".temp", ".tmp", ".txt",
		".wav", ".wmv", ".xls", ".xlsx", ".zip",
	}
	if len(extensions) != len(expected) {
		t.Fatalf("expected %d extensions, got %d (%v)", len(expected), len(extensions), extensions)
	}

	for i, exp := range expected {
		if extensions[i] != exp {
			t.Errorf("expected extension %s at index %d, got %s", exp, i, extensions[i])
		}
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
			expectedFiles: []string{".hidden.txt", "b.txt", "invalid*name.txt", "my.file.txt"},
			expectedSize:  700, // 150 + 200 + 250 + 100
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
		{
			name: "specific filename match",
			matchFn: func(name string) bool {
				return name == "a.json"
			},
			expectedFiles: []string{"a.json"},
			expectedSize:  100,
		},
		{
			name: "microsoft documents match",
			matchFn: func(name string) bool {
				ext := strings.ToLower(filepath.Ext(name))
				return ext == ".doc" || ext == ".docx" || ext == ".xls" || ext == ".xlsx" || ext == ".ppt" || ext == ".pptx"
			},
			expectedFiles: []string{"document.doc", "document.docx", "spreadsheet.xls", "spreadsheet.xlsx", "presentation.ppt", "presentation.pptx"},
			expectedSize:  8700, // 1200 + 1300 + 1400 + 1500 + 1600 + 1700
		},
		{
			name: "archive files match",
			matchFn: func(name string) bool {
				ext := strings.ToLower(filepath.Ext(name))
				return ext == ".zip" || ext == ".rar" || ext == ".7z" || ext == ".tar" || ext == ".gz" || ext == ".bz2"
			},
			expectedFiles: []string{"archive.zip", "archive.rar", "archive.7z", "archive.tar", "archive.gz", "archive.bz2"},
			expectedSize:  12300, // 1800 + 1900 + 2000 + 2100 + 2200 + 2300
		},
		{
			name: "audio files match",
			matchFn: func(name string) bool {
				ext := strings.ToLower(filepath.Ext(name))
				return ext == ".mp3" || ext == ".wav" || ext == ".flac" || ext == ".aac" || ext == ".ogg"
			},
			expectedFiles: []string{"music.mp3", "sound.wav", "audio.flac", "song.aac", "track.ogg"},
			expectedSize:  13000, // 2400 + 2500 + 2600 + 2700 + 2800
		},
		{
			name: "pdf documents match",
			matchFn: func(name string) bool {
				return strings.ToLower(filepath.Ext(name)) == ".pdf"
			},
			expectedFiles: []string{"document.pdf", "001/DOCUMENT.PDF"},
			expectedSize:  3200, // 2900 + 300
		},
		{
			name: "text files match",
			matchFn: func(name string) bool {
				ext := strings.ToLower(filepath.Ext(name))
				return ext == ".txt" || ext == ".md" || ext == ".rtf"
			},
			expectedFiles: []string{".hidden.txt", "b.txt", "invalid*name.txt", "my.file.txt", "readme.md", "notes.rtf"},
			expectedSize:  4100, // 150 + 200 + 250 + 100 + 300 + 3100
		},
		{
			name: "log files match",
			matchFn: func(name string) bool {
				return strings.ToLower(filepath.Ext(name)) == ".log"
			},
			expectedFiles: []string{"app.log"},
			expectedSize:  3200,
		},
		{
			name: "temporary files match",
			matchFn: func(name string) bool {
				ext := strings.ToLower(filepath.Ext(name))
				return ext == ".tmp" || ext == ".temp" || ext == ".swp"
			},
			expectedFiles: []string{"temp.tmp", "cache.temp", "swap.swp"},
			expectedSize:  10200, // 3300 + 3400 + 3500
		},
		{
			name: "backup files match",
			matchFn: func(name string) bool {
				ext := strings.ToLower(filepath.Ext(name))
				return ext == ".bak" || ext == ".backup" || ext == ".old"
			},
			expectedFiles: []string{"test.bak", "config.backup", "data.old", "002/config.BAK"},
			expectedSize:  8200, // 500 + 3600 + 3700 + 400
		},
		{
			name: "executable files match",
			matchFn: func(name string) bool {
				ext := strings.ToLower(filepath.Ext(name))
				return ext == ".exe" || ext == ".msi" || ext == ".dmg" || ext == ".pkg" || ext == ".deb" || ext == ".rpm"
			},
			expectedFiles: []string{"program.exe", "installer.msi", "app.dmg", "package.pkg", "software.deb", "rpm.rpm"},
			expectedSize:  24300, // 3800 + 3900 + 4000 + 4100 + 4200 + 4300
		},
		{
			name: "m4v video files match",
			matchFn: func(name string) bool {
				return strings.ToLower(filepath.Ext(name)) == ".m4v"
			},
			expectedFiles: []string{"video.m4v"},
			expectedSize:  4400,
		},
		{
			name: "all video files match including m4v, wmv, 3gp",
			matchFn: func(name string) bool {
				ext := strings.ToLower(filepath.Ext(name))
				return ext == ".mp4" || ext == ".mov" || ext == ".avi" || ext == ".mkv" || ext == ".m4v" || ext == ".wmv" || ext == ".3gp"
			},
			expectedFiles: []string{"sub/d.mp4", "sub/g.mov", "sub/h.avi", "sub/i.mkv", "video.m4v", "movie.wmv", "clip.3gp"},
			expectedSize:  15850, // 400 + 550 + 650 + 750 + 4400 + 4500 + 4600
		},
		// Edge case tests for safety
		{
			name: "edge case - files with multiple dots",
			matchFn: func(name string) bool {
				return strings.ToLower(filepath.Ext(name)) == ".txt"
			},
			expectedFiles: []string{".hidden.txt", "b.txt", "invalid*name.txt", "my.file.txt"},
			expectedSize:  700, // 150 + 200 + 250 + 100 - should match "my.file.txt" correctly
		},
		{
			name: "edge case - files without extensions don't match",
			matchFn: func(name string) bool {
				return strings.ToLower(filepath.Ext(name)) == ".txt"
			},
			expectedFiles: []string{".hidden.txt", "b.txt", "invalid*name.txt", "my.file.txt"},
			expectedSize:  700, // Files without extensions should not match
		},
		{
			name: "edge case - similar extensions don't cross-match",
			matchFn: func(name string) bool {
				return strings.ToLower(filepath.Ext(name)) == ".mp3"
			},
			expectedFiles: []string{"music.mp3"},
			expectedSize:  2400, // Should not match .mp4 files
		},
		{
			name: "edge case - doc vs docx distinction",
			matchFn: func(name string) bool {
				return strings.ToLower(filepath.Ext(name)) == ".doc"
			},
			expectedFiles: []string{"document.doc"},
			expectedSize:  1200, // Should not match .docx files
		},
		{
			name: "edge case - case insensitive matching",
			matchFn: func(name string) bool {
				return strings.ToLower(filepath.Ext(name)) == ".pdf"
			},
			expectedFiles: []string{"document.pdf", "001/DOCUMENT.PDF"},
			expectedSize:  3200, // 2900 + 300 - Case should not matter
		},
		{
			name: "edge case - backup files don't match other categories",
			matchFn: func(name string) bool {
				return strings.ToLower(filepath.Ext(name)) == ".bak"
			},
			expectedFiles: []string{"test.bak", "002/config.BAK"},
			expectedSize:  900, // 500 + 400 - Should not match .backup files
		},
		{
			name: "edge case - temporary files with different extensions",
			matchFn: func(name string) bool {
				ext := strings.ToLower(filepath.Ext(name))
				return ext == ".tmp" || ext == ".temp" || ext == ".swp"
			},
			expectedFiles: []string{"temp.tmp", "cache.temp", "swap.swp"},
			expectedSize:  10200, // Should match exactly these extensions
		},
		{
			name: "edge case - files without extensions are not matched",
			matchFn: func(name string) bool {
				return strings.ToLower(filepath.Ext(name)) == ""
			},
			expectedFiles: []string{"noextension"},
			expectedSize:  200, // Only files without extensions should match
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
