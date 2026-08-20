package system

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/eng618/eng/internal/ui"
)

func createTestTarGz(t *testing.T, archivePath string, files map[string]string) {
	t.Helper()
	outFile, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("failed to create tarball: %v", err)
	}
	defer outFile.Close()

	gzWriter := gzip.NewWriter(outFile)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	for name, content := range files {
		hdr := &tar.Header{
			Name:     name,
			Mode:     0o755,
			Size:     int64(len(content)),
			Typeflag: tar.TypeReg,
		}
		if err := tarWriter.WriteHeader(hdr); err != nil {
			t.Fatalf("failed to write header: %v", err)
		}
		if _, err := tarWriter.Write([]byte(content)); err != nil {
			t.Fatalf("failed to write content: %v", err)
		}
	}
}

func TestFindLatestIdeArchive(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Empty dir returns empty
	if got := findLatestIdeArchive(tmpDir); got != "" {
		t.Errorf("expected empty string for empty dir, got %q", got)
	}

	// 2. Create older IDE archive and newer generic archive
	olderIde := filepath.Join(tmpDir, "antigravity-ide-v1.tar.gz")
	newerGeneric := filepath.Join(tmpDir, "antigravity-v2.tar.gz")

	_ = os.WriteFile(olderIde, []byte("test"), 0o644)
	time.Sleep(10 * time.Millisecond)
	_ = os.WriteFile(newerGeneric, []byte("test"), 0o644)

	// Since olderIde has "ide" in name, it should be prioritized over generic antigravity
	got := findLatestIdeArchive(tmpDir)
	if got != olderIde {
		t.Errorf("expected %s, got %s", olderIde, got)
	}

	// 3. Create newer IDE archive
	newerIde := filepath.Join(tmpDir, "Antigravity IDE-v2.tar.gz")
	time.Sleep(10 * time.Millisecond)
	_ = os.WriteFile(newerIde, []byte("test"), 0o644)

	got = findLatestIdeArchive(tmpDir)
	if got != newerIde {
		t.Errorf("expected %s, got %s", newerIde, got)
	}
}

func TestInstallIdeArchive_Validation(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := t.TempDir()

	origExec := execCommand
	ui.DisableProgress = true
	defer func() {
		execCommand = origExec
		ui.DisableProgress = false
	}()

	// 1. Desktop App archive (has resources/app.asar, no cli.js) should fail validation
	desktopAppTar := filepath.Join(tmpDir, "Antigravity.tar.gz")
	createTestTarGz(t, desktopAppTar, map[string]string{
		"Antigravity/antigravity":        "binary",
		"Antigravity/resources/app.asar": "asar payload",
	})

	err := installIdeArchive(desktopAppTar, homeDir, false)
	if err == nil {
		t.Fatal("expected installIdeArchive to fail for Desktop Hub archive, but it succeeded")
	}
	if !strings.Contains(err.Error(), "appears to be the Antigravity Desktop App") {
		t.Errorf("unexpected error message: %v", err)
	}

	// 2. Valid IDE archive (has resources/app/out/cli.js and bin/antigravity-ide)
	ideTar := filepath.Join(tmpDir, "Antigravity IDE.tar.gz")
	createTestTarGz(t, ideTar, map[string]string{
		"Antigravity IDE/antigravity-ide":          "#!/bin/sh\necho 1.107.0",
		"Antigravity IDE/bin/antigravity-ide":      "#!/bin/sh\necho 1.107.0",
		"Antigravity IDE/resources/app/out/cli.js": "console.log('cli');",
	})

	execCommand = func(name string, args ...string) *exec.Cmd {
		return origExec("echo", "1.107.0")
	}

	err = installIdeArchive(ideTar, homeDir, false)
	if err != nil {
		t.Fatalf("expected valid IDE archive to install successfully, got: %v", err)
	}

	// Verify installed files and symlinks
	installedBin := filepath.Join(homeDir, ".local", "opt", "antigravity-ide", "bin", "antigravity-ide")
	if _, err := os.Stat(installedBin); err != nil {
		t.Errorf("expected installed bin at %s, got err: %v", installedBin, err)
	}

	symlinkPath := filepath.Join(homeDir, ".local", "bin", "agy-ide")
	if target, err := os.Readlink(symlinkPath); err != nil || target != installedBin {
		t.Errorf("expected symlink %s -> %s, got target=%s, err=%v", symlinkPath, installedBin, target, err)
	}
}

func TestDownloadArchive(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "4")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "test")
	}))
	defer ts.Close()

	homeDir := t.TempDir()
	ui.DisableProgress = true
	defer func() { ui.DisableProgress = false }()

	downloaded, err := downloadArchive(context.Background(), ts.URL, homeDir, false)
	if err != nil {
		t.Fatalf("downloadArchive failed: %v", err)
	}
	defer os.Remove(downloaded)

	content, err := os.ReadFile(downloaded)
	if err != nil || string(content) != "test" {
		t.Errorf("unexpected content: %s, err: %v", string(content), err)
	}
}
