package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindNonMovieFolders(t *testing.T) {
	dir := t.TempDir()

	createDir := func(name string) string {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatalf("failed to create dir %s: %v", p, err)
		}
		return p
	}
	writeFile := func(dir, fname string) {
		f := filepath.Join(dir, fname)
		if err := os.WriteFile(f, []byte("dummy"), 0o644); err != nil {
			t.Fatalf("failed to write file %s: %v", f, err)
		}
	}

	t.Run("movie file in root", func(t *testing.T) {
		movieDir := createDir("MovieRoot")
		writeFile(movieDir, "movie.mkv")
		folders, _ := findNonMovieFolders(false, dir, nil, nil)
		for _, f := range folders {
			if f == movieDir {
				t.Errorf("Should not mark folder with movie file for deletion: %s", f)
			}
		}
	})

	t.Run("non-movie file in root", func(t *testing.T) {
		nonMovieDir := createDir("NonMovieRoot")
		writeFile(nonMovieDir, "file.txt")
		folders, _ := findNonMovieFolders(false, dir, nil, nil)
		found := false
		for _, f := range folders {
			if f == nonMovieDir {
				found = true
			}
		}
		if !found {
			t.Errorf("Should mark folder with only non-movie files for deletion: %s", nonMovieDir)
		}
	})

	t.Run("empty folder", func(t *testing.T) {
		emptyDir := createDir("EmptyRoot")
		folders, _ := findNonMovieFolders(false, dir, nil, nil)
		found := false
		for _, f := range folders {
			if f == emptyDir {
				found = true
			}
		}
		if !found {
			t.Errorf("Should mark empty folder for deletion: %s", emptyDir)
		}
	})

	t.Run("nested movie file", func(t *testing.T) {
		nestedDir := createDir("NestedMovie")
		sub := createDir(filepath.Join("NestedMovie", "sub"))
		writeFile(sub, "movie.mp4")
		folders, _ := findNonMovieFolders(false, dir, nil, nil)
		for _, f := range folders {
			if f == nestedDir {
				t.Errorf("Should not mark parent folder with nested movie file for deletion: %s", f)
			}
		}
	})

	t.Run("nested non-movie file", func(t *testing.T) {
		nestedNonMovie := createDir("NestedNonMovie")
		sub := createDir(filepath.Join("NestedNonMovie", "sub"))
		writeFile(sub, "file.doc")
		folders, _ := findNonMovieFolders(false, dir, nil, nil)
		found := false
		for _, f := range folders {
			if f == nestedNonMovie {
				found = true
			}
		}
		if !found {
			t.Errorf("Should mark parent folder with only nested non-movie file for deletion: %s", nestedNonMovie)
		}
	})

	t.Run("mixed content", func(t *testing.T) {
		mixedDir := createDir("MixedContent")
		sub := createDir(filepath.Join("MixedContent", "sub"))
		writeFile(mixedDir, "file.txt")
		writeFile(sub, "movie.avi")
		folders, _ := findNonMovieFolders(false, dir, nil, nil)
		for _, f := range folders {
			if f == mixedDir {
				t.Errorf("Should not mark folder with any movie file (even nested) for deletion: %s", f)
			}
		}
	})

	t.Run("deleted folder", func(t *testing.T) {
		deletedDir := createDir("DeletedFolder")
		os.RemoveAll(deletedDir)
		folders, _ := findNonMovieFolders(false, dir, nil, nil)
		for _, f := range folders {
			if f == deletedDir {
				t.Errorf("Should not include deleted/missing folder: %s", f)
			}
		}
	})
}
