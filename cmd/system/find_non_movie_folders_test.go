//nolint:errcheck
package system

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindNonMovieFolders(t *testing.T) {
	dir := t.TempDir()

	t.Run("movie file in root", func(t *testing.T) {
		movieDir := filepath.Join(dir, "MovieRoot")
		os.Mkdir(movieDir, 0o755)
		os.WriteFile(filepath.Join(movieDir, "movie.mkv"), []byte("dummy"), 0o644)
		folders, _ := findNonMovieFolders(false, dir)
		for _, f := range folders {
			if f == movieDir {
				t.Errorf("Should not mark folder with movie file for deletion: %s", f)
			}
		}
	})

	t.Run("non-movie file in root", func(t *testing.T) {
		nonMovieDir := filepath.Join(dir, "NonMovieRoot")
		os.Mkdir(nonMovieDir, 0o755)
		os.WriteFile(filepath.Join(nonMovieDir, "file.txt"), []byte("dummy"), 0o644)
		folders, _ := findNonMovieFolders(false, dir)
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
		emptyDir := filepath.Join(dir, "EmptyRoot")
		os.Mkdir(emptyDir, 0o755)
		folders, _ := findNonMovieFolders(false, dir)
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
		nestedDir := filepath.Join(dir, "NestedMovie")
		os.MkdirAll(filepath.Join(nestedDir, "sub"), 0o755)
		os.WriteFile(filepath.Join(nestedDir, "sub", "movie.mp4"), []byte("dummy"), 0o644)
		folders, _ := findNonMovieFolders(false, dir)
		for _, f := range folders {
			if f == nestedDir {
				t.Errorf("Should not mark parent folder with nested movie file for deletion: %s", f)
			}
		}
	})

	t.Run("nested non-movie file", func(t *testing.T) {
		nestedNonMovie := filepath.Join(dir, "NestedNonMovie")
		os.MkdirAll(filepath.Join(nestedNonMovie, "sub"), 0o755)
		os.WriteFile(filepath.Join(nestedNonMovie, "sub", "file.doc"), []byte("dummy"), 0o644)
		folders, _ := findNonMovieFolders(false, dir)
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
		mixedDir := filepath.Join(dir, "MixedContent")
		os.MkdirAll(filepath.Join(mixedDir, "sub"), 0o755)
		os.WriteFile(filepath.Join(mixedDir, "file.txt"), []byte("dummy"), 0o644)
		os.WriteFile(filepath.Join(mixedDir, "sub", "movie.avi"), []byte("dummy"), 0o644)
		folders, _ := findNonMovieFolders(false, dir)
		for _, f := range folders {
			if f == mixedDir {
				t.Errorf("Should not mark folder with any movie file (even nested) for deletion: %s", f)
			}
		}
	})

	t.Run("deleted folder", func(t *testing.T) {
		deletedDir := filepath.Join(dir, "DeletedFolder")
		os.Mkdir(deletedDir, 0o755)
		os.RemoveAll(deletedDir)
		folders, _ := findNonMovieFolders(false, dir)
		for _, f := range folders {
			if f == deletedDir {
				t.Errorf("Should not include deleted/missing folder: %s", f)
			}
		}
	})
}
