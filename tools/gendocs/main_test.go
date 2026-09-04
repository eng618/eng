package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestFindStale(t *testing.T) {
	generated := map[string][]byte{"a.md": []byte("a"), "b.md": []byte("b")}
	existing := map[string]bool{"a.md": true, "old.md": true}

	stale := findStale(generated, existing)
	if len(stale) != 1 || stale[0] != "old.md" {
		t.Errorf("expected [old.md], got %v", stale)
	}
	if stale := findStale(generated, map[string]bool{"a.md": true}); len(stale) != 0 {
		t.Errorf("expected no stale pages, got %v", stale)
	}
}

func TestCheckDrift(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "same.md", "same")
	writeFile(t, dir, "changed.md", "old")

	generated := map[string][]byte{"same.md": []byte("same"), "changed.md": []byte("new")}
	if drift := checkDrift(dir, generated, nil); drift != 1 {
		t.Errorf("expected drift 1, got %d", drift)
	}

	writeFile(t, dir, "changed.md", "new")
	if drift := checkDrift(dir, generated, []string{"stale.md"}); drift != 1 {
		t.Errorf("expected drift 1 (stale only), got %d", drift)
	}

	if drift := checkDrift(dir, generated, nil); drift != 0 {
		t.Errorf("expected no drift, got %d", drift)
	}
}

func TestSyncGenerated(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "same.md", "same")
	writeFile(t, dir, "changed.md", "old")

	generated := map[string][]byte{
		"same.md":    []byte("same"),
		"changed.md": []byte("new"),
		"fresh.md":   []byte("fresh"),
	}
	created, updated, unchanged := syncGenerated(dir, generated)
	if created != 1 || updated != 1 || unchanged != 1 {
		t.Errorf("expected (1,1,1), got (%d,%d,%d)", created, updated, unchanged)
	}

	data, err := os.ReadFile(filepath.Join(dir, "changed.md"))
	if err != nil || string(data) != "new" {
		t.Errorf("expected changed.md rewritten, got %q, err %v", data, err)
	}
}
